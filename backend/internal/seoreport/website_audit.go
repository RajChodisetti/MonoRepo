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
	"net/mail"
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
	Listed           bool   // a website URL was present on the public Places listing
	Reachable        bool   // at least one real homepage viewport was captured
	ViewportCoverage string // desktop_and_mobile | desktop | mobile | none
	PublicEmail      string // valid mailto: link captured on the official page
	PublicPhone      string // valid tel: link captured on the official page
	PageEvidence     WebsitePageEvidence
	Source           string // "vision" | "fallback" | "social" | "aggregator" | "none"
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
	Title           string                `json:"title"`
	LoadedURL       string                `json:"loadedUrl"`
	HasMetaViewport bool                  `json:"hasMetaViewport"`
	Links           []capturedWebsiteLink `json:"links"`
	JSONLD          []string              `json:"jsonLd"`
}

// WebsitePageEvidence contains deterministic HTML/link signals captured from
// the reachable official page. These complement rather than replace the visual
// model review, and are safe to state precisely in report evidence.
type WebsitePageEvidence struct {
	Title           string
	LoadedScheme    string
	HasMetaViewport bool
	HasOrderCTA     bool
	HasMenuCTA      bool
	HasContactCTA   bool
}

type websiteCaptureArtifacts struct {
	DesktopDataURL string
	MobileDataURL  string
	VisionJPEG     []byte
	VisionViewport string
}

// buildWebsiteCaptureArtifacts preserves viewport provenance. When both
// captures exist, the LLM receives a single labeled-by-position composite:
// desktop on the left and mobile on the right. A successful capture is never
// published under the other viewport's field.
func buildWebsiteCaptureArtifacts(mobileJPEG, desktopJPEG []byte) websiteCaptureArtifacts {
	artifacts := websiteCaptureArtifacts{VisionViewport: "none"}
	if len(desktopJPEG) > 0 {
		artifacts.DesktopDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(desktopJPEG)
		artifacts.VisionJPEG = desktopJPEG
		artifacts.VisionViewport = "desktop"
	}
	if len(mobileJPEG) > 0 {
		artifacts.MobileDataURL = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(mobileJPEG)
		if len(artifacts.VisionJPEG) == 0 {
			artifacts.VisionJPEG = mobileJPEG
			artifacts.VisionViewport = "mobile"
		}
	}
	if len(desktopJPEG) > 0 && len(mobileJPEG) > 0 {
		if combined, err := combineWebsiteCaptures(desktopJPEG, mobileJPEG); err == nil {
			artifacts.VisionJPEG = combined
			artifacts.VisionViewport = "desktop_and_mobile"
		}
	}
	return artifacts
}

// combineWebsiteCaptures normalizes both viewports to the same height before
// placing desktop left and mobile right. Keeping this deterministic makes the
// visual-review provenance testable without adding another image dependency.
func combineWebsiteCaptures(desktopJPEG, mobileJPEG []byte) ([]byte, error) {
	desktop, _, err := image.Decode(bytes.NewReader(desktopJPEG))
	if err != nil {
		return nil, fmt.Errorf("decode desktop capture: %w", err)
	}
	mobile, _, err := image.Decode(bytes.NewReader(mobileJPEG))
	if err != nil {
		return nil, fmt.Errorf("decode mobile capture: %w", err)
	}

	const (
		panelHeight = 800
		gap         = 16
	)
	desktopWidth := scaledWidth(desktop.Bounds(), panelHeight)
	mobileWidth := scaledWidth(mobile.Bounds(), panelHeight)
	if desktopWidth <= 0 || mobileWidth <= 0 {
		return nil, fmt.Errorf("invalid capture dimensions")
	}
	canvas := image.NewRGBA(image.Rect(0, 0, desktopWidth+gap+mobileWidth, panelHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, 0, desktopWidth, panelHeight), scaleImageNearest(desktop, desktopWidth, panelHeight), image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(desktopWidth+gap, 0, desktopWidth+gap+mobileWidth, panelHeight), scaleImageNearest(mobile, mobileWidth, panelHeight), image.Point{}, draw.Src)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, canvas, &jpeg.Options{Quality: 84}); err != nil {
		return nil, fmt.Errorf("encode combined capture: %w", err)
	}
	return out.Bytes(), nil
}

