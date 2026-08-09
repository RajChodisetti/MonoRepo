package seoreport

import "testing"

func TestBuildPlaceMediaOmitsLinkOnlyCards(t *testing.T) {
	media := buildPlaceMedia(PlaceDetails{
		Website: "https://restaurant.example/menu",
		MapsURI: "https://maps.example/restaurant",
	}, nil)
	if media != nil {
		t.Fatalf("expected no media without a real image source, got %#v", media)
	}
}

func TestBuildPlaceMediaCardsAlwaysHaveImageSource(t *testing.T) {
	media := buildPlaceMedia(PlaceDetails{MapsURI: "https://maps.example/restaurant"}, []placePhoto{
		{Name: "places/example/photos/one"},
		{Name: "places/example/photos/two"},
	})
	if media == nil {
		t.Fatal("expected media for live Places photos")
	}
	for _, card := range append(media.MenuAndHighlights, media.PhotosAndVideos...) {
		if card.PhotoName == "" && card.ImageURL == "" {
			t.Fatalf("media card has no image source: %#v", card)
		}
	}
}
