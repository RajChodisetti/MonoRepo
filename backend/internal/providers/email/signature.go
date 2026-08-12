package email

import (
	"fmt"
	"html"
	"strings"
)

const (
	tuviWebsiteURL      = "https://tuvisolutions.com"
	tuviLogoURL         = "https://tuvisolutions.com/brand/tuvi-solutions-logo-transparent.png"
	tuviSignatureMarker = `data-tuvi-email-signature="true"`
)

// EnsureTuviSignature adds the shared Tuvi signature at the final email
// boundary. Callers keep authoring plain-text content without signature markup.
func EnsureTuviSignature(req SendRequest) SendRequest {
	textBody := strings.TrimSpace(req.TextBody)
	htmlBody := strings.TrimSpace(req.HTMLBody)

	if textBody != "" && !hasTuviTextSignature(textBody) {
		textBody += tuviTextSignature()
	}
	if htmlBody == "" && textBody != "" {
		htmlBody = textEmailAsHTML(stripTuviTextSignature(textBody))
	} else if htmlBody != "" && !hasTuviHTMLSignature(htmlBody) {
		htmlBody = appendTuviHTMLSignature(htmlBody)
	}

	req.TextBody = textBody
	req.HTMLBody = htmlBody
	return req
}

func hasTuviTextSignature(body string) bool {
	return strings.Contains(body, "Team Tuvi") && strings.Contains(body, "Tuvi Solutions")
}

func hasTuviHTMLSignature(body string) bool {
	return strings.Contains(body, tuviSignatureMarker) ||
		(strings.Contains(body, "Team Tuvi") && strings.Contains(body, "Tuvi Solutions"))
}

func stripTuviTextSignature(body string) string {
	trimmed := strings.TrimSpace(body)
	for _, suffix := range []string{
		strings.TrimSpace(tuviTextSignature()),
		"--\nThanks & Regards,\nTeam Tuvi\nTuvi Solutions\n" + tuviWebsiteURL,
	} {
		if strings.HasSuffix(trimmed, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(trimmed, suffix))
		}
	}
	return trimmed
}

func tuviTextSignature() string {
	return "\n\nThanks & Regards,\nTeam Tuvi\nTuvi Solutions\n" + tuviWebsiteURL
}

func textEmailAsHTML(body string) string {
	escaped := html.EscapeString(strings.TrimSpace(body))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	escaped = strings.ReplaceAll(escaped, "\n", "<br>\n")
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<body style="margin:0;padding:0;background:#ffffff;font-family:Arial,Helvetica,sans-serif;color:#111111;">
  <div style="max-width:620px;font-size:15px;line-height:1.55;">%s</div>%s
</body>
</html>`, escaped, tuviHTMLSignature())
}

func appendTuviHTMLSignature(body string) string {
	trimmed := strings.TrimRight(body, " \t\r\n")
	lower := strings.ToLower(trimmed)
	if index := strings.LastIndex(lower, "</body>"); index >= 0 {
		return trimmed[:index] + tuviHTMLSignature() + "\n" + trimmed[index:]
	}
	return trimmed + tuviHTMLSignature()
}

func tuviHTMLSignature() string {
	return fmt.Sprintf(`
  <div %s style="margin-top:24px;font-family:Arial,Helvetica,sans-serif;color:#111111;">
    <img src="%s" width="160" alt="Tuvi Solutions" style="display:block;width:160px;max-width:160px;height:auto;border:0;margin:0 0 8px 0;background:transparent;" />
    <div style="font-size:15px;line-height:1.3;font-weight:700;">Thanks &amp; Regards,</div>
    <div style="margin-top:2px;font-size:15px;line-height:1.3;font-weight:700;">Team Tuvi</div>
    <div style="margin-top:4px;font-size:15px;line-height:1.3;font-weight:700;color:#d71920;">Tuvi Solutions</div>
    <div style="margin-top:4px;font-size:14px;line-height:1.3;"><a href="%s" style="color:#1a5fb4;text-decoration:underline;">www.tuvisolutions.com</a></div>
  </div>`, tuviSignatureMarker, tuviLogoURL, tuviWebsiteURL)
}
