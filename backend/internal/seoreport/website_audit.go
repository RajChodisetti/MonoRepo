package seoreport

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

// WebsiteAudit is the visual review of a restaurant website homepage.
type WebsiteAudit struct {
	QualityScore     int    // 0–100, intentionally strict (most sites land 20–60)
	Review           string // short spoken/report summary
	Screenshot       string // data:image/jpeg;base64,... (desktop)
	MobileScreenshot string // data:image/jpeg;base64,... (phone viewport for scan mockup)
	Source           string // "vision" | "fallback" | "social" | "none"
}

func isSocialWebsite(website string) bool {
	parsed, err := url.Parse(strings.TrimSpace(website))
	if err != nil || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	_, ok := socialHosts[host]
	return ok
}

func hasHTTPSWebsite(website string) bool {
	parsed, err := url.Parse(strings.TrimSpace(website))
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Scheme, "https")
}

// normalizeWebsiteURL upgrades http→https so headless capture matches real mobile browsing.
func normalizeWebsiteURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		// Best-effort: rewrite scheme prefix even if Parse fails oddly.
		if strings.HasPrefix(strings.ToLower(raw), "http://") {
			return "https://" + raw[len("http://"):]
		}
		return raw
	}
	if strings.EqualFold(parsed.Scheme, "http") || parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String()
}

// AuditWebsite screenshots the homepage (Playwright if available, else ChromeDP)
// and scores design/UX via vision LLM. Strict by design.
func AuditWebsite(ctx context.Context, website string, llm llmlib.Client) WebsiteAudit {
	website = strings.TrimSpace(website)
	if website == "" {
		return WebsiteAudit{Source: "none"}
	}
	if !strings.Contains(website, "://") {
		website = "https://" + website
	}
	website = normalizeWebsiteURL(website)

	if isSocialWebsite(website) {
		return WebsiteAudit{
			QualityScore: 28,
			Review:       "This Google listing links to a social profile, not a dedicated restaurant website. Expect weaker branding, menus, and direct ordering.",
			Source:       "social",
		}
	}

	shotCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	// Mobile-first: iPhone UA often clears CF faster, and the scan phone mockup needs it.
	type shotResult struct {
		jpeg []byte
		err  error
	}
	mobileCh := make(chan shotResult, 1)
	deskCh := make(chan shotResult, 1)
	go func() {
		b, err := captureWebsiteJPEG(shotCtx, website, true)
		mobileCh <- shotResult{jpeg: b, err: err}
	}()
	go func() {
		b, err := captureWebsiteJPEG(shotCtx, website, false)
		deskCh <- shotResult{jpeg: b, err: err}
	}()

	mobileRes := <-mobileCh
	deskRes := <-deskCh

	jpegBytes := deskRes.jpeg
	if len(jpegBytes) == 0 {
		jpegBytes = mobileRes.jpeg
	}
	if len(jpegBytes) == 0 {
		return fallbackWebsiteAudit(website, fmt.Sprintf("screenshot failed: desk=%v mobile=%v", deskRes.err, mobileRes.err))
	}
	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpegBytes)
	mobileDataURL := ""
	if len(mobileRes.jpeg) > 0 {
		mobileDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(mobileRes.jpeg)
	} else {
		mobileDataURL = dataURL
	}

	attachShots := func(audit WebsiteAudit) WebsiteAudit {
		audit.Screenshot = dataURL
		audit.MobileScreenshot = mobileDataURL
		return audit
	}

	if llm == nil || !llm.Enabled() {
		return attachShots(fallbackWebsiteAudit(website, "llm unavailable"))
	}

	visionCtx, visionCancel := context.WithTimeout(ctx, 40*time.Second)
	defer visionCancel()

	prompt := `You are auditing a restaurant website homepage screenshot for local SEO and guest conversion.
Be VERY STRICT. Most restaurant sites should score between 20 and 60 out of 100.
Scores above 65 are rare. Only an exceptional modern site with clear menu, order path, mobile polish, and strong branding deserves 70+.
Never give 90+ unless the site looks world-class.

Judge: visual design quality, trust, clarity of cuisine/location, menu visibility, order/reserve CTA, mobile-looking layout, clutter, outdated templates, broken/empty hero, stock-photo feel.

Return ONLY compact JSON (no markdown):
{"score": <int 0-100>, "summary": "<2 short sentences>", "strengths": ["..."], "weaknesses": ["..."]}`

	raw, err := llm.CompleteVision(visionCtx, prompt, jpegBytes, "image/jpeg")
	if err != nil || strings.TrimSpace(raw) == "" {
		return attachShots(fallbackWebsiteAudit(website, fmt.Sprintf("vision failed: %v", err)))
	}

	score, summary := parseWebsiteVisionJSON(raw)
	score = clampStrictWebsiteQuality(score)
	if summary == "" {
		summary = "Homepage visual quality is average for a local restaurant site; tighten design, menu access, and booking CTAs."
	}
	return attachShots(WebsiteAudit{
		QualityScore: score,
		Review:       summary,
		Source:       "vision",
	})
}

func fallbackWebsiteAudit(website, reason string) WebsiteAudit {
	_ = reason
	score := 32
	if hasHTTPSWebsite(website) {
		score = 38
	}
	return WebsiteAudit{
		QualityScore: score,
		Review:       "We could not fully review the live homepage visuals, so this is a conservative estimate. A dedicated site helps, but design, menu clarity, and order CTAs still need a human-quality pass.",
		Source:       "fallback",
	}
}

