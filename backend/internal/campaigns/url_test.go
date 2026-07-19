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
	if draft.Subject != "A live demo for Spice Garden — AI receptionist, website & more" {
		t.Fatalf("subject = %q", draft.Subject)
	}
	for _, token := range []string{
		"{{CLICK_URL}}",
		"{{TEMPLATE_1_URL}}",
		"{{TEMPLATE_2_URL}}",
		"{{TEMPLATE_3_URL}}",
		"{{UNSUBSCRIBE_URL}}",
		"Cinematic personalized website",
		"Aurora personalized website",
		"Elysian personalized website",
		"Tuvi restaurant services presentation",
		"http://localhost:5500",
		"live website preview for Spice Garden",
		"Open Spice Garden demo",
	} {
		if !strings.Contains(draft.BodyHTML, token) {
			t.Fatalf("body_html missing %q", token)
		}
	}
	if strings.Contains(draft.BodyHTML, "%7b") {
		t.Fatal("body_html has URL-escaped CLICK_URL braces")
	}
	for _, banned := range []string{"We already built a preview"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	if !strings.Contains(draft.BodyText, "Tuvi restaurant services presentation") {
		t.Fatal("body_text missing presentation link")
	}
}

func TestRenderOutreachEmailUsesConfiguredPresentationURL(t *testing.T) {
	draft, err := RenderOutreachEmailWithLinks("Spice Garden", OutreachLinkConfig{
		PresentationURL: "https://tuvisolutions.com/services/restaurants",
	})
	if err != nil {
		t.Fatalf("RenderOutreachEmailWithLinks() error = %v", err)
	}
	if !strings.Contains(draft.BodyHTML, "https://tuvisolutions.com/services/restaurants") {
		t.Fatal("body_html missing configured Tuvi presentation URL")
	}
}
