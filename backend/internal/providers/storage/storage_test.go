package storage

import "testing"

func TestPublicURLPreservesBasePathAndEscapesKey(t *testing.T) {
	store := &S3{publicBaseURL: "https://cdn.example.test/assets"}
	got := store.PublicURL("restaurants/one/dining room.jpg")
	want := "https://cdn.example.test/assets/restaurants/one/dining%20room.jpg"
	if got != want {
		t.Fatalf("PublicURL() = %q, want %q", got, want)
	}
}

func TestPublicURLRejectsTraversal(t *testing.T) {
	store := &S3{publicBaseURL: "https://cdn.example.test"}
	if got := store.PublicURL("restaurants/../secret.jpg"); got != "" {
		t.Fatalf("PublicURL() = %q, want empty", got)
	}
}