func scaledWidth(bounds image.Rectangle, targetHeight int) int {
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 || targetHeight <= 0 {
		return 0
	}
	return maxInt(1, int(float64(bounds.Dx())/float64(bounds.Dy())*float64(targetHeight)+0.5))
}

func scaleImageNearest(src image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	bounds := src.Bounds()
	for y := 0; y < height; y++ {
		sourceY := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < width; x++ {
			sourceX := bounds.Min.X + x*bounds.Dx()/width
			dst.Set(x, y, src.At(sourceX, sourceY))
		}
	}
	return dst
}

func isSocialWebsite(website string) bool {
	_, ok := socialProfileFromURL(website, "places_website")
	return ok
}

// isLinkAggregatorWebsite recognizes Linktree as a listed destination without
// treating it as either a dedicated restaurant website or a social profile.
func isLinkAggregatorWebsite(website string) bool {
	website = strings.TrimSpace(website)
	if website == "" {
		return false
	}
	if !strings.Contains(website, "://") {
		website = "https://" + website
	}
	parsed, err := url.Parse(website)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil ||
		(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	return host == "linktr.ee" || strings.HasSuffix(host, ".linktr.ee")
}

// normalizeListedWebsiteURL preserves an explicit listed scheme. Scheme-less
// test/legacy input defaults to HTTPS, but an HTTP Places URL is attempted as
// HTTP and can still redirect naturally. Evidence records the final loaded URL.
func normalizeListedWebsiteURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return raw
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
	{Domain: "threads.net", Platform: "Threads"},
}

var blockedSocialPathSegments = map[string]struct{}{
	"about": {}, "account": {}, "accounts": {}, "ads": {}, "api": {},
	"auth": {}, "auth.php": {}, "authorize": {}, "business": {}, "compose": {},
	"developer": {}, "dialog": {}, "direct": {}, "directory": {}, "discover": {},
	"embed": {}, "events": {}, "explore": {}, "feed": {}, "foryou": {},
	"gaming": {}, "groups": {}, "help": {}, "home": {}, "intent": {},
	"jobs": {}, "l.php": {}, "learning": {}, "legal": {}, "link": {},
	"live": {}, "login": {}, "login.php": {}, "marketplace": {}, "messages": {},
	"messaging": {}, "mynetwork": {}, "notifications": {}, "oauth": {},
	"p": {}, "permalink.php": {}, "photo": {}, "photos": {}, "playlist": {},
	"plugins": {}, "post": {}, "posts": {}, "privacy": {}, "profile.php": {},
	"pulse": {}, "redir": {}, "redirect": {}, "reel": {}, "reels": {},
	"results": {}, "search": {}, "settings": {}, "share": {}, "share.php": {},
	"sharearticle": {}, "sharer": {}, "sharing": {}, "shorts": {}, "signin": {},
	"signup": {}, "status": {}, "statuses": {}, "stories": {}, "story": {},
	"story.php": {}, "terms": {}, "upload": {}, "video": {}, "videos": {},
	"watch": {}, "web": {},
}

