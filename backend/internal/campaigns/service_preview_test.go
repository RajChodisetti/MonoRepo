package campaigns

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

func TestListDemoLinksUsesPreferredElysianPreview(t *testing.T) {
	restaurantID := uuid.New()
	demoSiteID := uuid.New()
	campaignID := uuid.New()
	repo := &Mock{Campaigns: map[uuid.UUID]Campaign{
		campaignID: {
			ID:           campaignID,
			RestaurantID: restaurantID,
			DemoSiteID:   demoSiteID,
			DemoToken:    "opaque-demo-token",
			CreatedAt:    time.Now().UTC(),
		},
	}}
	demoRepo := &demos.Mock{Sites: map[string]demos.Site{
		"sample-cafe": {
			ID:           demoSiteID,
			RestaurantID: restaurantID,
			Slug:         "sample-cafe",
			Status:       demos.StatusPublished,
		},
	}}
	service := NewService(
		repo,
		demoRepo,
		restaurants.NewService(nil, nil),
		nil,
		config.AppURLsConfig{PublicWebURL: "https://demo.example.test"},
	)

	links, err := service.ListDemoLinks(context.Background(), auth.Principal{
		UserID: uuid.New(),
		Role:   auth.RoleInternalAdmin,
	}, restaurantID)
	if err != nil {
		t.Fatalf("ListDemoLinks() error = %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("ListDemoLinks() returned %d links, want 1", len(links))
	}

	preview, err := url.Parse(links[0].PreviewURL)
	if err != nil {
		t.Fatalf("parse preview URL: %v", err)
	}
	if got := preview.Query().Get("template"); got != preferredPreviewTemplateID {
		t.Fatalf("preview template = %q, want %q", got, preferredPreviewTemplateID)
	}
	if got := preview.Query().Get("slug"); got != "sample-cafe" {
		t.Fatalf("preview slug = %q, want sample-cafe", got)
	}
	if got := preview.Query().Get("token"); got != "opaque-demo-token" {
		t.Fatalf("preview token = %q, want opaque demo token", got)
	}
}
