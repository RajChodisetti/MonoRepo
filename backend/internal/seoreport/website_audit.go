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
	"sort"
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
	QualityScore     int    // 0–100, intentionally strict (most sites land 20–60)
	Review           string // short spoken/report summary
	Screenshot       string // data:image/jpeg;base64,... (desktop)
	MobileScreenshot string // data:image/jpeg;base64,... (phone viewport for scan mockup)
	Source           string // "vision" | "fallback" | "social" | "none"
	FailureReason    string // internal observability only; never serialized publicly
	MenuEvidence     MenuEvidence
	SocialPresence   SocialPresence
}

const placesMenuEvidenceLimitation = "Google Places exposes generic photos without Menu-section category metadata; only a verified website menu link or Menu JSON-LD counts as menu evidence."

type capturedWebsiteLink struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

type websitePageSignals struct {
	Links  []capturedWebsiteLink `json:"links"`
	JSONLD []string              `json:"jsonLd"`
}

type websiteCaptureArtifacts struct {
	DesktopDataURL string
	MobileDataURL  string
	VisionJPEG     []byte
}

// buildWebsiteCaptureArtifacts preserves viewport provenance. The vision
// review still prefers the desktop capture and falls back to mobile when the
// desktop capture is unavailable, but a successful capture is never published
// under the other viewport's field.
func buildWebsiteCaptureArtifacts(mobileJPEG, desktopJPEG []byte) websiteCaptureArtifacts {
	artifacts := websiteCaptureArtifacts{}
	if len(desktopJPEG) > 0 {
		artifacts.DesktopDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(desktopJPEG)
		artifacts.VisionJPEG = desktopJPEG
	}
	if len(mobileJPEG) > 0 {
		artifacts.MobileDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(mobileJPEG)
		if len(artifacts.VisionJPEG) == 0 {
			artifacts.VisionJPEG = mobileJPEG
		}
	}
	return artifacts
}

func isSocialWebsite(website string) bool {
	_, ok := socialProfileFromURL(website, "places_website")
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

func noWebsiteMenuEvidence() MenuEvidence {
	return MenuEvidence{
		Status:    "not_found",
		Source:    "places",
		Rationale: "No official website is listed, so no website menu could be verified. " + placesMenuEvidenceLimitation,
	}
}

func noWebsiteSocialPresence() SocialPresence {
	return SocialPresence{
		Status:    "not_found",
		Max:       3,
		Rationale: "No official website or social profile URL was available for verification.",
	}
}

func unknownWebsiteMenuEvidence(reason string) MenuEvidence {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "The official website could not be inspected."
	}
	return MenuEvidence{
		Status:    "unknown",
		Source:    "website",
		Rationale: reason + " " + placesMenuEvidenceLimitation,
	}
}

func unknownWebsiteSocialPresence(reason string) SocialPresence {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "The official website could not be inspected."
	}
	return SocialPresence{Status: "unknown", Max: 3, Rationale: reason}
}

var socialPlatformDomains = []struct {
	Domain   string
	Platform string
}{
	{Domain: "instagram.com", Platform: "Instagram"},
	{Domain: "facebook.com", Platform: "Facebook"},
	{Domain: "fb.com", Platform: "Facebook"},
	{Domain: "tiktok.com", Platform: "TikTok"},
	{Domain: "twitter.com", Platform: "X"},
	{Domain: "x.com", Platform: "X"},
	{Domain: "linkedin.com", Platform: "LinkedIn"},
	{Domain: "youtube.com", Platform: "YouTube"},
	{Domain: "youtu.be", Platform: "YouTube"},
	{Domain: "threads.net", Platform: "Threads"},
	{Domain: "linktr.ee", Platform: "Linktree"},
}

