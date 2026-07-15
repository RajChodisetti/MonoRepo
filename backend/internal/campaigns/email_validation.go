package campaigns

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	unresolvedEmailPlaceholder = regexp.MustCompile(`(?:\{\{[^{}\r\n]+\}\}|__(?:CLICK_URL|UNSUBSCRIBE_URL|TEMPLATE_[123]_URL)__)`)
	emailHTMLURLAttribute      = regexp.MustCompile(`(?i)(?:href|src)\s*=\s*["']([^"']+)["']`)
	emailTextURL               = regexp.MustCompile(`(?i)https?://[^\s<>"']+`)
)

func ValidateRenderedEmail(content DraftContent, requireHTTPS bool, allowedHosts ...string) error {
	combined := content.Subject + "\n" + content.BodyHTML + "\n" + content.BodyText
	if unresolvedEmailPlaceholder.MatchString(combined) {
		return fmt.Errorf("email contains an unresolved template placeholder")
	}
	if !requireHTTPS {
		return nil
	}

	var rawURLs []string
	for _, match := range emailHTMLURLAttribute.FindAllStringSubmatch(content.BodyHTML, -1) {
		if len(match) == 2 {
			rawURLs = append(rawURLs, match[1])
		}
	}
	for _, match := range emailTextURL.FindAllString(content.BodyHTML, -1) {
		rawURLs = append(rawURLs, strings.TrimRight(match, ".,;:!?)"))
	}
	for _, match := range emailTextURL.FindAllString(content.BodyText, -1) {
		rawURLs = append(rawURLs, strings.TrimRight(match, ".,;:!?)"))
	}

	allowed := make(map[string]struct{}, len(allowedHosts))
	for _, host := range allowedHosts {
		host = strings.ToLower(strings.TrimSpace(host))
		if host != "" {
			allowed[host] = struct{}{}
		}
	}

	for _, raw := range rawURLs {
		parsed, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" || parsed.User != nil {
			return fmt.Errorf("email contains an invalid public link")
		}
		if !strings.EqualFold(parsed.Scheme, "https") {
			return fmt.Errorf("email link for host %q must use https", parsed.Hostname())
		}
		if emailLinkHostIsLocal(parsed.Hostname()) {
			return fmt.Errorf("email link must not use local host %q", parsed.Hostname())
		}
		if len(allowed) > 0 {
			if parsed.Port() != "" {
				return fmt.Errorf("email link host %q must not use a custom port", parsed.Hostname())
			}
			if _, ok := allowed[strings.ToLower(parsed.Hostname())]; !ok {
				return fmt.Errorf("email link host %q is not approved", parsed.Hostname())
			}
			if err := validateCanonicalTuviLink(parsed); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCanonicalTuviLink(parsed *url.URL) error {
	host := strings.ToLower(parsed.Hostname())
	path := parsed.EscapedPath()
	switch host {
	case "api.tuvisolutions.com":
		for _, prefix := range []string{"/t/click/", "/t/open/", "/t/unsubscribe/"} {
			if strings.HasPrefix(path, prefix) && parsed.RawQuery == "" && parsed.Fragment == "" {
				return nil
			}
		}
	case "demo.tuvisolutions.com":
		if (path == "" || path == "/" || strings.HasPrefix(path, "/demo/")) && parsed.Fragment == "" {
			return nil
		}
	case "tuvisolutions.com":
		canonicalPath := strings.TrimSuffix(path, "/")
		if (canonicalPath == "" || canonicalPath == "/services/restaurants") && parsed.RawQuery == "" && parsed.Fragment == "" {
			return nil
		}
	default:
		// Callers may supply a different approved-host set outside production.
		return nil
	}
	return fmt.Errorf("email link for host %q does not use an approved production path", parsed.Hostname())
}

func emailLinkHostIsLocal(host string) bool {
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}
