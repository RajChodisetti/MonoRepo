package campaigns

import (
	"bytes"
	"embed"
	"fmt"
	"html"
	htmltemplate "html/template"
	"os"
	"strings"
	texttemplate "text/template"
)

//go:embed templates/*
var outreachTemplates embed.FS

const (
	placeholderClickURL       = "{{CLICK_URL}}"
	placeholderTemplate1URL   = "{{TEMPLATE_1_URL}}"
	placeholderTemplate2URL   = "{{TEMPLATE_2_URL}}"
	placeholderTemplate3URL   = "{{TEMPLATE_3_URL}}"
	placeholderUnsubscribeURL = "{{UNSUBSCRIBE_URL}}"

	defaultPresentationURL = "http://localhost:5500"
	defaultMarketingURL    = "http://localhost:3001"
)

type OutreachServiceLink struct {
	Title       string
	Description string
	// URL is raw href content (may include tracking placeholders like {{CLICK_URL}}).
	URL htmltemplate.URL
}

type OutreachEmailData struct {
	RestaurantName string
	ClickURL       string
	UnsubscribeURL string
	AccentFallback string
	Services       []OutreachServiceLink
}

func outreachSubject(restaurantName string) string {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}
	return fmt.Sprintf("A live demo for %s — AI receptionist, website & more", name)
}

func presentationSiteURL() string {
	if v := strings.TrimSpace(os.Getenv("PRESENTATION_SITE_URL")); v != "" {
		return v
	}
	return defaultPresentationURL
}

func marketingSiteURL() string {
	if v := strings.TrimSpace(os.Getenv("PUBLIC_MARKETING_URL")); v != "" {
		return v
	}
	return defaultMarketingURL
}

func buildOutreachEmailData(restaurantName string) OutreachEmailData {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}

	// Use __CLICK_URL__ (not {{CLICK_URL}}) so html/template does not URL-escape braces;
	// injectOutreachPlaceholders rewrites these after render.
	demoURL := htmltemplate.URL("__CLICK_URL__")
	services := []OutreachServiceLink{
		{
			Title:       "AI Voice Receptionist",
			Description: "24/7 calls, bookings & callbacks",
			URL:         demoURL,
		},
		{
			Title:       "Presentation Websites",
			Description: "Modern sites from your real menu & photos",
			URL:         htmltemplate.URL(presentationSiteURL()),
		},
		{
			Title:       "Online Reservations",
			Description: "Guests book tables on your demo site",
			URL:         demoURL,
		},
		{
			Title:       "Custom Apps",
			Description: "QR ordering, loyalty & more",
			URL:         htmltemplate.URL(marketingSiteURL()),
		},
	}

	return OutreachEmailData{
		RestaurantName: html.EscapeString(name),
		ClickURL:       placeholderClickURL,
		UnsubscribeURL: placeholderUnsubscribeURL,
		AccentFallback: "#d4a853",
		Services:       services,
	}
}

func RenderOutreachEmail(restaurantName string) (DraftContent, error) {
	data := buildOutreachEmailData(restaurantName)

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

	return DraftContent{
		Subject:  outreachSubject(restaurantName),
		BodyHTML: htmlBody,
		BodyText: textBody,
	}, nil
}

func injectOutreachPlaceholders(htmlBody string) string {
	return strings.NewReplacer(
		"__CLICK_URL__", placeholderClickURL,
		"__UNSUBSCRIBE_URL__", placeholderUnsubscribeURL,
		"__TEMPLATE_1_URL__", placeholderTemplate1URL,
		"__TEMPLATE_2_URL__", placeholderTemplate2URL,
		"__TEMPLATE_3_URL__", placeholderTemplate3URL,
	).Replace(htmlBody)
}
