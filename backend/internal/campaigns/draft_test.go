package campaigns_test

import (
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
)

func TestBuildDraftIncludesPlaceholders(t *testing.T) {
	draft := campaigns.BuildDraft(campaigns.DraftInput{
		RestaurantName: "Aurora Cafe",
		DemoWebURL:     "http://localhost:3000",
		DemoSlug:       "aurora-cafe",
		DemoToken:      "demo-token",
	})

	if !strings.Contains(draft.Subject, "Aurora Cafe") {
		t.Fatalf("subject = %q, want restaurant name", draft.Subject)
	}
	if !strings.Contains(draft.BodyHTML, "{{CLICK_URL}}") {
		t.Fatal("body_html missing click placeholder")
	}
	if !strings.Contains(draft.BodyText, "{{UNSUBSCRIBE_URL}}") {
		t.Fatal("body_text missing unsubscribe placeholder")
	}
}

func TestInjectTrackingReplacesPlaceholders(t *testing.T) {
	draft := campaigns.DraftContent{
		Subject:  "Hello",
		BodyHTML: `<a href="{{CLICK_URL}}">Go</a>`,
		BodyText: "Go: {{CLICK_URL}}",
	}

	result := campaigns.InjectTracking(
		draft,
		"http://localhost:8080/t/click/abc",
		"http://localhost:8080/t/unsubscribe/xyz",
		"http://localhost:8080/t/open/open.png",
		true,
	)

	if strings.Contains(result.BodyHTML, "{{CLICK_URL}}") {
		t.Fatal("click placeholder was not replaced in html")
	}
	if !strings.Contains(result.BodyHTML, "t/open/open.png") {
		t.Fatal("open pixel missing from html")
	}
	if strings.Contains(result.BodyText, "{{CLICK_URL}}") {
		t.Fatal("click placeholder was not replaced in text")
	}
}
