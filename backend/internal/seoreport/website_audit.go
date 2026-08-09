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
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

const (
	websiteAuditBudget         = 10 * time.Second
	websiteCaptureBudget       = 6500 * time.Millisecond
	websiteVisionBudget        = 3200 * time.Millisecond
	maxParallelWebsiteCaptures = 4
)

// Browser work is the heaviest part of a report. Four slots allow the mobile
// and desktop views of two reports to run together without unbounded Chromium
// fan-out under traffic spikes.
var websiteCaptureSlots = make(chan struct{}, maxParallelWebsiteCaptures)

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

// AuditWebsite screenshots the homepage with bounded ChromeDP workers and
// scores design/UX via vision LLM. Strict by design.
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

	auditCtx, auditCancel := context.WithTimeout(ctx, websiteAuditBudget)
	defer auditCancel()
	validationCtx, validationCancel := context.WithTimeout(auditCtx, websiteDNSBudget)
	_, validationErr := validatePublicWebsiteURL(validationCtx, website, net.DefaultResolver)
	validationCancel()
	if validationErr != nil {
		return fallbackWebsiteAudit(website, fmt.Sprintf("unsafe website destination: %v", validationErr))
	}
	shotCtx, shotCancel := context.WithTimeout(auditCtx, websiteCaptureBudget)

	// Mobile-first: iPhone UA often clears CF faster, and the scan phone mockup needs it.
	type shotResult struct {
		mobile bool
		jpeg   []byte
		err    error
	}
	shots := make(chan shotResult, 2)
	for _, mobile := range []bool{true, false} {
		mobile := mobile
		go func() {
			b, err := captureWebsiteJPEG(shotCtx, website, mobile)
			shots <- shotResult{mobile: mobile, jpeg: b, err: err}
		}()
	}

	var mobileRes, deskRes shotResult
	for received := 0; received < 2; {
		select {
		case result := <-shots:
			if result.mobile {
				mobileRes = result
			} else {
				deskRes = result
			}
			received++
		case <-shotCtx.Done():
			received = 2
		}
	}
	shotCancel()

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

	visionCtx, visionCancel := context.WithTimeout(auditCtx, websiteVisionBudget)
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
	select {
	case websiteCaptureSlots <- struct{}{}:
		defer func() { <-websiteCaptureSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
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

func captureWithChromedp(ctx context.Context, pageURL string, mobile bool) ([]byte, error) {
	validationCtx, validationCancel := context.WithTimeout(ctx, websiteDNSBudget)
	_, err := validatePublicWebsiteURL(validationCtx, pageURL, net.DefaultResolver)
	validationCancel()
	if err != nil {
		return nil, fmt.Errorf("website destination blocked: %w", err)
	}
	proxy, err := startSafeBrowserProxy(net.DefaultResolver)
	if err != nil {
		return nil, fmt.Errorf("start safe browser proxy: %w", err)
	}
	defer proxy.Close()

	width, height := 1280, 800
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	if mobile {
		width, height = 390, 844
		ua = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1"
	}
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-dns-prefetch", true),
		chromedp.Flag("disable-preconnect", true),
		chromedp.Flag("disable-quic", true),
		chromedp.Flag("site-per-process", true),
		chromedp.Flag("enable-features", "NetworkService"),
		chromedp.Flag("disable-features", "Translate,BlinkGenPropertyTrees,ServiceWorker"),
		chromedp.Flag("force-webrtc-ip-handling-policy", "disable_non_proxied_udp"),
		chromedp.Flag("proxy-bypass-list", "<-loopback>"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.ProxyServer(proxy.URL()),
		chromedp.UserAgent(ua),
		chromedp.WindowSize(width, height),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()
	installBrowserRequestGuard(browserCtx, browserCancel, net.DefaultResolver)

	var png []byte
	var title, bodyText string

	viewportOpts := []chromedp.EmulateViewportOption{chromedp.EmulateScale(1)}
	wait := 900 * time.Millisecond
	if mobile {
		viewportOpts = []chromedp.EmulateViewportOption{
			chromedp.EmulateScale(2),
			chromedp.EmulateMobile,
			chromedp.EmulateTouch,
			chromedp.EmulatePortrait,
		}
		wait = 1200 * time.Millisecond
	}

	err = chromedp.Run(browserCtx,
		fetch.Enable(),
		chromedp.EmulateViewport(int64(width), int64(height), viewportOpts...),
		chromedp.Navigate(pageURL),
		chromedp.Sleep(wait),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait out Cloudflare / bot interstitial when present.
			deadline := time.Now().Add(2500 * time.Millisecond)
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

const (
	browserRequestGuardWorkers = 8
	browserRequestGuardQueue   = 256
)

// installBrowserRequestGuard pauses every network request before it is sent.
// Redirect hops and subresources produce their own Fetch.requestPaused event.
// The safe proxy remains the final socket-level enforcement boundary, so a DNS
// answer changing after this check still cannot redirect Chromium internally.
func installBrowserRequestGuard(
	browserCtx context.Context,
	cancelBrowser context.CancelFunc,
	resolver websiteHostResolver,
) {
	requests := make(chan *fetch.EventRequestPaused, browserRequestGuardQueue)
	for range browserRequestGuardWorkers {
		go func() {
			for {
				select {
				case <-browserCtx.Done():
					return
				case event := <-requests:
					if event == nil || event.Request == nil {
						continue
					}
					chromedpContext := chromedp.FromContext(browserCtx)
					if chromedpContext == nil || chromedpContext.Target == nil {
						cancelBrowser()
						return
					}
					executorCtx := cdp.WithExecutor(browserCtx, chromedpContext.Target)
					_ = guardBrowserRequest(
						executorCtx,
						resolver,
						event.Request.URL,
						func(ctx context.Context) error {
							return fetch.ContinueRequest(event.RequestID).Do(ctx)
						},
						func(ctx context.Context) error {
							return fetch.FailRequest(event.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
						},
					)
				}
			}
		}()
	}
	chromedp.ListenTarget(browserCtx, func(event any) {
		paused, ok := event.(*fetch.EventRequestPaused)
		if !ok {
			return
		}
		select {
		case requests <- paused:
		default:
			// An adversarial page cannot create unbounded validation goroutines.
			// Cancel the capture and let the report return its bounded fallback.
			cancelBrowser()
		}
	})
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