var blockedSocialRedirectQueryKeys = map[string]struct{}{
	"continue": {}, "destination": {}, "href": {}, "next": {}, "q": {},
	"redirect": {}, "redirect_uri": {}, "redirect_url": {}, "target": {}, "u": {}, "url": {},
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
	if err != nil || parsed.Hostname() == "" || parsed.User != nil ||
		(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) {
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
	if len(segments) == 0 || hasBlockedSocialRedirectQuery(parsed.Query()) {
		return SocialProfile{}, false
	}
	for _, segment := range segments {
		if _, blocked := blockedSocialPathSegments[strings.ToLower(segment)]; blocked {
			return SocialProfile{}, false
		}
	}

	handle, canonicalPath, ok := canonicalSocialProfilePath(platform, segments)
	if !ok {
		return SocialProfile{}, false
	}

	parsed.Scheme = "https"
	parsed.Host = canonicalHost
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Path = canonicalPath
	canonical := strings.TrimSuffix(parsed.String(), "/")
	return SocialProfile{
		Platform: platform,
		Handle:   handle,
		URL:      canonical,
		Source:   source,
	}, true
}

func hasBlockedSocialRedirectQuery(values url.Values) bool {
	for key := range values {
		if _, blocked := blockedSocialRedirectQueryKeys[strings.ToLower(strings.TrimSpace(key))]; blocked {
			return true
		}
	}
	return false
}

// canonicalSocialProfilePath recognizes only canonical profile shapes for each
// platform. Content pages and action endpoints are deliberately
// excluded even when they contain a usable-looking account name.
func canonicalSocialProfilePath(platform string, segments []string) (string, string, bool) {
	switch platform {
	case "Instagram":
		if len(segments) != 1 || !validSocialIdentifier(segments[0], 30, "._") {
			return "", "", false
		}
		return segments[0], "/" + segments[0], true
	case "Facebook":
		if len(segments) == 1 && !strings.EqualFold(segments[0], "pages") &&
			validSocialIdentifier(segments[0], 100, "._-") {
			return segments[0], "/" + segments[0], true
		}
		// Preserve the older canonical business-page shape while still rejecting
		// arbitrary Facebook content paths.
		if len(segments) == 3 && strings.EqualFold(segments[0], "pages") &&
			validSocialIdentifier(segments[1], 100, "._-") && allASCIIDigits(segments[2]) {
			return segments[1], "/pages/" + segments[1] + "/" + segments[2], true
		}
	case "LinkedIn":
		if len(segments) != 2 {
			return "", "", false
		}
		kind := strings.ToLower(segments[0])
		if kind != "company" && kind != "in" && kind != "school" && kind != "showcase" {
			return "", "", false
		}
		if !validSocialIdentifier(segments[1], 100, "_-") {
			return "", "", false
		}
		return segments[1], "/" + kind + "/" + segments[1], true
	case "TikTok":
		if len(segments) != 1 || !strings.HasPrefix(segments[0], "@") {
			return "", "", false
		}
		handle := strings.TrimPrefix(segments[0], "@")
		if !validSocialIdentifier(handle, 32, "._") {
			return "", "", false
		}
		return handle, "/@" + handle, true
	case "YouTube":
		if len(segments) == 1 && strings.HasPrefix(segments[0], "@") {
			handle := strings.TrimPrefix(segments[0], "@")
			if validSocialIdentifier(handle, 30, "._-") {
				return handle, "/@" + handle, true
			}
			return "", "", false
		}
		if len(segments) != 2 {
			return "", "", false
		}
		kind := strings.ToLower(segments[0])
		if kind != "channel" && kind != "c" && kind != "user" {
			return "", "", false
		}
		if !validSocialIdentifier(segments[1], 100, "_-") {
			return "", "", false
		}
		return segments[1], "/" + kind + "/" + segments[1], true
	case "X":
		if len(segments) != 1 || !validSocialIdentifier(segments[0], 15, "_") {
			return "", "", false
		}
		return segments[0], "/" + segments[0], true
	case "Threads":
		if len(segments) != 1 || !strings.HasPrefix(segments[0], "@") {
			return "", "", false
		}
		handle := strings.TrimPrefix(segments[0], "@")
		if !validSocialIdentifier(handle, 30, "._") {
			return "", "", false
		}
		return handle, "/@" + handle, true
	}
	return "", "", false
}

func validSocialIdentifier(value string, maxLength int, extraASCII string) bool {
	if value == "" || len(value) > maxLength {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || strings.ContainsRune(extraASCII, char) {
			continue
		}
		return false
	}
	return true
}

func allASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
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
		// Link aggregators and arbitrary profile-shaped records are not social
		// proof. Re-parse each URL so only a verified canonical platform profile
		// reaches the report or earns listing-completeness points.
		canonical, valid := socialProfileFromURL(profile.URL, profile.Source)
		if !valid {
			continue
		}
		profile = canonical
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
	baseURL := strings.TrimSpace(signals.LoadedURL)
	if baseURL == "" {
		baseURL = pageURL
	}
	base, _ := url.Parse(baseURL)
	menu := MenuEvidence{
		Status:    "not_found",
		Source:    "website",
		Rationale: "No menu link or Menu JSON-LD was found on the inspected homepage. " + placesMenuEvidenceLimitation,
	}
	if absolute, ok := resolveEvidenceURL(base, pageURL); ok && looksLikeDirectMenuURL(absolute) {
		menu.Status = "present"
		menu.HasWebsiteLink = true
		menu.MenuURL = absolute
		menu.Source = "places_website_menu_url"
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

// extractPublicWebsiteContacts accepts only explicit mailto: and tel: anchors
// captured from the official page. Visible text, arbitrary page copy, and
// profile inventory are intentionally excluded from this public-contact signal.
func extractPublicWebsiteContacts(signals websitePageSignals) (string, string) {
	email := ""
	phone := ""
	for _, link := range signals.Links {
		if email == "" {
			if candidate, ok := publicEmailFromLink(link.Href); ok {
				email = candidate
			}
		}
		if phone == "" {
			if candidate, ok := publicPhoneFromLink(link.Href); ok {
				phone = candidate
			}
		}
		if email != "" && phone != "" {
			break
		}
	}
	return email, phone
}

func extractWebsitePageEvidence(pageURL string, signals websitePageSignals) WebsitePageEvidence {
	evidence := WebsitePageEvidence{
		Title:           strings.TrimSpace(signals.Title),
		HasMetaViewport: signals.HasMetaViewport,
		HasMenuCTA:      looksLikeDirectMenuURL(pageURL),
	}
	baseURL := strings.TrimSpace(signals.LoadedURL)
	if baseURL == "" {
		baseURL = pageURL
	}
	base, _ := url.Parse(baseURL)
	if loaded, err := url.Parse(strings.TrimSpace(signals.LoadedURL)); err == nil {
		switch strings.ToLower(loaded.Scheme) {
		case "http", "https":
			evidence.LoadedScheme = strings.ToLower(loaded.Scheme)
		}
	}
	for _, link := range signals.Links {
		blob := strings.ToLower(strings.TrimSpace(link.Text + " " + link.Href))
		if !evidence.HasMenuCTA {
			if absolute, ok := resolveEvidenceURL(base, link.Href); ok && looksLikeMenuLink(link, absolute) {
				evidence.HasMenuCTA = true
			}
		}
		if !evidence.HasOrderCTA {
			for _, marker := range []string{"order", "reserve", "reservation", "book a table", "book-now", "/book", "booking", "delivery", "takeaway", "takeout"} {
				if strings.Contains(blob, marker) {
					evidence.HasOrderCTA = true
					break
				}
			}
		}
		if !evidence.HasContactCTA {
			_, validEmail := publicEmailFromLink(link.Href)
			_, validPhone := publicPhoneFromLink(link.Href)
			evidence.HasContactCTA = validEmail || validPhone || strings.Contains(blob, "contact") || strings.Contains(blob, "call us") || strings.Contains(blob, "email us")
		}
	}
	return evidence
}

func publicEmailFromLink(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !strings.EqualFold(parsed.Scheme, "mailto") || parsed.Host != "" {
		return "", false
	}
	candidate := parsed.Opaque
	if candidate == "" {
		candidate = parsed.Path
	}
	candidate, err = url.PathUnescape(strings.TrimSpace(candidate))
	if err != nil || candidate == "" || len(candidate) > 254 || strings.ContainsAny(candidate, "\r\n,;") {
		return "", false
	}
	address, err := mail.ParseAddress(candidate)
	if err != nil || !strings.EqualFold(address.Address, candidate) {
		return "", false
	}
	parts := strings.Split(address.Address, "@")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" || !strings.Contains(parts[1], ".") {
		return "", false
	}
	return address.Address, true
}

func publicPhoneFromLink(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !strings.EqualFold(parsed.Scheme, "tel") || parsed.Host != "" {
		return "", false
	}
	candidate := parsed.Opaque
	if candidate == "" {
		candidate = parsed.Path
	}
	candidate, err = url.PathUnescape(strings.TrimSpace(candidate))
	if err != nil || candidate == "" || len(candidate) > 40 || strings.ContainsAny(candidate, "\r\n") {
		return "", false
	}
	if parameter := strings.IndexByte(candidate, ';'); parameter >= 0 {
		candidate = strings.TrimSpace(candidate[:parameter])
	}
	digits := 0
	for index, char := range candidate {
		switch {
		case char >= '0' && char <= '9':
			digits++
		case char == '+' && index == 0:
		case char == ' ' || char == '-' || char == '.' || char == '(' || char == ')':
		default:
			return "", false
		}
	}
	if digits < 7 || digits > 20 {
		return "", false
	}
	return candidate, true
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
	rawTarget := strings.TrimSpace(link.Href)
	if rawTarget == "" {
		return false
	}
	rawURL, err := url.Parse(rawTarget)
	if err != nil {
		return false
	}
	if rawURL.Scheme != "" && rawURL.Scheme != "http" && rawURL.Scheme != "https" {
		return false
	}
	// A fragment-only link is an in-page control, not independent menu evidence.
	if rawURL.Host == "" && rawURL.Path == "" && rawURL.RawQuery == "" && rawURL.Fragment != "" {
		return false
	}

	target, err := url.Parse(strings.TrimSpace(absolute))
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" || target.User != nil {
		return false
	}
	if blockedNonMenuPath(target.Path) {
		return false
	}
	if semanticMenuPath(target.Path) {
		return true
	}

	text := normalizedMenuLinkText(link.Text)
	if text == "" {
		return false
	}
	if menuDocumentPath(target.Path) && containsMenuWord(text) &&
		(text == "menu" || text == "menus" || !navigationMenuLabel(text)) {
		return true
	}
	if navigationMenuLabel(text) {
		return false
	}
	// Text evidence must be stronger than a bare "Menu" label and point to a
	// real path. This prevents homepage/query toggles and mislabeled About links
	// from becoming ten menu points.
	if target.Path == "" || target.Path == "/" {
		return false
	}
	return explicitFoodMenuLinkText(text)
}

func looksLikeDirectMenuURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
		return false
	}
	return !blockedNonMenuPath(parsed.Path) && semanticMenuPath(parsed.Path)
}

func semanticMenuPath(rawPath string) bool {
	pathValue, err := url.PathUnescape(strings.TrimSpace(rawPath))
	if err != nil || pathValue == "" || pathValue == "/" {
		return false
	}
	words := normalizedWords(pathValue)
	wordSet := make(map[string]struct{}, len(words))
	for _, word := range words {
		wordSet[word] = struct{}{}
	}
	for _, blocked := range []string{"toggle", "navigation", "navbar", "hamburger", "sidebar", "dropdown"} {
		if _, found := wordSet[blocked]; found {
			return false
		}
	}
	if _, found := wordSet["menu"]; found {
		return true
	}
	if _, found := wordSet["menus"]; found {
		return true
	}
	if _, found := wordSet["foodmenu"]; found {
		return true
	}
	_, hasFood := wordSet["food"]
	_, hasOur := wordSet["our"]
	_, hasDrink := wordSet["drink"]
	_, hasDrinks := wordSet["drinks"]
	if hasFood && (hasOur || hasDrink || hasDrinks) {
		return true
	}
	return false
}

func blockedNonMenuPath(rawPath string) bool {
	for _, word := range normalizedWords(rawPath) {
		switch word {
		case "about", "contact", "toggle", "navigation", "navbar", "hamburger", "sidebar", "dropdown", "mobile":
			return true
		}
	}
	return false
}

