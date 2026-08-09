package seoreport

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

const (
	websiteAuditBudget         = 13 * time.Second
	websiteCaptureBudget       = 9 * time.Second
	websiteVisionBudget        = 3 * time.Second
	maxParallelWebsiteCaptures = 2
)

// Browser work is the heaviest part of a report. Each report uses one browser
// session for both viewports, so two slots match the report-generation limit
// without starting four Chromium process trees on a two-CPU production host.
var websiteCaptureSlots = make(chan struct{}, maxParallelWebsiteCaptures)

// WebsiteAudit is the visual review of a restaurant website homepage.
type WebsiteAudit struct {
	QualityScore     int    // 0–100 when a visual review completed; 0 for unavailable audits
	Review           string // short spoken/report summary
	Screenshot       string // data:image/jpeg;base64,... (desktop)
	MobileScreenshot string // data:image/jpeg;base64,... (phone viewport for scan mockup)
	Source           string // "vision" | "fallback" | "social" | "none"
	FailureReason    string // internal observability only; never serialized publicly
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
			QualityScore: 0,
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
	mobileJPEG, desktopJPEG, shotErr := captureWebsiteJPEGPair(shotCtx, website)
	shotCancel()

	displayJPEG := desktopJPEG
	if len(displayJPEG) == 0 {
		displayJPEG = mobileJPEG
	}
	if len(displayJPEG) == 0 {
		return fallbackWebsiteAudit(website, fmt.Sprintf("screenshot failed: %v", shotErr))
	}
	desktopDataURL, mobileDataURL := websiteCaptureDataURLs(mobileJPEG, desktopJPEG)

	attachShots := func(audit WebsiteAudit) WebsiteAudit {
		audit.Screenshot = desktopDataURL
		audit.MobileScreenshot = mobileDataURL
		return audit
	}

	if llm == nil || !llm.Enabled() {
		return attachShots(fallbackWebsiteAudit(website, "llm unavailable"))
	}

	visionCtx, visionCancel := context.WithTimeout(auditCtx, websiteVisionBudget)
	defer visionCancel()

	visionJPEG := displayJPEG
	viewDescription := "one available homepage capture"
	if combined, combineErr := combineWebsiteAuditJPEG(mobileJPEG, desktopJPEG); combineErr == nil && len(combined) > 0 {
		visionJPEG = combined
		if len(mobileJPEG) > 0 && len(desktopJPEG) > 0 {
			viewDescription = "a side-by-side mobile capture (left) and desktop capture (right)"
		}
	}

	prompt := `You are auditing ` + viewDescription + ` of a restaurant website for local SEO and guest conversion.
Score only what is visible. Do not infer hidden pages, performance, accessibility conformance, or features outside the captures.
Use this explicit 100-point rubric: responsive mobile/desktop usability 20; cuisine, location, and menu clarity 20; order/reserve CTA clarity 20; trust and contact cues 15; visual hierarchy and brand quality 15; absence of broken, empty, cluttered, or misleading content 10.
Use the full 0-100 range supported by the evidence: 0-20 unusable/broken, 21-40 major gaps, 41-60 functional but average, 61-80 strong, 81-100 exceptional.

Return ONLY compact JSON (no markdown):
{"score": <int 0-100>, "summary": "<2 short sentences>", "strengths": ["..."], "weaknesses": ["..."]}`

	raw, err := llm.CompleteVision(visionCtx, prompt, visionJPEG, "image/jpeg")
	if err != nil || strings.TrimSpace(raw) == "" {
		return attachShots(fallbackWebsiteAudit(website, fmt.Sprintf("vision failed: %v", err)))
	}

	score, summary, valid := parseWebsiteVisionJSON(raw)
	if !valid {
		return attachShots(fallbackWebsiteAudit(website, "vision returned invalid score JSON"))
	}
	score = clampWebsiteQuality(score)
	if summary == "" {
		summary = "Homepage visual quality is average for a local restaurant site; tighten design, menu access, and booking CTAs."
	}
	return attachShots(WebsiteAudit{
		QualityScore: score,
		Review:       summary,
		Source:       "vision",
	})
}

