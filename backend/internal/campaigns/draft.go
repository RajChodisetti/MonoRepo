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
	DemoWebURL     string
	DemoSlug       string
	DemoToken      string
}

func BuildDraft(input DraftInput) DraftContent {
	name := strings.TrimSpace(input.RestaurantName)
	if name == "" {
		name = "your restaurant"
	}

	demoURL := buildDemoURL(input.DemoWebURL, input.DemoSlug, input.DemoToken)

	subject := fmt.Sprintf("A modern website preview for %s", name)
	textBody := fmt.Sprintf(`Hi there,

We built a personalized demo website for %s and would love your feedback.

View your demo: {{CLICK_URL}}

If you would rather not receive these emails, unsubscribe here: {{UNSUBSCRIBE_URL}}

— Tuvi Solutions
`, name)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.5; color: #222;">
  <p>Hi there,</p>
  <p>We built a personalized demo website for <strong>%s</strong> and would love your feedback.</p>
  <p><a href="{{CLICK_URL}}" style="display:inline-block;padding:12px 20px;background:#111;color:#fff;text-decoration:none;border-radius:6px;">View your demo</a></p>
  <p style="font-size:12px;color:#666;">Demo link (copy/paste): %s</p>
  <p style="font-size:12px;color:#888;">If you would rather not receive these emails, <a href="{{UNSUBSCRIBE_URL}}">unsubscribe here</a>.</p>
  <p>— Tuvi Solutions</p>
</body>
</html>`, name, demoURL)

	return DraftContent{
		Subject:  subject,
		BodyHTML: htmlBody,
		BodyText: textBody,
	}
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

func InjectTracking(content DraftContent, clickURL, unsubscribeURL, openPixelURL string, openTracking bool) DraftContent {
	replacer := strings.NewReplacer(
		"{{CLICK_URL}}", clickURL,
		"{{UNSUBSCRIBE_URL}}", unsubscribeURL,
	)
	html := replacer.Replace(content.BodyHTML)
	if openTracking && openPixelURL != "" {
		html += fmt.Sprintf(`<img src="%s" width="1" height="1" alt="" style="display:none;" />`, openPixelURL)
	}
	return DraftContent{
		Subject:  content.Subject,
		BodyHTML: html,
		BodyText: replacer.Replace(content.BodyText),
	}
}
