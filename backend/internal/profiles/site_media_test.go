package profiles

import "testing"

func TestTemporaryGoogleMediaURL(t *testing.T) {
	for _, value := range []string{
		"https://lh3.googleusercontent.com/photo",
		"https://images.ggpht.com/photo",
	} {
		if !isTemporaryGoogleMediaURL(value) {
			t.Fatalf("isTemporaryGoogleMediaURL(%q) = false", value)
		}
	}
	if isTemporaryGoogleMediaURL("https://cdn.example.test/photo.jpg") {
		t.Fatal("durable CDN URL classified as temporary Google media")
	}
}