func fallbackWebsiteAudit(_ string, reason string) WebsiteAudit {
	return WebsiteAudit{
		// Do not manufacture a visual-quality number when capture or review did
		// not complete. scoreWebsite separately credits verified site presence.
		QualityScore:  0,
		Review:        "We could not fully review the live homepage visuals, so no visual-quality score was assigned. The report credits only the dedicated website signals that were directly observed.",
		Source:        "fallback",
		FailureReason: strings.TrimSpace(reason),
	}
}

func clampWebsiteQuality(score int) int {
	// The review prompt is already strict. Preserve its evidence instead of
	// forcing every answer into an artificial 20–60 cluster.
	return clamp(score, 0, 100)
}

func websiteCaptureDataURLs(mobileJPEG, desktopJPEG []byte) (desktopDataURL, mobileDataURL string) {
	if len(desktopJPEG) > 0 {
		desktopDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(desktopJPEG)
	}
	if len(mobileJPEG) > 0 {
		mobileDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(mobileJPEG)
	}
	return desktopDataURL, mobileDataURL
}

func parseWebsiteVisionJSON(raw string) (int, string, bool) {
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
		Score      *int     `json:"score"`
		Summary    string   `json:"summary"`
		Strengths  []string `json:"strengths"`
		Weaknesses []string `json:"weaknesses"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || parsed.Score == nil {
		return 0, "", false
	}
	summary := strings.TrimSpace(parsed.Summary)
	if summary == "" && len(parsed.Weaknesses) > 0 {
		summary = strings.Join(parsed.Weaknesses, " ")
	}
	return *parsed.Score, summary, true
}

// combineWebsiteAuditJPEG gives the vision reviewer both captured viewports in
// one bounded image because the provider accepts a single image per request.
func combineWebsiteAuditJPEG(mobileJPEG, desktopJPEG []byte) ([]byte, error) {
	if len(mobileJPEG) == 0 {
		return desktopJPEG, nil
	}
	if len(desktopJPEG) == 0 {
		return mobileJPEG, nil
	}
	mobile, _, err := image.Decode(bytes.NewReader(mobileJPEG))
	if err != nil {
		return nil, fmt.Errorf("decode mobile capture: %w", err)
	}
	desktop, _, err := image.Decode(bytes.NewReader(desktopJPEG))
	if err != nil {
		return nil, fmt.Errorf("decode desktop capture: %w", err)
	}

	const gutter = 16
	mobileBounds := mobile.Bounds()
	desktopBounds := desktop.Bounds()
	width := mobileBounds.Dx() + gutter + desktopBounds.Dx()
	height := max(mobileBounds.Dy(), desktopBounds.Dy())
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, 0, mobileBounds.Dx(), mobileBounds.Dy()), mobile, mobileBounds.Min, draw.Src)
	desktopX := mobileBounds.Dx() + gutter
	draw.Draw(canvas, image.Rect(desktopX, 0, desktopX+desktopBounds.Dx(), desktopBounds.Dy()), desktop, desktopBounds.Min, draw.Src)

	var combined bytes.Buffer
	if err := jpeg.Encode(&combined, canvas, &jpeg.Options{Quality: 82}); err != nil {
		return nil, fmt.Errorf("encode combined captures: %w", err)
	}
	return combined.Bytes(), nil
}

func captureWebsiteJPEGPair(ctx context.Context, pageURL string) ([]byte, []byte, error) {
	select {
	case websiteCaptureSlots <- struct{}{}:
		defer func() { <-websiteCaptureSlots }()
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	mobilePNG, desktopPNG, err := capturePairWithChromedp(ctx, pageURL)
	if err != nil {
		return nil, nil, err
	}
	var mobileJPEG, desktopJPEG []byte
	if len(mobilePNG) > 0 {
		mobileJPEG, err = pngToJPEG(mobilePNG, 82)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(desktopPNG) > 0 {
		desktopJPEG, err = pngToJPEG(desktopPNG, 82)
		if err != nil && len(mobileJPEG) == 0 {
			return nil, nil, err
		}
	}
	return mobileJPEG, desktopJPEG, nil
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

func capturePairWithChromedp(ctx context.Context, pageURL string) ([]byte, []byte, error) {
	validationCtx, validationCancel := context.WithTimeout(ctx, websiteDNSBudget)
	_, err := validatePublicWebsiteURL(validationCtx, pageURL, net.DefaultResolver)
	validationCancel()
	if err != nil {
		return nil, nil, fmt.Errorf("website destination blocked: %w", err)
	}
	proxy, err := startSafeBrowserProxy(net.DefaultResolver)
	if err != nil {
		return nil, nil, fmt.Errorf("start safe browser proxy: %w", err)
	}
	defer proxy.Close()

	const (
		mobileWidth   = 390
		mobileHeight  = 844
		desktopWidth  = 1280
		desktopHeight = 800
		mobileUA      = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1"
		desktopUA     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	)
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
		chromedp.UserAgent(mobileUA),
		chromedp.WindowSize(mobileWidth, mobileHeight),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()
	installBrowserRequestGuard(browserCtx, browserCancel, net.DefaultResolver)

	var mobilePNG, desktopPNG []byte
	waitForUsablePage := func(wait time.Duration, title, bodyText *string) chromedp.Tasks {
		return chromedp.Tasks{
			chromedp.Sleep(wait),
			chromedp.ActionFunc(func(ctx context.Context) error {
				// Wait out Cloudflare / bot interstitial when present.
				deadline := time.Now().Add(2500 * time.Millisecond)
				for time.Now().Before(deadline) {
					var currentTitle, currentBody string
					if err := chromedp.Title(&currentTitle).Do(ctx); err != nil {
						return err
					}
					_ = chromedp.Evaluate(`document.body ? (document.body.innerText || '').slice(0, 2000) : ''`, &currentBody).Do(ctx)
					if isBotBlockPage(currentTitle, currentBody) {
						return fmt.Errorf("bot blocked by website waf")
					}
					if !isTransientChallenge(currentTitle, currentBody) {
						return nil
					}
					if err := chromedp.Sleep(900 * time.Millisecond).Do(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			chromedp.Title(title),
			chromedp.Evaluate(`document.body ? (document.body.innerText || '').slice(0, 12000) : ''`, bodyText),
			chromedp.ActionFunc(func(context.Context) error {
				if isBotBlockPage(*title, *bodyText) || isTransientChallenge(*title, *bodyText) {
					return fmt.Errorf("bot blocked by website waf")
				}
				return nil
			}),
		}
	}

	var mobileTitle, mobileBody string
	mobileActions := chromedp.Tasks{
		fetch.Enable(),
		chromedp.EmulateViewport(mobileWidth, mobileHeight,
			chromedp.EmulateScale(2),
			chromedp.EmulateMobile,
			chromedp.EmulateTouch,
			chromedp.EmulatePortrait,
		),
		chromedp.Navigate(pageURL),
	}
	mobileActions = append(mobileActions, waitForUsablePage(1200*time.Millisecond, &mobileTitle, &mobileBody)...)
	mobileActions = append(mobileActions, chromedp.CaptureScreenshot(&mobilePNG))
	mobileErr := chromedp.Run(browserCtx, mobileActions)
	if len(mobilePNG) < 100 {
		mobilePNG = nil
		if mobileErr == nil {
			mobileErr = fmt.Errorf("empty mobile chromedp screenshot")
		}
	}

	// Reuse the sandboxed process and proxy, but reload after switching the UA
	// and viewport so server-rendered and load-time desktop variants are real.
	var desktopTitle, desktopBody string
	desktopActions := chromedp.Tasks{
		emulation.SetUserAgentOverride(desktopUA),
		chromedp.EmulateViewport(desktopWidth, desktopHeight, chromedp.EmulateScale(1)),
		chromedp.Navigate(pageURL),
	}
	desktopActions = append(desktopActions, waitForUsablePage(900*time.Millisecond, &desktopTitle, &desktopBody)...)
	desktopActions = append(desktopActions, chromedp.CaptureScreenshot(&desktopPNG))
	desktopErr := chromedp.Run(browserCtx, desktopActions)
	if len(desktopPNG) < 100 {
		desktopPNG = nil
		if desktopErr == nil {
			desktopErr = fmt.Errorf("empty desktop chromedp screenshot")
		}
	}
	if len(mobilePNG) == 0 && len(desktopPNG) == 0 {
		return nil, nil, fmt.Errorf("mobile capture: %v; desktop capture: %v", mobileErr, desktopErr)
	}
	return mobilePNG, desktopPNG, nil
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
