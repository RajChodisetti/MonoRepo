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
	URL         htmltemplate.URL
}

type OutreachLinkConfig struct {
	PresentationURL string
	MarketingURL    string
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

func buildOutreachEmailData(restaurantName string, links OutreachLinkConfig) OutreachEmailData {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}
	presentationURL := strings.TrimSpace(links.PresentationURL)
	if presentationURL == "" {
		presentationURL = defaultPresentationURL
	}
	marketingURL := strings.TrimSpace(links.MarketingURL)
	if marketingURL == "" {
		marketingURL = defaultMarketingURL
	}

	// Use an internal marker so html/template does not URL-escape placeholder
	// braces; injectOutreachPlaceholders rewrites it after render.
	reservationDemoURL := htmltemplate.URL("__TEMPLATE_3_URL__")
	services := []OutreachServiceLink{
		{
			Title:       "AI Voice Receptionist",
			Description: "24/7 calls, reservation requests & callbacks",
			URL:         htmltemplate.URL(presentationURL),
		},
		{
			Title:       "Presentation Websites",
			Description: "Modern sites from your real menu & photos",
			URL:         htmltemplate.URL(presentationURL),
		},
		{
			Title:       "Reservation Requests",
			Description: "Guests submit table requests on your demo site",
			URL:         reservationDemoURL,
		},
		{
			Title:       "Custom Apps",
			Description: "QR ordering, loyalty & more",
			URL:         htmltemplate.URL(marketingURL),
		},
	}

	return OutreachEmailData{
		RestaurantName: name,
		ClickURL:       placeholderClickURL,
		UnsubscribeURL: placeholderUnsubscribeURL,
		AccentFallback: "#d4a853",
		Services:       services,
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
