package profiles

import (
	"encoding/json"
	"testing"
)

func TestSanitizePublicSiteContentRemovesLegacyScrapedImages(t *testing.T) {
	content := SiteContent{
		Thumbnail:     "https://legacy.example/thumbnail.jpg",
		GalleryImages: []GalleryImage{{URL: "https://legacy.example/gallery.jpg"}},
		MenuItems: []SiteMenuItem{{
			Name: "Dish", ImageURL: "https://legacy.example/dish.jpg",
			Images: json.RawMessage(`[{"url":"https://legacy.example/dish-2.jpg"}]`),
		}},
	}

	sanitized := SanitizePublicSiteContent(content)
	if sanitized.Thumbnail != "" || len(sanitized.GalleryImages) != 0 {
		t.Fatalf("sanitized = %#v, want no profile/gallery images", sanitized)
	}
	if sanitized.MenuItems[0].ImageURL != "" || len(sanitized.MenuItems[0].Images) != 0 {
		t.Fatalf("menu item = %#v, want no legacy image fields", sanitized.MenuItems[0])
	}
}