func clampStrictWebsiteQuality(score int) int {
	score = clamp(score, 0, 100)
	// Aggressive ceiling: visual quality must stay in the 20–60 band almost always.
	if score > 60 {
		score = 48 + (score-60)/4 // 80 → 53, 100 → 58
	}
	if score > 60 {
		score = 60
	}
	if score > 0 && score < 20 {
		score = 20
	}
	return score
}

func parseWebsiteVisionJSON(raw string) (int, string) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var parsed struct {
		Score      int      `json:"score"`
		Summary    string   `json:"summary"`
		Strengths  []string `json:"strengths"`
		Weaknesses []string `json:"weaknesses"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return 40, strings.TrimSpace(raw)
	}
	summary := strings.TrimSpace(parsed.Summary)
	if summary == "" && len(parsed.Weaknesses) > 0 {
		summary = strings.Join(parsed.Weaknesses, " ")
	}
	return parsed.Score, summary
}

func captureWebsiteJPEG(ctx context.Context, pageURL string, mobile bool) ([]byte, error) {
	// Prefer Playwright script when Node + playwright are available.
	if png, err := captureWithPlaywright(ctx, pageURL, mobile); err == nil && len(png) > 0 {
		return pngToJPEG(png, 82)
	}
	png, err := captureWithChromedp(ctx, pageURL, mobile)
	if err != nil {
		return nil, err
	}
	return pngToJPEG(png, 82)
}

func isBotBlockPage(title, body string) bool {
	blob := strings.ToLower(strings.TrimSpace(title + "\n" + body))
	if blob == "" {
		return false
	}
	// Hard blocks only — transient CF "checking your browser" is waited out separately.
	needles := []string{
		"sorry, you have been blocked",
		"you have been blocked",
		"why have i been blocked",
		"access denied",
		"attention required",
		"cf-browser-verification",
		"enable javascript and cookies",
		"security service to protect",
		"unusual traffic",
		"automated requests",
		"request blocked",
	}
	for _, n := range needles {
		if strings.Contains(blob, n) {
			return true
		}
	}
	return false
}

func isTransientChallenge(title, body string) bool {
	blob := strings.ToLower(strings.TrimSpace(title + "\n" + body))
	return strings.Contains(blob, "checking your browser") ||
		strings.Contains(blob, "just a moment") ||
		strings.Contains(blob, "this will only take a few seconds")
}

func captureWithPlaywright(ctx context.Context, pageURL string, mobile bool) ([]byte, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("runtime caller unavailable")
	}
	script := filepath.Join(filepath.Dir(thisFile), "scripts", "seo_screenshot.mjs")
	if _, err := os.Stat(script); err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("node"); err != nil {
		return nil, err
	}
	mode := "desktop"
	if mobile {
		mode = "mobile"
	}
	cmd := exec.CommandContext(ctx, "node", script, pageURL, mode)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "blocked_by_waf") || (cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 3) {
			return nil, fmt.Errorf("bot blocked by website waf")
		}
		return nil, fmt.Errorf("playwright: %w (%s)", err, msg)
	}
	if stdout.Len() < 100 {
		return nil, fmt.Errorf("playwright returned empty screenshot")
	}
	return stdout.Bytes(), nil
}

func captureWithChromedp(ctx context.Context, pageURL string, mobile bool) ([]byte, error) {
	width, height := 1280, 800
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	if mobile {
		width, height = 390, 844
		ua = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1"
	}
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(ua),
		chromedp.WindowSize(width, height),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var png []byte
	var title, bodyText string

	viewportOpts := []chromedp.EmulateViewportOption{chromedp.EmulateScale(1)}
	wait := 2500 * time.Millisecond
	if mobile {
		viewportOpts = []chromedp.EmulateViewportOption{
			chromedp.EmulateScale(2),
			chromedp.EmulateMobile,
			chromedp.EmulateTouch,
			chromedp.EmulatePortrait,
		}
		wait = 2800 * time.Millisecond
	}

	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(int64(width), int64(height), viewportOpts...),
		chromedp.Navigate(pageURL),
		chromedp.Sleep(wait),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait out Cloudflare / bot interstitial when present.
			deadline := time.Now().Add(14 * time.Second)
			for time.Now().Before(deadline) {
				var t, b string
				if err := chromedp.Title(&t).Do(ctx); err != nil {
					return err
				}
				_ = chromedp.Evaluate(`document.body ? (document.body.innerText || '').slice(0, 2000) : ''`, &b).Do(ctx)
				if isBotBlockPage(t, b) {
					return fmt.Errorf("bot blocked by website waf")
				}
				if !isTransientChallenge(t, b) {
					return nil
				}
				if err := chromedp.Sleep(900 * time.Millisecond).Do(ctx); err != nil {
					return err
				}
			}
			return nil
		}),
		chromedp.Title(&title),
		chromedp.Evaluate(`document.body ? (document.body.innerText || '').slice(0, 12000) : ''`, &bodyText),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if isBotBlockPage(title, bodyText) || isTransientChallenge(title, bodyText) {
				return fmt.Errorf("bot blocked by website waf")
			}
			return nil
		}),
		chromedp.CaptureScreenshot(&png),
	)
	if err != nil {
		return nil, err
	}
	if len(png) < 100 {
		return nil, fmt.Errorf("empty chromedp screenshot")
	}
	return png, nil
}

func pngToJPEG(src []byte, quality int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
