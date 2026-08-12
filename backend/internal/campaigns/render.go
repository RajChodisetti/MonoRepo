package campaigns

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/*
var outreachTemplates embed.FS

const (
	placeholderClickURL     = "{{CLICK_URL}}"
	placeholderTemplate1URL = "{{TEMPLATE_1_URL}}"
	placeholderTemplate2URL = "{{TEMPLATE_2_URL}}"
	placeholderTemplate3URL = "{{TEMPLATE_3_URL}}"

	defaultPresentationURL = "http://localhost:5500"
	defaultWebsiteURL      = "https://tuvisolutions.com"

	// Already live at this path (web/public/brand/); a fixed production URL
	// regardless of environment, since a recipient's inbox can never load an
	// image from a dev/staging host.
	outreachLogoURL         = "https://tuvisolutions.com/brand/tuvi-solutions-logo.png"
	outreachSignatureMarker = `data-tuvi-outreach-signature="true"`
)

type OutreachServiceLink struct {
	Title       string
	Description string
	URL         htmltemplate.URL
}

type OutreachLinkConfig struct {
	PresentationURL string
	MarketingURL    string
}

type OutreachEmailData struct {
	RestaurantName string
	ClickURL       string
	AccentFallback string
	Services       []OutreachServiceLink
	WebsiteURL     string
	LogoURL        string
}

func outreachSubject(restaurantName string) string {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}
	return fmt.Sprintf("A live demo for %s — AI receptionist, website & more", name)
}

func buildOutreachEmailData(restaurantName string, links OutreachLinkConfig) OutreachEmailData {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}
	presentationURL := strings.TrimSpace(links.PresentationURL)
	if presentationURL == "" {
		presentationURL = defaultPresentationURL
	}
	websiteURL := strings.TrimSpace(links.MarketingURL)
	if websiteURL == "" {
		websiteURL = defaultWebsiteURL
	}

	// Use an internal marker so html/template does not URL-escape placeholder
	// braces; injectOutreachPlaceholders rewrites it after render.
	services := []OutreachServiceLink{
		{
			Title:       "Elysian personalized website",
			Description: "A polished reservation-focused experience for your guests",
			URL:         htmltemplate.URL("__TEMPLATE_3_URL__"),
		},
		{
			Title:       "Aurora personalized website",
			Description: "A warm, modern alternative using the same restaurant content",
			URL:         htmltemplate.URL("__TEMPLATE_2_URL__"),
		},
		{
			Title:       "Cinematic personalized website",
			Description: "A bold, story-led website built from your restaurant details",
			URL:         htmltemplate.URL("__TEMPLATE_1_URL__"),
		},
		{
			Title:       "Tuvi restaurant services presentation",
			Description: "See the complete AI receptionist, website, reservation, and growth offering",
			URL:         htmltemplate.URL(presentationURL),
		},
	}

	return OutreachEmailData{
		RestaurantName: name,
		ClickURL:       placeholderClickURL,
		AccentFallback: "#d4a853",
		Services:       services,
		WebsiteURL:     websiteURL,
		LogoURL:        outreachLogoURL,
	}
}

func RenderOutreachEmail(restaurantName string) (DraftContent, error) {
	return RenderOutreachEmailWithLinks(restaurantName, OutreachLinkConfig{})
}

func RenderOutreachEmailWithLinks(restaurantName string, links OutreachLinkConfig) (DraftContent, error) {
	data := buildOutreachEmailData(restaurantName, links)

	htmlTmpl, err := htmltemplate.ParseFS(outreachTemplates, "templates/outreach.html")
	if err != nil {
		return DraftContent{}, fmt.Errorf("parse outreach html template: %w", err)
	}
	textTmpl, err := texttemplate.ParseFS(outreachTemplates, "templates/outreach.txt")
	if err != nil {
		return DraftContent{}, fmt.Errorf("parse outreach text template: %w", err)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return DraftContent{}, fmt.Errorf("render outreach html: %w", err)
	}
	htmlBody := injectOutreachPlaceholders(htmlBuf.String())

	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return DraftContent{}, fmt.Errorf("render outreach text: %w", err)
	}
	textBody := injectOutreachPlaceholders(textBuf.String())

	return EnsureOutreachSignature(DraftContent{
		Subject:  outreachSubject(restaurantName),
		BodyHTML: htmlBody,
		BodyText: textBody,
	}), nil
}

func injectOutreachPlaceholders(htmlBody string) string {
	return strings.NewReplacer(
		"__CLICK_URL__", placeholderClickURL,
		"__TEMPLATE_1_URL__", placeholderTemplate1URL,
		"__TEMPLATE_2_URL__", placeholderTemplate2URL,
		"__TEMPLATE_3_URL__", placeholderTemplate3URL,
	).Replace(htmlBody)
}

func EnsureOutreachSignature(content DraftContent) DraftContent {
	data := buildOutreachEmailData("", OutreachLinkConfig{})
	content.BodyHTML = strings.ReplaceAll(content.BodyHTML, "Tuvi Marketing Team", "Team Tuvi")
	content.BodyText = strings.ReplaceAll(content.BodyText, "Tuvi Marketing Team", "Team Tuvi")
	if strings.TrimSpace(content.BodyHTML) == "" && strings.TrimSpace(content.BodyText) != "" {
		content.BodyHTML = outreachTextBodyAsHTML(content.BodyText, data)
	} else if strings.TrimSpace(content.BodyHTML) != "" && !outreachHTMLHasSignature(content.BodyHTML) {
		content.BodyHTML = appendOutreachSignatureHTML(content.BodyHTML, data)
	}
	if strings.TrimSpace(content.BodyText) != "" && !outreachTextHasSignature(content.BodyText) {
		content.BodyText = strings.TrimRight(content.BodyText, " \t\r\n") + outreachSignatureText(data)
	}
	return content
}

func appendOutreachSignatureHTML(body string, data OutreachEmailData) string {
	trimmed := strings.TrimRight(body, " \t\r\n")
	signature := outreachSignatureHTML(data)
	lower := strings.ToLower(trimmed)
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return trimmed[:idx] + signature + "\n" + trimmed[idx:]
	}
	return trimmed + signature
}

func outreachTextBodyAsHTML(textBody string, data OutreachEmailData) string {
	escaped := htmltemplate.HTMLEscapeString(strings.TrimSpace(textBody))
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\n")
	escaped = strings.ReplaceAll(escaped, "\n", "<br />\n")
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#ffffff;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#111111;">
  <div style="font-size:15px;line-height:1.55;color:#111111;">
    %s
  </div>%s
</body>
</html>`, escaped, outreachSignatureHTML(data))
}

func outreachHTMLHasSignature(body string) bool {
	return strings.Contains(body, outreachSignatureMarker) || strings.Contains(body, "Team Tuvi")
}

func outreachTextHasSignature(body string) bool {
	return strings.Contains(body, "Team Tuvi") && strings.Contains(body, "Tuvi Solutions")
}

func outreachSignatureHTML(data OutreachEmailData) string {
	logoURL := htmltemplate.HTMLEscapeString(data.LogoURL)
	websiteURL := htmltemplate.HTMLEscapeString(data.WebsiteURL)
	return fmt.Sprintf(`

              <!-- Signature -->
              <div %s style="margin-top:26px;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
                <img src="%s" width="108" height="108" alt="Tuvi Solutions" style="display:block;width:108px;height:108px;border:0;margin:0 0 14px 0;" />
                <p style="margin:0;font-size:15px;line-height:1.2;font-weight:700;color:#111111;">Thanks &amp; Regards,</p>
                <p style="margin:2px 0 0;font-size:15px;line-height:1.2;font-weight:700;color:#111111;">Team Tuvi</p>
                <p style="margin:4px 0 0;font-size:15px;line-height:1.2;font-weight:700;color:#d71920;">Tuvi Solutions</p>
                <p style="margin:4px 0 0;font-size:14px;line-height:1.25;">
                  <a href="%s" style="color:#1a5fb4;text-decoration:underline;">www.tuvisolutions.com</a>
                </p>
              </div>`, outreachSignatureMarker, logoURL, websiteURL)
}

func outreachSignatureText(data OutreachEmailData) string {
	websiteURL := strings.TrimSpace(data.WebsiteURL)
	if websiteURL == "" {
		websiteURL = defaultWebsiteURL
	}
	return fmt.Sprintf("\n\n--\nThanks & Regards,\nTeam Tuvi\nTuvi Solutions\n%s\n", websiteURL)
}
