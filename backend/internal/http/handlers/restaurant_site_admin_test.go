package handlers

import "testing"

func TestGeneratedSiteURL(t *testing.T) {
	got := generatedSiteURL("https://demo.example.test", 42, "3")
	want := "https://demo.example.test?id=42&template=3"
	if got != want {
		t.Fatalf("generatedSiteURL() = %q, want %q", got, want)
	}
}
