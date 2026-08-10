package seoreport

import "strings"

// buildPlaceMedia assembles Google Maps-style menu/highlights + photos carousels
// exclusively from live Places photo resource names.
func buildPlaceMedia(place PlaceDetails, photos []placePhoto) *PlaceMedia {
	mapsURI := strings.TrimSpace(place.MapsURI)
	menuAndHighlights := make([]MediaCard, 0, 8)
	photosAndVideos := make([]MediaCard, 0, 12)

	photoStart := 0
	if len(photos) > 0 {
		photo := photos[0]
		menuAndHighlights = append(menuAndHighlights, mediaCardForPhoto(
			photo, "photo", "Listing photo", mapsURI, place.Website,
		))
		photoStart = 1
	}

	for i := photoStart; i < len(photos) && len(menuAndHighlights) < 7; i++ {
		photo := photos[i]
		label := "Popular"
		switch {
		case i == photoStart+1:
			label = "Guest favourite"
		case i > photoStart+1:
			label = firstNonEmpty(photo.Attribution, "Highlight")
		}
		menuAndHighlights = append(menuAndHighlights, mediaCardForPhoto(
			photo, "highlight", truncateRunes(label, 42), mapsURI, place.Website,
		))
	}

	if len(photosAndVideos) == 0 && len(photos) > 0 {
		photosAndVideos = append(photosAndVideos, mediaCardForPhoto(
			photos[0], "photo", "All", mapsURI, place.Website,
		))
		if len(photos) > 1 {
			card := mediaCardForPhoto(photos[1], "latest", "Latest", mapsURI, place.Website)
			card.Subtitle = firstNonEmpty(card.Subtitle, "From Google listing")
			photosAndVideos = append(photosAndVideos, card)
		}
		for i := 2; i < len(photos) && len(photosAndVideos) < 8; i++ {
			photosAndVideos = append(photosAndVideos, mediaCardForPhoto(
				photos[i], "photo", firstNonEmpty(photos[i].Attribution, "Photo"), mapsURI, place.Website,
			))
		}
	}

	if len(menuAndHighlights) == 0 && len(photosAndVideos) == 0 {
		return nil
	}

	return &PlaceMedia{
		MenuAndHighlights: menuAndHighlights,
		PhotosAndVideos:   photosAndVideos,
		MapsURI:           mapsURI,
	}
}

func mediaCardForPhoto(photo placePhoto, kind, label, mapsURI, website string) MediaCard {
	return MediaCard{
		Kind:               kind,
		Label:              label,
		Subtitle:           strings.TrimSpace(photo.Attribution),
		PhotoName:          photo.Name,
		Href:               firstNonEmpty(photo.GoogleMapsURI, mapsURI, website),
		AuthorAttributions: append([]AuthorAttribution(nil), photo.AuthorAttributions...),
		GoogleMapsURI:      photo.GoogleMapsURI,
		FlagContentURI:     photo.FlagContentURI,
	}
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if max < 1 || len(r) <= max {
		return string(r)
	}
	return string(r[:max-1]) + "…"
}
