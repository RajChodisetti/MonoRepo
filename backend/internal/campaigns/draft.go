package campaigns

import (
	"fmt"
	"net/url"
	"strings"
)

type DraftContent struct {
	Subject  string
	BodyHTML string
	BodyText string
}

type DraftInput struct {
	RestaurantName string
	SiteIndex      int
	DemoWebURL     string
	DemoSlug       string
	DemoToken      string
}

type TrackingURLs struct {
	Click       string
	Template1   string
	Template2   string
	Template3   string
	Open        string
	Unsubscribe string
}

func BuildDraft(input DraftInput) DraftContent {
	draft, err := RenderOutreachEmail(input.RestaurantName)
	if err != nil {
		name := strings.TrimSpace(input.RestaurantName)
		if name == "" {
			name = "your restaurant"
		}
		return DraftContent{
			Subject:  outreachSubject(name),
			BodyHTML: fmt.Sprintf("<p>A live demo for %s. View: {{CLICK_URL}}</p>", name),
			BodyText: fmt.Sprintf("View your demo: {{CLICK_URL}}\nUnsubscribe: {{UNSUBSCRIBE_URL}}\n"),
		}
	}
	return draft
}

func buildDemoURL(webBase, slug, demoToken string) string {
	base := strings.TrimRight(strings.TrimSpace(webBase), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	values := url.Values{}
	if demoToken != "" {
		values.Set("token", demoToken)
	}
	query := values.Encode()
	path := fmt.Sprintf("%s/demo/%s", base, strings.TrimSpace(slug))
	if query == "" {
		return path
	}
	return path + "?" + query
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
		placeholderUnsubscribeURL, urls.Unsubscribe,
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
