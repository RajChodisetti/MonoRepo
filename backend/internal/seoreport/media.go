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
		menuAndHighlights = append(menuAndHighlights, MediaCard{
			Kind:      "menu",
			Label:     "Menu",
			PhotoName: photos[0].Name,
			Href:      firstNonEmpty(place.Website, mapsURI),
		})
		photoStart = 1
	} else if mapsURI != "" || place.Website != "" {
		menuAndHighlights = append(menuAndHighlights, MediaCard{
			Kind:  "menu",
			Label: "Menu",
			Href:  firstNonEmpty(place.Website, mapsURI),
		})
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
		menuAndHighlights = append(menuAndHighlights, MediaCard{
			Kind:      "highlight",
			Label:     truncateRunes(label, 42),
			PhotoName: photo.Name,
			Href:      firstNonEmpty(photo.GoogleMapsURI, mapsURI),
		})
	}

	if len(photosAndVideos) == 0 && len(photos) > 0 {
		photosAndVideos = append(photosAndVideos, MediaCard{
			Kind:      "photo",
			Label:     "All",
			PhotoName: photos[0].Name,
			Href:      mapsURI,
		})
		if len(photos) > 1 {
			photosAndVideos = append(photosAndVideos, MediaCard{
				Kind:      "latest",
				Label:     "Latest",
				Subtitle:  "From Google listing",
				PhotoName: photos[1].Name,
				Href:      mapsURI,
			})
		}
		for i := 2; i < len(photos) && len(photosAndVideos) < 8; i++ {
			photosAndVideos = append(photosAndVideos, MediaCard{
				Kind:      "photo",
				Label:     firstNonEmpty(photos[i].Attribution, "Photo"),
				PhotoName: photos[i].Name,
				Href:      firstNonEmpty(photos[i].GoogleMapsURI, mapsURI),
			})
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

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if max < 1 || len(r) <= max {
		return string(r)
	}
	return string(r[:max-1]) + "…"
}