func socialProfileFromURL(rawURL, source string) (SocialProfile, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return SocialProfile{}, false
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return SocialProfile{}, false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	platform := ""
	canonicalHost := ""
	for _, candidate := range socialPlatformDomains {
		if host == candidate.Domain || strings.HasSuffix(host, "."+candidate.Domain) {
			platform = candidate.Platform
			canonicalHost = candidate.Domain
			break
		}
	}
	if platform == "" {
		return SocialProfile{}, false
	}

	segments := nonEmptyPathSegments(parsed.Path)
	if len(segments) == 0 {
		return SocialProfile{}, false
	}
	blocked := map[string]struct{}{
		"accounts": {}, "about": {}, "dialog": {}, "explore": {}, "home": {},
		"intent": {}, "login": {}, "p": {}, "plugins": {}, "reel": {},
		"share": {}, "sharer": {}, "watch": {},
	}
	first := strings.ToLower(segments[0])
	if _, skip := blocked[first]; skip {
		return SocialProfile{}, false
	}
	handle := strings.TrimPrefix(segments[0], "@")
	if (platform == "LinkedIn" || platform == "YouTube") && len(segments) > 1 {
		handle = strings.TrimPrefix(segments[1], "@")
	}
	if strings.TrimSpace(handle) == "" {
		return SocialProfile{}, false
	}

	parsed.Scheme = "https"
	parsed.Host = canonicalHost
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Path = "/" + strings.Join(segments, "/")
	canonical := strings.TrimSuffix(parsed.String(), "/")
	return SocialProfile{
		Platform: platform,
		Handle:   handle,
		URL:      canonical,
		Source:   source,
	}, true
}

func nonEmptyPathSegments(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func socialPresenceFromProfiles(profiles []SocialProfile) SocialPresence {
	seen := make(map[string]struct{}, len(profiles))
	unique := make([]SocialProfile, 0, len(profiles))
	for _, profile := range profiles {
		key := strings.ToLower(profile.Platform + "|" + profile.Handle + "|" + profile.URL)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, profile)
	}
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Platform == unique[j].Platform {
			return unique[i].Handle < unique[j].Handle
		}
		return unique[i].Platform < unique[j].Platform
	})
	if len(unique) == 0 {
		return SocialPresence{
			Status:    "not_found",
			Max:       3,
			Rationale: "No canonical social profile link was found on the official website.",
		}
	}
	return SocialPresence{
		Status:    "present",
		Max:       3,
		Profiles:  unique,
		Rationale: fmt.Sprintf("Found %d canonical social profile link(s) from the official web presence.", len(unique)),
	}
}

func extractWebsiteEvidence(pageURL string, signals websitePageSignals) (MenuEvidence, SocialPresence) {
	base, _ := url.Parse(pageURL)
	menu := MenuEvidence{
		Status:    "not_found",
		Source:    "website",
		Rationale: "No menu link or Menu JSON-LD was found on the inspected homepage. " + placesMenuEvidenceLimitation,
	}
	profiles := make([]SocialProfile, 0, 4)
	for _, link := range signals.Links {
		absolute, ok := resolveEvidenceURL(base, link.Href)
		if !ok {
			continue
		}
		if profile, social := socialProfileFromURL(absolute, "website_link"); social {
			profiles = append(profiles, profile)
		}
		if !menu.HasWebsiteLink && looksLikeMenuLink(link, absolute) {
			menu.Status = "present"
			menu.HasWebsiteLink = true
			menu.MenuURL = absolute
			menu.Source = "website_link"
		}
	}
	for _, raw := range signals.JSONLD {
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			continue
		}
		menuURL, found := findStructuredMenu(value, base)
		if !found {
			continue
		}
		menu.Status = "present"
		menu.HasStructuredData = true
		if menu.MenuURL == "" {
			menu.MenuURL = menuURL
		}
		if menu.Source == "website" {
			menu.Source = "website_jsonld"
		} else if menu.Source != "website_jsonld" {
			menu.Source = "website_link_and_jsonld"
		}
		break
	}
	if menu.Status == "present" {
		parts := make([]string, 0, 2)
		if menu.HasWebsiteLink {
			parts = append(parts, "a menu link")
		}
		if menu.HasStructuredData {
			parts = append(parts, "Menu JSON-LD")
		}
		menu.Rationale = "Verified " + strings.Join(parts, " and ") + " on the official website. " + placesMenuEvidenceLimitation
	}
	return menu, socialPresenceFromProfiles(profiles)
}

func resolveEvidenceURL(base *url.URL, raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil {
		return "", false
	}
	if base != nil {
		parsed = base.ResolveReference(parsed)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false
	}
	if parsed.Hostname() == "" || parsed.User != nil {
		return "", false
	}
	parsed.Fragment = ""
	return parsed.String(), true
}

func looksLikeMenuLink(link capturedWebsiteLink, absolute string) bool {
	blob := strings.ToLower(strings.Join([]string{link.Text, absolute}, " "))
	for _, marker := range []string{"menu", "food & drink", "food-and-drink", "our food", "our-food"} {
		if strings.Contains(blob, marker) {
			return true
		}
	}
	return false
}

