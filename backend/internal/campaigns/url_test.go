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
		"{{UNSUBSCRIBE_URL}}",
		"AI Voice Receptionist",
		"Presentation Websites",
		"Online Reservations",
		"Custom Apps",
		"http://localhost:5500",
		"http://localhost:3001",
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
	// Long template-card section removed
	for _, banned := range []string{"We already built a preview", "Cinematic", "Aurora", "Elysian"} {
		if strings.Contains(draft.BodyHTML, banned) {
			t.Fatalf("body_html should not contain %q", banned)
		}
	}
	if !strings.Contains(draft.BodyText, "AI Voice Receptionist") {
		t.Fatal("body_text missing AI Voice Receptionist")
	}
}
