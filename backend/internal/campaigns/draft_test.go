package campaigns_test

import (
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
)

func TestBuildDraftIncludesPlaceholders(t *testing.T) {
	draft := campaigns.BuildDraft(campaigns.DraftInput{
		RestaurantName:      "Aurora Cafe",
		SiteIndex:           2,
		DemoWebURL:          "http://localhost:3000",
		DemoSlug:            "aurora-cafe",
		DemoToken:           "demo-token",
		PresentationSiteURL: "https://tuvisolutions.com/services/restaurants",
	})

	if !strings.Contains(draft.Subject, "Aurora Cafe") {
		t.Fatalf("subject = %q, want restaurant name", draft.Subject)
	}
	if !strings.Contains(draft.Subject, "Quick idea") {
		t.Fatalf("subject = %q, want conversational hook", draft.Subject)
	}
	for _, placeholder := range []string{
		"{{TEMPLATE_1_URL}}",
		"{{TEMPLATE_2_URL}}",
		"{{TEMPLATE_3_URL}}",
		"{{UNSUBSCRIBE_URL}}",
	} {
		if strings.Contains(draft.BodyHTML, placeholder) {
			t.Fatalf("body_html should not include legacy template placeholder %s", placeholder)
		}
	}
	if strings.Contains(draft.BodyText, "{{UNSUBSCRIBE_URL}}") {
		t.Fatal("body_text should not include the removed unsubscribe placeholder")
	}
	for _, name := range []string{"Personalized demo websites", "Services catalog", "Cinematic personalized website", "Aurora personalized website", "Elysian personalized website", "Tuvi restaurant services presentation"} {
		if strings.Contains(draft.BodyHTML, name) {
			t.Fatalf("body_html should not include service-list label %q", name)
		}
	}
	for _, banned := range []string{"We already built"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	for _, banned := range []string{"{{CLICK_URL}}", "Preview:", "Tuvi overview", "No more emails"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	if count := strings.Count(draft.BodyHTML, "https://tuvisolutions.com"); count < 1 {
		t.Fatalf("body_html has %d signature website links, want at least 1", count)
	}
	if count := strings.Count(draft.BodyHTML, "href="); count != 1 {
		t.Fatalf("body_html has %d links, want exactly 1", count)
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

func TestInjectTrackingReplacesPlaceholders(t *testing.T) {
	draft := campaigns.DraftContent{
		Subject:  "Hello",
		BodyHTML: `<a href="{{CLICK_URL}}">Go</a><a href="{{TEMPLATE_2_URL}}">Aurora</a>`,
		BodyText: "Go: {{CLICK_URL}} Aurora: {{TEMPLATE_2_URL}}",
	}

	result := campaigns.InjectTracking(
		draft,
		campaigns.TrackingURLs{
			Click:     "http://localhost:8080/t/click/abc",
			Template1: "http://localhost:8080/t/click/one",
			Template2: "http://localhost:8080/t/click/two",
			Template3: "http://localhost:8080/t/click/three",
			Open:      "http://localhost:8080/t/open/open.png",
		},
		true,
	)

	for _, placeholder := range []string{"{{CLICK_URL}}", "{{TEMPLATE_2_URL}}"} {
		if strings.Contains(result.BodyHTML, placeholder) {
			t.Fatalf("html still contains placeholder %s", placeholder)
		}
	}
	if !strings.Contains(result.BodyHTML, "t/open/open.png") {
		t.Fatal("open pixel missing from html")
	}
	if strings.Contains(result.BodyText, "{{CLICK_URL}}") || strings.Contains(result.BodyText, "{{TEMPLATE_2_URL}}") {
		t.Fatal("text placeholders were not replaced")
	}
}
