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
	if draft.Subject != "We built a live demo for Spice Garden — website, AI receptionist & more" {
		t.Fatalf("subject = %q", draft.Subject)
	}
	for _, token := range []string{"Cinematic", "Aurora", "Elysian", "{{CLICK_URL}}", "{{TEMPLATE_3_URL}}"} {
		if !strings.Contains(draft.BodyHTML, token) {
			t.Fatalf("body_html missing %q", token)
		}
	}
	for _, keyword := range []string{"AI Voice Receptionist", "Presentation Websites", "Explore your live demo"} {
		if !strings.Contains(draft.BodyHTML, keyword) {
			t.Fatalf("body_html missing service/cta keyword %q", keyword)
		}
	}
	if !strings.Contains(draft.BodyText, "AI Voice Receptionist") {
		t.Fatal("body_text missing AI Voice Receptionist")
	}
}