func findStructuredMenu(value any, base *url.URL) (string, bool) {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if menuURL, ok := findStructuredMenu(item, base); ok {
				return menuURL, true
			}
		}
	case map[string]any:
		if structuredTypeContainsMenu(typed["@type"]) {
			return structuredMenuURL(typed, base), true
		}
		for key, raw := range typed {
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			if lowerKey == "hasmenu" || lowerKey == "menu" {
				if menuURL := structuredMenuURL(raw, base); menuURL != "" {
					return menuURL, true
				}
				if raw != nil {
					return "", true
				}
			}
			if menuURL, ok := findStructuredMenu(raw, base); ok {
				return menuURL, true
			}
		}
	}
	return "", false
}

func structuredTypeContainsMenu(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "Menu")
	case []any:
		for _, item := range typed {
			if structuredTypeContainsMenu(item) {
				return true
			}
		}
	}
	return false
}

func structuredMenuURL(value any, base *url.URL) string {
	switch typed := value.(type) {
	case string:
		if resolved, ok := resolveEvidenceURL(base, typed); ok {
			return resolved
		}
	case map[string]any:
		for _, key := range []string{"url", "@id"} {
			if raw, ok := typed[key].(string); ok {
				if resolved, valid := resolveEvidenceURL(base, raw); valid {
					return resolved
				}
			}
		}
	}
	return ""
}

