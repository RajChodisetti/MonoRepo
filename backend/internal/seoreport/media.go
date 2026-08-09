package seoreport

import "strings"

// buildPlaceMedia assembles neutral listing-photo evidence exclusively from
// live Places photo resource names. The Places response does not identify a
// photo as a menu, favourite, popular, or latest image, so this boundary must
// not manufacture those labels.
func buildPlaceMedia(place PlaceDetails, photos []placePhoto) *PlaceMedia {
	mapsURI := strings.TrimSpace(place.MapsURI)
	photosAndVideos := make([]MediaCard, 0, 10)

	for i := 0; i < len(photos) && len(photosAndVideos) < 10; i++ {
		photo := photos[i]
		photosAndVideos = append(photosAndVideos, MediaCard{
			Kind:      "photo",
			Label:     truncateRunes(firstNonEmpty(photo.Attribution, "Listing photo"), 42),
			PhotoName: photo.Name,
			Href:      firstNonEmpty(photo.GoogleMapsURI, mapsURI),
		})
	}

	if len(photosAndVideos) == 0 {
		return nil
	}

	return &PlaceMedia{
		PhotosAndVideos: photosAndVideos,
		MapsURI:         mapsURI,
	}
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if max < 1 || len(r) <= max {
		return string(r)
	}
	return string(r[:max-1]) + "…"
}
