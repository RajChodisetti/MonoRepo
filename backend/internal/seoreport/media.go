package seoreport

import "strings"

// buildPlaceMedia assembles Google Maps-style menu/highlights + photos carousels.
func buildPlaceMedia(place PlaceDetails, photos []placePhoto, content *profilesSiteMedia) *PlaceMedia {
	mapsURI := strings.TrimSpace(place.MapsURI)
	menuAndHighlights := make([]MediaCard, 0, 8)
	photosAndVideos := make([]MediaCard, 0, 12)

	if content != nil {
		for _, item := range content.MenuItems {
			img := strings.TrimSpace(item.ImageURL)
			name := strings.TrimSpace(item.Name)
			if img == "" || name == "" {
				continue
			}
			if len(menuAndHighlights) == 0 {
				menuAndHighlights = append(menuAndHighlights, MediaCard{
					Kind:     "menu",
					Label:    "Menu",
					ImageURL: img,
					Href:     firstNonEmpty(place.Website, mapsURI),
				})
			}
			menuAndHighlights = append(menuAndHighlights, MediaCard{
				Kind:     "highlight",
				Label:    truncateRunes(name, 42),
				ImageURL: img,
				Href:     mapsURI,
			})
			if len(menuAndHighlights) >= 7 {
				break
			}
		}
		for _, g := range content.Gallery {
			img := firstNonEmpty(strings.TrimSpace(g.URL), strings.TrimSpace(g.ThumbnailURL))
			if img == "" {
				continue
			}
			label := firstNonEmpty(strings.TrimSpace(g.Title), "Photo")
			photosAndVideos = append(photosAndVideos, MediaCard{
				Kind:     "photo",
				Label:    label,
				ImageURL: img,
				Href:     mapsURI,
			})
			if len(photosAndVideos) >= 10 {
				break
			}
		}
	}

  hasMenuCard := len(menuAndHighlights) > 0 && menuAndHighlights[0].Kind == "menu"
  photoStart := 0
  if !hasMenuCard {
    if len(photos) > 0 {
      menuAndHighlights = append([]MediaCard{{
        Kind:      "menu",
        Label:     "Menu",
        PhotoName: photos[0].Name,
        Href:      firstNonEmpty(place.Website, mapsURI),
      }}, menuAndHighlights...)
      photoStart = 1
    } else if mapsURI != "" || place.Website != "" {
      menuAndHighlights = append([]MediaCard{{
        Kind:  "menu",
        Label: "Menu",
        Href:  firstNonEmpty(place.Website, mapsURI),
      }}, menuAndHighlights...)
    }
  }

  // Fill highlights from Places photos when inventory dishes are missing.
  for i := photoStart; i < len(photos) && len(menuAndHighlights) < 7; i++ {
    inventoryHighlights := 0
    for _, c := range menuAndHighlights {
      if c.Kind == "highlight" && c.ImageURL != "" {
        inventoryHighlights++
      }
    }
    if inventoryHighlights >= 3 {
      break
    }
    p := photos[i]
    label := "Popular"
    switch {
    case i == photoStart+1:
      label = "Guest favourite"
    case i > photoStart+1:
      label = firstNonEmpty(p.Attribution, "Highlight")
    }
    menuAndHighlights = append(menuAndHighlights, MediaCard{
      Kind:      "highlight",
      Label:     truncateRunes(label, 42),
      PhotoName: p.Name,
      Href:      firstNonEmpty(p.GoogleMapsURI, mapsURI),
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
	} else if len(photosAndVideos) > 0 {
		photosAndVideos[0].Kind = "photo"
		photosAndVideos[0].Label = "All"
		if len(photosAndVideos) > 1 {
			photosAndVideos[1].Kind = "latest"
			if photosAndVideos[1].Label == "Photo" || photosAndVideos[1].Label == "" {
				photosAndVideos[1].Label = "Latest"
			}
			photosAndVideos[1].Subtitle = firstNonEmpty(photosAndVideos[1].Subtitle, "Recent upload")
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

// profilesSiteMedia is a thin view of inventory media for buildPlaceMedia.
type profilesSiteMedia struct {
	MenuItems []siteMenuMedia
	Gallery   []siteGalleryMedia
}

type siteMenuMedia struct {
	Name     string
	ImageURL string
}

type siteGalleryMedia struct {
	Title        string
	URL          string
	ThumbnailURL string
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if max < 1 || len(r) <= max {
		return string(r)
	}
	return string(r[:max-1]) + "…"
}