// AuditWebsite screenshots the homepage with bounded ChromeDP workers and
// scores design/UX via vision LLM. Strict by design.
func AuditWebsite(ctx context.Context, website string, llm llmlib.Client) WebsiteAudit {
	website = strings.TrimSpace(website)
	if website == "" {
		return WebsiteAudit{
			Source:         "none",
			MenuEvidence:   noWebsiteMenuEvidence(),
			SocialPresence: noWebsiteSocialPresence(),
		}
	}
	if !strings.Contains(website, "://") {
		website = "https://" + website
	}
	website = normalizeWebsiteURL(website)

	if isSocialWebsite(website) {
		profile, _ := socialProfileFromURL(website, "places_website")
		return WebsiteAudit{
			QualityScore: 28,
			Review:       "This Google listing links to a social profile, not a dedicated restaurant website. Expect weaker branding, menus, and direct ordering.",
			Source:       "social",
			MenuEvidence: MenuEvidence{
				Status:    "not_found",
				Source:    "places_website",
				Rationale: "The listing website is a social profile, not a verified menu page. " + placesMenuEvidenceLimitation,
			},
			SocialPresence: socialPresenceFromProfiles([]SocialProfile{profile}),
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
	mobileJPEG, desktopJPEG, signals, shotErr := captureWebsiteJPEGPairWithSignals(shotCtx, website)
	shotCancel()

	captures := buildWebsiteCaptureArtifacts(mobileJPEG, desktopJPEG)
	if len(captures.VisionJPEG) == 0 {
		return fallbackWebsiteAudit(website, fmt.Sprintf("screenshot failed: %v", shotErr))
	}

	attachShots := func(audit WebsiteAudit) WebsiteAudit {
		audit.Screenshot = captures.DesktopDataURL
		audit.MobileScreenshot = captures.MobileDataURL
		audit.MenuEvidence, audit.SocialPresence = extractWebsiteEvidence(website, signals)
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

	raw, err := llm.CompleteVision(visionCtx, prompt, captures.VisionJPEG, "image/jpeg")
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
	score := 32
	if hasHTTPSWebsite(website) {
		score = 38
	}
	return WebsiteAudit{
		QualityScore:   score,
		Review:         "We could not fully review the live homepage visuals, so this is a conservative estimate. A dedicated site helps, but design, menu clarity, and order CTAs still need a human-quality pass.",
		Source:         "fallback",
		FailureReason:  strings.TrimSpace(reason),
		MenuEvidence:   unknownWebsiteMenuEvidence("The official website capture did not complete, so menu presence is unknown."),
		SocialPresence: unknownWebsiteSocialPresence("The official website capture did not complete, so social presence is unknown."),
	}
}

func clampStrictWebsiteQuality(score int) int {
	score = clamp(score, 0, 100)
	// The prompt is deliberately strict, but exceptional results must remain
	// reachable so the downstream 20-point website metric can reach its maximum.
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

func captureWebsiteJPEGPair(ctx context.Context, pageURL string) ([]byte, []byte, error) {
	mobileJPEG, desktopJPEG, _, err := captureWebsiteJPEGPairWithSignals(ctx, pageURL)
	return mobileJPEG, desktopJPEG, err
}

func captureWebsiteJPEGPairWithSignals(ctx context.Context, pageURL string) ([]byte, []byte, websitePageSignals, error) {
	select {
	case websiteCaptureSlots <- struct{}{}:
		defer func() { <-websiteCaptureSlots }()
	case <-ctx.Done():
		return nil, nil, websitePageSignals{}, ctx.Err()
	}

	mobilePNG, desktopPNG, signals, err := capturePairWithChromedpSignals(ctx, pageURL)
	if err != nil {
		return nil, nil, signals, err
	}
	var mobileJPEG, desktopJPEG []byte
	if len(mobilePNG) > 0 {
		mobileJPEG, err = pngToJPEG(mobilePNG, 82)
		if err != nil {
			return nil, nil, signals, err
		}
	}
	if len(desktopPNG) > 0 {
		desktopJPEG, err = pngToJPEG(desktopPNG, 82)
		if err != nil && len(mobileJPEG) == 0 {
			return nil, nil, signals, err
		}
	}
	return mobileJPEG, desktopJPEG, signals, nil
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

func capturePairWithChromedpSignals(ctx context.Context, pageURL string) ([]byte, []byte, websitePageSignals, error) {
	validationCtx, validationCancel := context.WithTimeout(ctx, websiteDNSBudget)
	_, err := validatePublicWebsiteURL(validationCtx, pageURL, net.DefaultResolver)
	validationCancel()
	if err != nil {
		return nil, nil, websitePageSignals{}, fmt.Errorf("website destination blocked: %w", err)
	}
	proxy, err := startSafeBrowserProxy(net.DefaultResolver)
	if err != nil {
		return nil, nil, websitePageSignals{}, fmt.Errorf("start safe browser proxy: %w", err)
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
	var mobileSignals, desktopSignals websitePageSignals
	collectSignals := func(destination *websitePageSignals) chromedp.Action {
		return chromedp.Evaluate(`(() => ({
			links: Array.from(document.querySelectorAll('a[href]')).slice(0, 240).map((a) => ({
				href: String(a.href || '').slice(0, 2048),
				text: String(a.innerText || a.getAttribute('aria-label') || '').trim().slice(0, 180)
			})),
			jsonLd: Array.from(document.querySelectorAll('script[type="application/ld+json"]')).slice(0, 30).map((node) => String(node.textContent || '').slice(0, 20000))
		}))()`, destination)
	}
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
	mobileActions = append(mobileActions, collectSignals(&mobileSignals), chromedp.CaptureScreenshot(&mobilePNG))
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
	desktopActions = append(desktopActions, collectSignals(&desktopSignals), chromedp.CaptureScreenshot(&desktopPNG))
	desktopErr := chromedp.Run(browserCtx, desktopActions)
	if len(desktopPNG) < 100 {
		desktopPNG = nil
		if desktopErr == nil {
			desktopErr = fmt.Errorf("empty desktop chromedp screenshot")
		}
	}
	signals := mergeWebsiteSignals(mobileSignals, desktopSignals)
	if len(mobilePNG) == 0 && len(desktopPNG) == 0 {
		return nil, nil, signals, fmt.Errorf("mobile capture: %v; desktop capture: %v", mobileErr, desktopErr)
	}
	return mobilePNG, desktopPNG, signals, nil
}

func mergeWebsiteSignals(groups ...websitePageSignals) websitePageSignals {
	merged := websitePageSignals{}
	seenLinks := make(map[string]struct{})
	seenJSON := make(map[string]struct{})
	for _, group := range groups {
		for _, link := range group.Links {
			key := strings.TrimSpace(link.Href) + "\x00" + strings.TrimSpace(link.Text)
			if key == "\x00" {
				continue
			}
			if _, exists := seenLinks[key]; exists {
				continue
			}
			seenLinks[key] = struct{}{}
			merged.Links = append(merged.Links, link)
		}
		for _, raw := range group.JSONLD {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			if _, exists := seenJSON[raw]; exists {
				continue
			}
			seenJSON[raw] = struct{}{}
			merged.JSONLD = append(merged.JSONLD, raw)
		}
	}
	return merged
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
