package campaigns

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

type DraftContent struct {
	Subject  string
	BodyHTML string
	BodyText string
}

type DraftInput struct {
	RestaurantName      string
	SiteIndex           int
	DemoWebURL          string
	DemoSlug            string
	DemoToken           string
	PresentationSiteURL string
	MarketingSiteURL    string
}

type TrackingURLs struct {
	Click     string
	Template1 string
	Template2 string
	Template3 string
	Open      string
}

func BuildDraft(input DraftInput) DraftContent {
	draft, err := RenderOutreachEmailWithLinks(input.RestaurantName, OutreachLinkConfig{
		PresentationURL: input.PresentationSiteURL,
		MarketingURL:    input.MarketingSiteURL,
	})
	if err != nil {
		name := strings.TrimSpace(input.RestaurantName)
		if name == "" {
			name = "your restaurant"
		}
		return EnsureOutreachSignature(DraftContent{
			Subject: outreachSubject(name),
			BodyHTML: fmt.Sprintf(
				`<p>A live demo for %s. <a href="{{CLICK_URL}}">View the demo</a>.</p>`,
				html.EscapeString(name),
			),
			BodyText: fmt.Sprintf("View your demo: {{CLICK_URL}}\n"),
		})
	}
	return EnsureOutreachSignature(draft)
}

func buildDemoURL(webBase, slug, demoToken string) string {
	base := strings.TrimRight(strings.TrimSpace(webBase), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	values := url.Values{}
	values.Set("slug", strings.TrimSpace(slug))
	if demoToken != "" {
		values.Set("token", demoToken)
	}
	return base + "/?" + values.Encode()
}

func buildTokenGatedDemoPreviewURL(webBase, slug, demoToken, templateID string) string {
	baseURL := buildDemoURL(webBase, slug, demoToken)
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	values := parsed.Query()
	values.Set("template", strings.TrimSpace(templateID))
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func buildTemplatePreviewURL(webBase string, siteIndex int, templateID, demoToken string) string {
	base := strings.TrimRight(strings.TrimSpace(webBase), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	values := url.Values{}
	values.Set("id", fmt.Sprintf("%d", siteIndex))
	values.Set("template", strings.TrimSpace(templateID))
	if demoToken != "" {
		values.Set("token", demoToken)
	}
	return base + "/?" + values.Encode()
}

func InjectTracking(content DraftContent, urls TrackingURLs, openTracking bool) DraftContent {
	replacer := strings.NewReplacer(
		placeholderClickURL, urls.Click,
		placeholderTemplate1URL, urls.Template1,
		placeholderTemplate2URL, urls.Template2,
		placeholderTemplate3URL, urls.Template3,
	)
	html := replacer.Replace(content.BodyHTML)
	if openTracking && urls.Open != "" {
		html += fmt.Sprintf(`<img src="%s" width="1" height="1" alt="" style="display:none;" />`, urls.Open)
	}
	return DraftContent{
		Subject:  content.Subject,
		BodyHTML: html,
		BodyText: replacer.Replace(content.BodyText),
	}
}
