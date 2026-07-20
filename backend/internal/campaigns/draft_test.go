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
	if !strings.Contains(draft.Subject, "live demo") {
		t.Fatalf("subject = %q, want sales hook", draft.Subject)
	}
	for _, placeholder := range []string{
		"{{CLICK_URL}}",
		"{{UNSUBSCRIBE_URL}}",
	} {
		if !strings.Contains(draft.BodyHTML, placeholder) {
			t.Fatalf("body_html missing placeholder %s", placeholder)
		}
	}
	for _, placeholder := range []string{
		"{{TEMPLATE_1_URL}}",
		"{{TEMPLATE_2_URL}}",
		"{{TEMPLATE_3_URL}}",
	} {
		if strings.Contains(draft.BodyHTML, placeholder) {
			t.Fatalf("body_html should not include legacy template placeholder %s", placeholder)
		}
	}
	if !strings.Contains(draft.BodyText, "{{UNSUBSCRIBE_URL}}") {
		t.Fatal("body_text missing unsubscribe placeholder")
	}
	for _, name := range []string{"Personalized demo websites", "Services catalog"} {
		if !strings.Contains(draft.BodyHTML, name) {
			t.Fatalf("body_html missing service %q", name)
		}
	}
	for _, name := range []string{"Cinematic personalized website", "Aurora personalized website", "Elysian personalized website", "Tuvi restaurant services presentation"} {
		if strings.Contains(draft.BodyHTML, name) {
			t.Fatalf("body_html should not include old service %q", name)
		}
	}
	for _, banned := range []string{"We already built"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	if !strings.Contains(draft.BodyHTML, "https://tuvisolutions.com/services/restaurants") {
		t.Fatal("body_html missing Services catalog URL")
	}
	if count := strings.Count(draft.BodyHTML, "{{CLICK_URL}}"); count != 1 {
		t.Fatalf("body_html has %d personalized demo links, want 1", count)
	}
	if count := strings.Count(draft.BodyHTML, "https://tuvisolutions.com/services/restaurants"); count != 1 {
		t.Fatalf("body_html has %d Services catalog links, want 1", count)
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
			Click:       "http://localhost:8080/t/click/abc",
			Template1:   "http://localhost:8080/t/click/one",
			Template2:   "http://localhost:8080/t/click/two",
			Template3:   "http://localhost:8080/t/click/three",
			Unsubscribe: "http://localhost:8080/t/unsubscribe/xyz",
			Open:        "http://localhost:8080/t/open/open.png",
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
