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

type SignatureDetails struct {
	Name              string
	Title             string
	AdditionalDetails string
}

func DefaultSignatureDetails() SignatureDetails {
	return SignatureDetails{
		Name:  "Praveen Maurya",
		Title: "Business Development Manager",
	}
}

// EnsureTuviSignature adds the shared Tuvi signature at the final email
// boundary. Callers keep authoring plain-text content without signature markup.
func EnsureTuviSignature(req SendRequest) SendRequest {
	textBody := strings.TrimSpace(req.TextBody)
	htmlBody := strings.TrimSpace(req.HTMLBody)
	signature := DefaultSignatureDetails()
	if req.Signature != nil {
		signature = normalizeSignatureDetails(*req.Signature)
	}

	if textBody != "" && !hasTuviTextSignature(textBody, signature) {
		textBody += tuviTextSignature(signature)
	}
	if htmlBody == "" && textBody != "" {
		htmlBody = textEmailAsHTML(stripTuviTextSignature(textBody, signature), signature)
	} else if htmlBody != "" && !hasTuviHTMLSignature(htmlBody) {
		htmlBody = appendTuviHTMLSignature(htmlBody, signature)
	}

	req.TextBody = textBody
	req.HTMLBody = htmlBody
	return req
}

func hasTuviTextSignature(body string, signature SignatureDetails) bool {
	return strings.HasSuffix(strings.TrimSpace(body), strings.TrimSpace(tuviTextSignature(signature)))
}

func hasTuviHTMLSignature(body string) bool {
	return strings.Contains(body, tuviSignatureMarker) ||
		(strings.Contains(body, "Team Tuvi") && strings.Contains(body, "Tuvi Solutions"))
}

func stripTuviTextSignature(body string, signature SignatureDetails) string {
	trimmed := strings.TrimSpace(body)
	for _, suffix := range []string{
		strings.TrimSpace(tuviTextSignature(signature)),
		strings.TrimSpace(tuviTextSignature(DefaultSignatureDetails())),
		"--\nThanks & Regards,\nTeam Tuvi\nTuvi Solutions\n" + tuviWebsiteURL,
	} {
		if strings.HasSuffix(trimmed, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(trimmed, suffix))
		}
	}
	return trimmed
}

func tuviTextSignature(signature SignatureDetails) string {
	lines := []string{"Thanks & Regards,", signature.Name}
	if signature.Title != "" {
		lines = append(lines, signature.Title)
	}
	if signature.AdditionalDetails != "" {
		lines = append(lines, strings.Split(signature.AdditionalDetails, "\n")...)
	}
	lines = append(lines, "Tuvi Solutions", tuviWebsiteURL)
	return "\n\n" + strings.Join(lines, "\n")
}

func textEmailAsHTML(body string, signature SignatureDetails) string {
	escaped := html.EscapeString(strings.TrimSpace(body))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	escaped = strings.ReplaceAll(escaped, "\n", "<br>\n")
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<body style="margin:0;padding:0;background:#ffffff;font-family:Arial,Helvetica,sans-serif;color:#111111;">
  <div style="max-width:620px;font-size:15px;line-height:1.55;">%s</div>%s
</body>
</html>`, escaped, tuviHTMLSignature(signature))
}

func appendTuviHTMLSignature(body string, signature SignatureDetails) string {
	trimmed := strings.TrimRight(body, " \t\r\n")
	lower := strings.ToLower(trimmed)
	if index := strings.LastIndex(lower, "</body>"); index >= 0 {
		return trimmed[:index] + tuviHTMLSignature(signature) + "\n" + trimmed[index:]
	}
	return trimmed + tuviHTMLSignature(signature)
}

func tuviHTMLSignature(signature SignatureDetails) string {
	title := ""
	if signature.Title != "" {
		title = fmt.Sprintf(`<div style="margin-top:2px;font-size:15px;line-height:1.3;">%s</div>`, html.EscapeString(signature.Title))
	}
	details := ""
	if signature.AdditionalDetails != "" {
		escaped := html.EscapeString(signature.AdditionalDetails)
		escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
		escaped = strings.ReplaceAll(escaped, "\r", "\n")
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		details = fmt.Sprintf(`<div style="margin-top:4px;font-size:14px;line-height:1.35;">%s</div>`, escaped)
	}
	return fmt.Sprintf(`
  <div %s style="margin-top:24px;font-family:Arial,Helvetica,sans-serif;color:#111111;">
    <img src="%s" width="160" alt="Tuvi Solutions" style="display:block;width:160px;max-width:160px;height:auto;border:0;margin:0 0 8px 0;background:transparent;" />
    <div style="font-size:15px;line-height:1.3;font-weight:700;">Thanks &amp; Regards,</div>
    <div style="margin-top:2px;font-size:15px;line-height:1.3;font-weight:700;">%s</div>
    %s
    %s
    <div style="margin-top:4px;font-size:15px;line-height:1.3;font-weight:700;color:#d71920;">Tuvi Solutions</div>
    <div style="margin-top:4px;font-size:14px;line-height:1.3;"><a href="%s" style="color:#1a5fb4;text-decoration:underline;">www.tuvisolutions.com</a></div>
  </div>`, tuviSignatureMarker, tuviLogoURL, html.EscapeString(signature.Name), title, details, tuviWebsiteURL)
}

func normalizeSignatureDetails(signature SignatureDetails) SignatureDetails {
	if strings.TrimSpace(signature.Name) == "" && strings.TrimSpace(signature.Title) == "" && strings.TrimSpace(signature.AdditionalDetails) == "" {
		return DefaultSignatureDetails()
	}
	signature.Name = strings.TrimSpace(signature.Name)
	if signature.Name == "" {
		signature.Name = DefaultSignatureDetails().Name
	}
	signature.Title = strings.TrimSpace(signature.Title)
	signature.AdditionalDetails = strings.TrimSpace(strings.ReplaceAll(signature.AdditionalDetails, "\r\n", "\n"))
	signature.AdditionalDetails = strings.ReplaceAll(signature.AdditionalDetails, "\r", "\n")
	return signature
}
