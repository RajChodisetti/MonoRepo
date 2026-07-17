package places

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestListPhotoURLsRefreshesResourcesAndReturnsAttribution(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("X-Goog-Api-Key") != "server-secret" {
			t.Fatalf("missing server-side API key header")
		}
		switch r.URL.Path {
		case "/places/place-1":
			if r.Header.Get("X-Goog-FieldMask") != "photos" {
				t.Fatalf("field mask = %q, want photos", r.Header.Get("X-Goog-FieldMask"))
			}
			_, _ = w.Write([]byte(`{"photos":[{"name":"places/place-1/photos/photo-1","widthPx":1200,"heightPx":800,"authorAttributions":[{"displayName":"Owner","uri":"//maps.google.com/contrib/1"}]}]}`))
		case "/places/place-1/photos/photo-1/media":
			if r.URL.Query().Get("skipHttpRedirect") != "true" || r.URL.Query().Get("maxWidthPx") != "1600" {
				t.Fatalf("unexpected media query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"photoUri":"https://images.example.test/photo-1.jpg"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(config.PlacesConfig{
		APIKey:        "server-secret",
		APIBaseURL:    server.URL,
		PhotoLimit:    10,
		PhotoMaxWidth: 1600,
		Timeout:       time.Second,
	})
	photos, err := client.ListPhotoURLs(context.Background(), "place-1")
	if err != nil {
		t.Fatalf("ListPhotoURLs() error = %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
	if len(photos) != 1 || photos[0].URL != "https://images.example.test/photo-1.jpg" {
		t.Fatalf("photos = %#v", photos)
	}
	if len(photos[0].AuthorAttributions) != 1 || photos[0].AuthorAttributions[0].URI != "https://maps.google.com/contrib/1" {
		t.Fatalf("attributions = %#v", photos[0].AuthorAttributions)
	}
}

func TestListPhotoURLsDoesNotExposeProviderResponseOrKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"server-secret owner@example.test"}`))
	}))
	defer server.Close()

	client := NewClient(config.PlacesConfig{
		APIKey:        "server-secret",
		APIBaseURL:    server.URL,
		PhotoLimit:    10,
		PhotoMaxWidth: 1600,
		Timeout:       time.Second,
	})
	_, err := client.ListPhotoURLs(context.Background(), "place-1")
	if err == nil {
		t.Fatal("ListPhotoURLs() error = nil")
	}
	message := err.Error()
	for _, secret := range []string{"server-secret", "owner@example.test"} {
		if strings.Contains(message, secret) {
			t.Fatalf("error leaked %q: %s", secret, message)
		}
	}
	if !strings.Contains(message, strconv.Itoa(http.StatusForbidden)) {
		t.Fatalf("error = %q, want status code", message)
	}
}

func TestListPhotoURLsRequiresConfiguration(t *testing.T) {
	client := NewClient(config.PlacesConfig{PhotoLimit: 10, PhotoMaxWidth: 1600, Timeout: time.Second})
	_, err := client.ListPhotoURLs(context.Background(), "place-1")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("error = %v, want ErrNotConfigured", err)
	}
}
