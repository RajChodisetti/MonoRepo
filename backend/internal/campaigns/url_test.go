package campaigns

import (
	"strings"
	"testing"
)

func TestBuildTemplatePreviewURL(t *testing.T) {
	got := buildTemplatePreviewURL("http://localhost:3000", 4, "2", "demo-token")
	want := "http://localhost:3000/?id=4&template=2&token=demo-token"
	if got != want {
		t.Fatalf("buildTemplatePreviewURL() = %q, want %q", got, want)
	}

	gotNoToken := buildTemplatePreviewURL("http://localhost:3000/", 0, "1", "")
	wantNoToken := "http://localhost:3000/?id=0&template=1"
	if gotNoToken != wantNoToken {
		t.Fatalf("buildTemplatePreviewURL() without token = %q, want %q", gotNoToken, wantNoToken)
	}
}

func TestRenderOutreachEmail(t *testing.T) {
	draft, err := RenderOutreachEmail("Spice Garden")
	if err != nil {
		t.Fatalf("RenderOutreachEmail() error = %v", err)
	}
	if draft.Subject != "Quick idea for Spice Garden" {
		t.Fatalf("subject = %q", draft.Subject)
	}
	for _, token := range []string{
		"{{UNSUBSCRIBE_URL}}",
		"quick preview",
		"not a long pitch",
		"Would it make sense",
		"Unsubscribe",
	} {
		if !strings.Contains(draft.BodyHTML, token) {
			t.Fatalf("body_html missing %q", token)
		}
	}
	for _, token := range []string{
		"{{TEMPLATE_1_URL}}",
		"{{TEMPLATE_2_URL}}",
		"{{TEMPLATE_3_URL}}",
		"http://localhost:5500",
		"Personalized demo websites",
		"Services catalog",
		"Cinematic personalized website",
		"Aurora personalized website",
		"Elysian personalized website",
		"Tuvi restaurant services presentation",
		"Open Spice Garden demo",
		"Business outreach from Tuvi Solutions",
		"Opt out:",
		"No more emails",
		"Tuvi overview",
		"Preview:",
		"Best,",
	} {
		if strings.Contains(draft.BodyHTML, token) {
			t.Fatalf("body_html should not include %q", token)
		}
	}
	if strings.Contains(draft.BodyHTML, "%7b") {
		t.Fatal("body_html has URL-escaped CLICK_URL braces")
	}
	if strings.Contains(draft.BodyHTML, "{{CLICK_URL}}") {
		t.Fatal("body_html should not contain personalized demo placeholder")
	}
	if count := strings.Count(draft.BodyHTML, "href="); count != 2 {
		t.Fatalf("body_html has %d links, want exactly 2", count)
	}
	for _, banned := range []string{"We already built a preview"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	if !strings.Contains(draft.BodyText, "quick preview") {
		t.Fatal("body_text missing conversational body")
	}
	for _, token := range []string{"tuvi-solutions-logo-transparent.png", "Team Tuvi", "Thanks &amp; Regards,", "tuvisolutions.com"} {
		if !strings.Contains(draft.BodyHTML, token) {
			t.Fatalf("body_html missing signature token %q", token)
		}
	}
	for _, token := range []string{"Thanks & Regards,", "Team Tuvi", "Tuvi Solutions", "https://tuvisolutions.com"} {
		if !strings.Contains(draft.BodyText, token) {
			t.Fatalf("body_text missing signature token %q", token)
		}
	}
}

func TestRenderOutreachEmailUsesConfiguredPresentationURL(t *testing.T) {
	draft, err := RenderOutreachEmailWithLinks("Spice Garden", OutreachLinkConfig{
		PresentationURL: "https://tuvisolutions.com/services/restaurants",
	})
	if err != nil {
		t.Fatalf("RenderOutreachEmailWithLinks() error = %v", err)
	}
	if strings.Contains(draft.BodyHTML, "https://tuvisolutions.com/services/restaurants") {
		t.Fatal("body_html should not use presentation URL in the conversational base email")
	}
}
