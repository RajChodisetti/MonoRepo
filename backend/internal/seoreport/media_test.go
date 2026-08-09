package seoreport

import "testing"

func TestBuildPlaceMediaPreservesPhotoAttributionAndSourceLinks(t *testing.T) {
	place := PlaceDetails{
		MapsURI: "https://maps.example/listing",
		Website: "https://restaurant.example",
	}
	photos := []placePhoto{
		{Name: "photos/owner", Attribution: "Restaurant owner", GoogleMapsURI: "https://maps.example/photo/owner"},
		{Name: "photos/guest", Attribution: "Dining guest", GoogleMapsURI: "https://maps.example/photo/guest"},
		{Name: "photos/local", Attribution: "Local guide", GoogleMapsURI: "https://maps.example/photo/local"},
		{Name: "photos/listing", Attribution: "Listing contributor"},
	}

	media := buildPlaceMedia(place, photos)
	if media == nil {
		t.Fatal("buildPlaceMedia returned nil")
	}

	wantByPhoto := make(map[string]placePhoto, len(photos))
	for _, photo := range photos {
		wantByPhoto[photo.Name] = photo
	}

	cardCount := 0
	for _, cards := range [][]MediaCard{media.MenuAndHighlights, media.PhotosAndVideos} {
		for _, card := range cards {
			if card.PhotoName == "" {
				continue
			}
			cardCount++
			photo, ok := wantByPhoto[card.PhotoName]
			if !ok {
				t.Fatalf("unexpected photo card %q", card.PhotoName)
			}
			if card.Subtitle != photo.Attribution {
				t.Errorf("card %q subtitle=%q, want contributor %q", card.PhotoName, card.Subtitle, photo.Attribution)
			}
			wantHref := firstNonEmpty(photo.GoogleMapsURI, place.MapsURI, place.Website)
			if card.Href != wantHref {
				t.Errorf("card %q href=%q, want %q", card.PhotoName, card.Href, wantHref)
			}
			if card.Kind == "menu" {
				t.Errorf("card %q is an unverified menu classification", card.PhotoName)
			}
		}
	}
	if cardCount == 0 {
		t.Fatal("expected photo-backed media cards")
	}

	first := media.MenuAndHighlights[0]
	if first.Kind != "photo" || first.Label != "Listing photo" {
		t.Fatalf("first card=(kind %q, label %q), want fail-closed listing photo", first.Kind, first.Label)
	}
}

func TestBuildPlaceMediaFallsBackFromPhotoLinkToListingThenWebsite(t *testing.T) {
	photo := placePhoto{Name: "photos/one", Attribution: "Contributor"}

	withListing := buildPlaceMedia(PlaceDetails{
		MapsURI: "https://maps.example/listing",
		Website: "https://restaurant.example",
	}, []placePhoto{photo})
	if got := withListing.MenuAndHighlights[0].Href; got != "https://maps.example/listing" {
		t.Fatalf("listing fallback href=%q", got)
	}

	withWebsite := buildPlaceMedia(PlaceDetails{
		Website: "https://restaurant.example",
	}, []placePhoto{photo})
	if got := withWebsite.MenuAndHighlights[0].Href; got != "https://restaurant.example" {
		t.Fatalf("website fallback href=%q", got)
	}
}
