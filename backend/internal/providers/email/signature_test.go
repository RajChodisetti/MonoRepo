package email

import (
	"strings"
	"testing"
)

func TestEnsureTuviSignatureBuildsMultipartContentFromPlainText(t *testing.T) {
	req := EnsureTuviSignature(SendRequest{
		Subject:  "Hello",
		TextBody: "Hi Casey,\n\nThis is the saved template.",
	})

	for _, token := range []string{"Thanks & Regards,", "Team Tuvi", "Tuvi Solutions", tuviWebsiteURL} {
		if !strings.Contains(req.TextBody, token) {
			t.Fatalf("TextBody missing %q", token)
		}
	}
	for _, token := range []string{tuviSignatureMarker, tuviLogoURL, "background:transparent", "www.tuvisolutions.com"} {
		if !strings.Contains(req.HTMLBody, token) {
			t.Fatalf("HTMLBody missing %q", token)
		}
	}
	if !strings.Contains(req.HTMLBody, "This is the saved template.") {
		t.Fatal("HTMLBody does not contain the authored message")
	}
}

func TestEnsureTuviSignatureIsIdempotent(t *testing.T) {
	once := EnsureTuviSignature(SendRequest{TextBody: "Hello"})
	twice := EnsureTuviSignature(once)

	if once.TextBody != twice.TextBody || once.HTMLBody != twice.HTMLBody {
		t.Fatal("EnsureTuviSignature changed content on the second call")
	}
	if strings.Count(twice.HTMLBody, tuviSignatureMarker) != 1 {
		t.Fatal("HTML signature was duplicated")
	}
}

func TestEnsureTuviSignatureEscapesPlainTextWhenBuildingHTML(t *testing.T) {
	req := EnsureTuviSignature(SendRequest{TextBody: "Hello <script>alert('x')</script>"})
	if strings.Contains(req.HTMLBody, "<script>") || !strings.Contains(req.HTMLBody, "&lt;script&gt;") {
		t.Fatalf("HTMLBody did not escape authored plain text: %q", req.HTMLBody)
	}
}
