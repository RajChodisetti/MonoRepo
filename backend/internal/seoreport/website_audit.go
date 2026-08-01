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
	QualityScore int    // 0–100, intentionally strict (most sites land 20–60)
	Review       string // short spoken/report summary
	Screenshot   string // data:image/jpeg;base64,...
	Source       string // "vision" | "fallback" | "social" | "none"
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
	if isSocialWebsite(website) {
		return WebsiteAudit{
			QualityScore: 28,
			Review:       "This Google listing links to a social profile, not a dedicated restaurant website. Expect weaker branding, menus, and direct ordering.",
			Source:       "social",
		}
	}

	shotCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()

	jpegBytes, err := captureWebsiteJPEG(shotCtx, website)
	if err != nil || len(jpegBytes) == 0 {
		return fallbackWebsiteAudit(website, fmt.Sprintf("screenshot failed: %v", err))
	}
	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpegBytes)

	if llm == nil || !llm.Enabled() {
		audit := fallbackWebsiteAudit(website, "llm unavailable")
		audit.Screenshot = dataURL
		return audit
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
		audit := fallbackWebsiteAudit(website, fmt.Sprintf("vision failed: %v", err))
		audit.Screenshot = dataURL
		return audit
	}

	score, summary := parseWebsiteVisionJSON(raw)
	score = clampStrictWebsiteQuality(score)
	if summary == "" {
		summary = "Homepage visual quality is average for a local restaurant site; tighten design, menu access, and booking CTAs."
	}
	return WebsiteAudit{
		QualityScore: score,
		Review:       summary,
		Screenshot:   dataURL,
		Source:       "vision",
	}
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

func captureWebsiteJPEG(ctx context.Context, pageURL string) ([]byte, error) {
	// Prefer Playwright script when Node + playwright are available.
	if png, err := captureWithPlaywright(ctx, pageURL); err == nil && len(png) > 0 {
		return pngToJPEG(png, 82)
	}
	png, err := captureWithChromedp(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	return pngToJPEG(png, 82)
}

func captureWithPlaywright(ctx context.Context, pageURL string) ([]byte, error) {
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
	cmd := exec.CommandContext(ctx, "node", script, pageURL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("playwright: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	if stdout.Len() < 100 {
		return nil, fmt.Errorf("playwright returned empty screenshot")
	}
	return stdout.Bytes(), nil
}

func captureWithChromedp(ctx context.Context, pageURL string) ([]byte, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1280, 800),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var png []byte
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(pageURL),
		chromedp.Sleep(1800*time.Millisecond),
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
