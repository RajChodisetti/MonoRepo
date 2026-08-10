package seoreport

import "testing"

func TestBuildPlaceMediaPreservesPhotoAttributionAndSourceLinks(t *testing.T) {
	place := PlaceDetails{
		MapsURI: "https://maps.example/listing",
		Website: "https://restaurant.example",
	}
	photos := []placePhoto{
		{
			Name: "photos/owner", Attribution: "Restaurant owner, Photographer", GoogleMapsURI: "https://maps.example/photo/owner",
			FlagContentURI: "https://maps.example/photo/owner/flag",
			AuthorAttributions: []AuthorAttribution{
				{DisplayName: "Restaurant owner", URI: "https://maps.example/contrib/owner", PhotoURI: "https://images.example/owner"},
				{DisplayName: "Photographer", URI: "https://maps.example/contrib/photographer", PhotoURI: "https://images.example/photographer"},
			},
		},
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
			if card.GoogleMapsURI != photo.GoogleMapsURI || card.FlagContentURI != photo.FlagContentURI {
				t.Errorf("card %q source links=%q/%q, want %q/%q", card.PhotoName, card.GoogleMapsURI, card.FlagContentURI, photo.GoogleMapsURI, photo.FlagContentURI)
			}
			if len(card.AuthorAttributions) != len(photo.AuthorAttributions) {
				t.Errorf("card %q author attributions=%#v, want %#v", card.PhotoName, card.AuthorAttributions, photo.AuthorAttributions)
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
	if len(first.AuthorAttributions) != 2 || first.AuthorAttributions[1].URI != "https://maps.example/contrib/photographer" {
		t.Fatalf("all photo authors were not retained: %#v", first.AuthorAttributions)
	}
	if first.Href != first.GoogleMapsURI {
		t.Fatalf("photo-specific source %q was replaced by %q", first.GoogleMapsURI, first.Href)
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

func TestBuildPlaceMediaDoesNotInventMenuWithoutPhotoEvidence(t *testing.T) {
	media := buildPlaceMedia(PlaceDetails{
		MapsURI: "https://maps.example/listing",
		Website: "https://restaurant.example",
	}, nil)
	if media != nil {
		t.Fatalf("media=%#v, want nil when no photo evidence was returned", media)
	}
}
