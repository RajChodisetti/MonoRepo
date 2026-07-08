package campaigns

import (
	"bytes"
	"embed"
	"fmt"
	"html"
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
)

type OutreachServiceBullet struct {
	Title       string
	Description string
}

type OutreachTemplateCard struct {
	Number             string
	Name               string
	Tagline            string
	Bullets            []string
	AccentColor        string
	PreviewPlaceholder string
	PreviewURL         string
}

type OutreachEmailData struct {
	RestaurantName string
	ClickURL       string
	UnsubscribeURL string
	AccentFallback string
	ServiceRows    [][]OutreachServiceBullet
	Templates      []OutreachTemplateCard
}

var outreachServices = []OutreachServiceBullet{
	{
		Title:       "AI Voice Receptionist",
		Description: "Answers calls 24/7, handles reservations, and offers callbacks — so you never miss a booking.",
	},
	{
		Title:       "Presentation Websites",
		Description: "Modern sites built from your real menu, photos, hours, and reviews.",
	},
	{
		Title:       "Online Reservations",
		Description: "Table booking integrated into your site — guests reserve without picking up the phone.",
	},
	{
		Title:       "Custom Apps",
		Description: "QR ordering, loyalty programs, and other tools tailored to how you run service.",
	},
}

var outreachTemplateCards = []struct {
	Number             string
	Name               string
	Tagline            string
	Bullets            []string
	AccentColor        string
	PreviewPlaceholder string
}{
	{
		Number:             "1",
		Name:               "Cinematic",
		Tagline:            "Warm editorial dining with scroll storytelling.",
		Bullets:            []string{"Elegant hero and menu sections", "Ideal for classic and upscale dining"},
		AccentColor:        "#c9a96e",
		PreviewPlaceholder: "TEMPLATE_1_URL",
	},
	{
		Number:             "2",
		Name:               "Aurora",
		Tagline:            "Futuristic glass design with a bold, modern feel.",
		Bullets:            []string{"Interactive sections and motion", "Great for trend-forward brands"},
		AccentColor:        "#22d3ee",
		PreviewPlaceholder: "TEMPLATE_2_URL",
	},
	{
		Number:             "3",
		Name:               "Elysian",
		Tagline:            "Ultra-premium gold and black fine dining aesthetic.",
		Bullets:            []string{"Curated dish and gallery focus", "Built for high-end hospitality"},
		AccentColor:        "#D4AF37",
		PreviewPlaceholder: "TEMPLATE_3_URL",
	},
}

func outreachSubject(restaurantName string) string {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}
	return fmt.Sprintf("We built a live demo for %s — website, AI receptionist & more", name)
}

func buildOutreachEmailData(restaurantName string) OutreachEmailData {
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = "your restaurant"
	}

	cards := make([]OutreachTemplateCard, 0, len(outreachTemplateCards))
	for _, card := range outreachTemplateCards {
		cards = append(cards, OutreachTemplateCard{
			Number:             card.Number,
			Name:               card.Name,
			Tagline:            card.Tagline,
			Bullets:            append([]string(nil), card.Bullets...),
			AccentColor:        card.AccentColor,
			PreviewPlaceholder: card.PreviewPlaceholder,
			PreviewURL:         "{{" + card.PreviewPlaceholder + "}}",
		})
	}

	services := make([]OutreachServiceBullet, len(outreachServices))
	copy(services, outreachServices)

	serviceRows := [][]OutreachServiceBullet{
		{services[0], services[1]},
		{services[2], services[3]},
	}

	return OutreachEmailData{
		RestaurantName: html.EscapeString(name),
		ClickURL:       placeholderClickURL,
		UnsubscribeURL: placeholderUnsubscribeURL,
		AccentFallback: "#d4a853",
		ServiceRows:    serviceRows,
		Templates:      cards,
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

	return DraftContent{
		Subject:  outreachSubject(restaurantName),
		BodyHTML: htmlBody,
		BodyText: textBuf.String(),
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