func menuDocumentPath(rawPath string) bool {
	lowerPath := strings.ToLower(strings.TrimSpace(rawPath))
	for _, extension := range []string{".pdf", ".doc", ".docx"} {
		if strings.HasSuffix(lowerPath, extension) {
			return true
		}
	}
	return false
}

func normalizedMenuLinkText(text string) string {
	return strings.Join(normalizedWords(text), " ")
}

func containsMenuWord(text string) bool {
	for _, word := range strings.Fields(text) {
		if word == "menu" || word == "menus" {
			return true
		}
	}
	return false
}

func navigationMenuLabel(text string) bool {
	if !containsMenuWord(text) {
		return false
	}
	for _, marker := range []string{"navigation", "navbar", "hamburger", "toggle", "open", "close", "mobile", "sidebar", "dropdown", "main", "site"} {
		for _, word := range strings.Fields(text) {
			if word == marker {
				return true
			}
		}
	}
	return text == "menu" || text == "menus"
}

func explicitFoodMenuLinkText(text string) bool {
	words := strings.Fields(text)
	wordSet := make(map[string]struct{}, len(words))
	for _, word := range words {
		wordSet[word] = struct{}{}
	}
	if containsMenuWord(text) {
		for _, qualifier := range []string{"food", "restaurant", "dinner", "lunch", "brunch", "breakfast", "drink", "drinks", "beverage", "dessert", "kids", "view", "see", "explore", "download", "our", "order", "online", "takeaway", "takeout", "delivery"} {
			if _, found := wordSet[qualifier]; found {
				return true
			}
		}
	}
	_, hasFood := wordSet["food"]
	_, hasOur := wordSet["our"]
	_, hasDrink := wordSet["drink"]
	_, hasDrinks := wordSet["drinks"]
	if hasFood && (hasOur || hasDrink || hasDrinks) {
		return true
	}
	_, hasOrder := wordSet["order"]
	if hasOrder {
		for _, qualifier := range []string{"online", "food", "takeaway", "takeout", "delivery"} {
			if _, found := wordSet[qualifier]; found {
				return true
			}
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
				if hasSubstantiveMenuStructure(raw) {
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
	case []any:
		for _, item := range typed {
			if resolved := structuredMenuURL(item, base); resolved != "" {
				return resolved
			}
		}
	case string:
		if resolved, ok := resolveStructuredMenuURL(base, typed); ok {
			return resolved
		}
	case map[string]any:
		for _, key := range []string{"url", "@id"} {
			if raw, ok := typed[key].(string); ok {
				if resolved, valid := resolveStructuredMenuURL(base, raw); valid {
					return resolved
				}
			}
		}
	}
	return ""
}

func resolveStructuredMenuURL(base *url.URL, raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.ContainsAny(raw, "\r\n\t ") {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Path == "" && parsed.Host == "" && (parsed.Fragment != "" || parsed.RawQuery != "")) {
		return "", false
	}
	// Reject arbitrary schema Text values that merely happen to parse as a
	// relative reference. A relative target must still look like a path, menu
	// slug, or supported document.
	if !parsed.IsAbs() && !strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "./") &&
		!strings.HasPrefix(raw, "../") && !strings.Contains(raw, "/") &&
		!semanticMenuPath("/"+parsed.Path) && !menuDocumentPath(parsed.Path) {
		return "", false
	}
	return resolveEvidenceURL(base, raw)
}

func hasSubstantiveMenuStructure(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if hasSubstantiveMenuStructure(item) {
				return true
			}
		}
	case map[string]any:
		if structuredTypeContainsMenu(typed["@type"]) {
			return true
		}
		if structuredTypeContainsMenuPart(typed["@type"]) && hasNonEmptyStructuredMap(typed) {
			return true
		}
		for key, raw := range typed {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "hasmenusection", "hasmenuitem", "menusection", "menuitem", "itemlistelement":
				if hasNonEmptyStructuredValue(raw) {
					return true
				}
			}
			if hasSubstantiveMenuStructure(raw) {
				return true
			}
		}
	}
	return false
}

