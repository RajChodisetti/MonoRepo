package handlers

import "testing"

func TestGeneratedSiteURL(t *testing.T) {
	got := generatedSiteURL("https://demo.example.test", "d419296d-17f6-4bc8-b522-943933451d74", 42, "3")
	want := "https://demo.example.test?id=42&restaurant_id=d419296d-17f6-4bc8-b522-943933451d74&template=3"
	if got != want {
		t.Fatalf("generatedSiteURL() = %q, want %q", got, want)
	}
}