func structuredTypeContainsMenuPart(value any) bool {
	switch typed := value.(type) {
	case string:
		typeName := strings.TrimSpace(typed)
		return strings.EqualFold(typeName, "MenuSection") || strings.EqualFold(typeName, "MenuItem")
	case []any:
		for _, item := range typed {
			if structuredTypeContainsMenuPart(item) {
				return true
			}
		}
	}
	return false
}

func hasNonEmptyStructuredValue(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		for _, item := range typed {
			if hasNonEmptyStructuredValue(item) {
				return true
			}
		}
	case map[string]any:
		return hasNonEmptyStructuredMap(typed)
	}
	return false
}

func hasNonEmptyStructuredMap(value map[string]any) bool {
	for key, item := range value {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "@type", "@context":
			continue
		}
		if hasNonEmptyStructuredValue(item) {
			return true
		}
	}
	return false
}

// AuditWebsite screenshots the homepage with bounded ChromeDP workers and
// scores design/UX via vision LLM. Strict by design.
func AuditWebsite(ctx context.Context, website string, llm llmlib.Client) WebsiteAudit {
	listedWebsite := strings.TrimSpace(website)
	if listedWebsite == "" {
		return WebsiteAudit{
			ViewportCoverage: "none",
			Source:           "none",
			MenuEvidence:     noWebsiteMenuEvidence(),
			SocialPresence:   noWebsiteSocialPresence(),
		}
	}
	website = normalizeListedWebsiteURL(listedWebsite)
	if isLinkAggregatorWebsite(listedWebsite) {
		return WebsiteAudit{
			Listed:           true,
			ViewportCoverage: "none",
			Review:           "This Google listing links to a Linktree aggregator, not a dedicated restaurant website. No website reachability or visual-quality points were awarded.",
			Source:           "aggregator",
			MenuEvidence: MenuEvidence{
				Status:    "not_found",
				Source:    "places_website",
				Rationale: "A Linktree destination is not a verified restaurant menu page. " + placesMenuEvidenceLimitation,
			},
			SocialPresence: SocialPresence{
				Status:    "not_found",
				Max:       3,
				Rationale: "Linktree is a link aggregator, not a verified social profile. No downstream canonical platform profile was independently discovered.",
			},
		}
	}

	if isSocialWebsite(listedWebsite) {
		profile, _ := socialProfileFromURL(listedWebsite, "places_website")
		return WebsiteAudit{
			Listed:           true,
			ViewportCoverage: "none",
			Review:           "This Google listing links to a social profile, not a dedicated restaurant website. No website reachability or visual-quality points were awarded.",
			Source:           "social",
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
		audit.Listed = true
		audit.Reachable = true
		audit.ViewportCoverage = captures.VisionViewport
		if audit.Source == "fallback" {
			audit.Review = "The live homepage was reached, but automated visual-quality analysis did not complete. Only reachability was scored; no visual-quality claim was made."
		}
		audit.Screenshot = captures.DesktopDataURL
		audit.MobileScreenshot = captures.MobileDataURL
		audit.MenuEvidence, audit.SocialPresence = extractWebsiteEvidence(website, signals)
		audit.PublicEmail, audit.PublicPhone = extractPublicWebsiteContacts(signals)
		audit.PageEvidence = extractWebsitePageEvidence(website, signals)
		return audit
	}

	if llm == nil || !llm.Enabled() {
		return attachShots(fallbackWebsiteAudit(website, "llm unavailable"))
	}

	visionCtx, visionCancel := context.WithTimeout(auditCtx, websiteVisionBudget)
	defer visionCancel()

	prompt := websiteVisionPrompt(captures.VisionViewport)

	raw, err := llm.CompleteVision(visionCtx, prompt, captures.VisionJPEG, "image/jpeg")
	if err != nil || strings.TrimSpace(raw) == "" {
		return attachShots(fallbackWebsiteAudit(website, fmt.Sprintf("vision failed: %v", err)))
	}

	return attachShots(websiteAuditFromVisionResponse(website, raw))
}

func websiteVisionPrompt(viewportCoverage string) string {
	viewportInstruction := "Only the desktop viewport is available. Evaluate desktop presentation only and do not claim mobile responsiveness was assessed."
	switch viewportCoverage {
	case "desktop_and_mobile":
		viewportInstruction = "The input is one side-by-side composite: desktop is on the LEFT and mobile is on the RIGHT. Evaluate and compare both viewports, including true mobile readability and CTA usability."
	case "mobile":
		viewportInstruction = "Only the mobile viewport is available. Evaluate mobile presentation only and do not claim desktop presentation was assessed."
	}
	return `You are auditing a restaurant website homepage screenshot for local SEO and guest conversion.
` + viewportInstruction + `
Be VERY STRICT. Most restaurant sites should score between 20 and 60 out of 100.
Scores above 65 are rare. Only an exceptional modern site with clear menu, order path, strong supplied-viewport usability, and strong branding deserves 70+.
Never give 90+ unless the site looks world-class.

Judge: visual design quality, trust, clarity of cuisine/location, menu visibility, order/reserve CTA, readability in only the supplied viewport(s), clutter, outdated templates, broken/empty hero, stock-photo feel.

Return ONLY compact JSON (no markdown):
{"score": <int 0-100>, "summary": "<2 short sentences>", "strengths": ["..."], "weaknesses": ["..."]}`
}

func fallbackWebsiteAudit(website, reason string) WebsiteAudit {
	return WebsiteAudit{
		Listed:           strings.TrimSpace(website) != "",
		Reachable:        false,
		ViewportCoverage: "none",
		Review:           "A website is listed, but its live homepage could not be verified as reachable. No reachability or visual-quality points were awarded.",
		Source:           "fallback",
		FailureReason:    strings.TrimSpace(reason),
		MenuEvidence:     unknownWebsiteMenuEvidence("The official website capture did not complete, so menu presence is unknown."),
		SocialPresence:   unknownWebsiteSocialPresence("The official website capture did not complete, so social presence is unknown."),
	}
}

func websiteAuditFromVisionResponse(website, raw string) WebsiteAudit {
	score, summary, valid := parseWebsiteVisionJSON(raw)
	if !valid {
		return fallbackWebsiteAudit(website, "vision response failed contract validation")
	}
	return WebsiteAudit{
		Listed:       strings.TrimSpace(website) != "",
		QualityScore: clampStrictWebsiteQuality(score),
		Review:       summary,
		Source:       "vision",
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

func parseWebsiteVisionJSON(raw string) (int, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", false
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil || len(object) != 4 {
		return 0, "", false
	}
	required := []string{"score", "summary", "strengths", "weaknesses"}
	for _, key := range required {
		if _, present := object[key]; !present {
			return 0, "", false
		}
	}

	var score int
	var summary string
	var strengths, weaknesses []string
	if err := json.Unmarshal(object["score"], &score); err != nil || score < 0 || score > 100 {
		return 0, "", false
	}
	if err := json.Unmarshal(object["summary"], &summary); err != nil {
		return 0, "", false
	}
	if err := json.Unmarshal(object["strengths"], &strengths); err != nil {
		return 0, "", false
	}
	if err := json.Unmarshal(object["weaknesses"], &weaknesses); err != nil {
		return 0, "", false
	}
	if strengths == nil || weaknesses == nil {
		return 0, "", false
	}
	summary = strings.TrimSpace(summary)
	if summary == "" || len([]rune(summary)) > 800 {
		return 0, "", false
	}
	return score, summary, true
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
		return chromedp.Evaluate(`(() => {
			const anchors = Array.from(document.querySelectorAll('a[href]'));
			const priority = anchors.filter((a) => /^(mailto:|tel:)/i.test(a.getAttribute('href') || '') || /(menu|food-and-drink|order|reserve|book|delivery|takeaway|takeout|contact|instagram|facebook|tiktok|youtube|threads|twitter|linkedin|x\.com|linktr\.ee)/i.test((a.getAttribute('href') || '') + ' ' + (a.innerText || ''))).slice(0, 80);
			return {
			title: String(document.title || '').trim().slice(0, 240),
			loadedUrl: String(window.location.href || '').slice(0, 2048),
			hasMetaViewport: Boolean(document.querySelector('meta[name="viewport" i][content]')),
			links: priority.concat(anchors.slice(0, 240)).map((a) => ({
				href: String(a.href || '').slice(0, 2048),
				text: String(a.innerText || a.getAttribute('aria-label') || '').trim().slice(0, 180)
			})),
			jsonLd: Array.from(document.querySelectorAll('script[type="application/ld+json"]')).slice(0, 30).map((node) => String(node.textContent || '').slice(0, 20000))
			};
		})()`, destination)
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
		if strings.TrimSpace(group.Title) != "" {
			merged.Title = strings.TrimSpace(group.Title)
		}
		if strings.TrimSpace(group.LoadedURL) != "" {
			merged.LoadedURL = strings.TrimSpace(group.LoadedURL)
		}
		merged.HasMetaViewport = merged.HasMetaViewport || group.HasMetaViewport
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
