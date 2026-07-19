package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"

	placesprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/places"
)

type PlaceIDRepository interface {
	GetGooglePlaceID(ctx context.Context, restaurantID uuid.UUID) (string, error)
}

type Service struct {
	repository Repository
	profiles   PlaceIDRepository
	photos     placesprovider.PhotoResolver
	objects    ObjectStore
	log        *slog.Logger
	photoCalls singleflight.Group
}

func NewService(
	repository Repository,
	profiles PlaceIDRepository,
	photos placesprovider.PhotoResolver,
	objects ObjectStore,
	log *slog.Logger,
) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		repository: repository,
		profiles:   profiles,
		photos:     photos,
		objects:    objects,
		log:        log,
	}
}

func (service *Service) PublicForRestaurant(
	ctx context.Context,
	restaurantID uuid.UUID,
	limit int,
) []PublicMedia {
	if service == nil || service.repository == nil {
		return []PublicMedia{}
	}
	if limit < 1 {
		limit = 6
	}

	items := service.publicOwnedAssets(ctx, restaurantID, limit)
	if len(items) >= limit || service.profiles == nil || service.photos == nil {
		return items[:min(len(items), limit)]
	}

	placeID, err := service.profiles.GetGooglePlaceID(ctx, restaurantID)
	if err != nil || strings.TrimSpace(placeID) == "" {
		return items
	}
	hints, err := service.repository.ListClassificationHints(ctx, restaurantID)
	if err != nil {
		service.log.WarnContext(ctx, "media_classification_hints_unavailable", "restaurant_id", restaurantID, "error", err)
		hints = nil
	}
	hintsByFingerprint := make(map[string]ClassificationHint, len(hints))
	for _, hint := range hints {
		if hint.SourceFingerprint != "" {
			hintsByFingerprint[hint.SourceFingerprint] = hint
		}
	}

	// Ask for the provider maximum because menu-document classifications are
	// filtered after resolution and must never leak into website media.
	resolved, err, _ := service.photoCalls.Do(restaurantID.String(), func() (any, error) {
		return service.photos.ListPhotoURLs(ctx, placeID, 10)
	})
	if err != nil {
		service.log.WarnContext(ctx, "google_places_demo_media_unavailable", "restaurant_id", restaurantID, "error", err)
		return items
	}
	photos, ok := resolved.([]placesprovider.Photo)
	if !ok {
		service.log.WarnContext(ctx, "google_places_demo_media_invalid", "restaurant_id", restaurantID)
		return items
	}
	live := make([]PublicMedia, 0, len(photos))
	for _, photo := range photos {
		hint, classified := hintsByFingerprint[photo.SourceFingerprint]
		if !classified {
			// Public demos fail closed: a freshly resolved Google photo is not
			// website-safe until its exact resource fingerprint was classified.
			continue
		}
		mediaType := normalizeGoogleMediaType(hint.MediaType)
		if !hint.PublicEligible || mediaType == "menu_document" {
			continue
		}
		if !IsWebsiteMediaType(mediaType) {
			mediaType = "other"
		}
		attributions := make([]Attribution, 0, len(photo.AuthorAttributions))
		for _, attribution := range photo.AuthorAttributions {
			attributions = append(attributions, Attribution{
				DisplayName: attribution.DisplayName,
				URI:         attribution.URI,
				PhotoURI:    attribution.PhotoURI,
			})
		}
		live = append(live, PublicMedia{
			URL:                photo.URL,
			SourceKind:         SourceGoogleLive,
			MediaType:          mediaType,
			AltText:            liveAltText(mediaType),
			WidthPx:            photo.WidthPx,
			HeightPx:           photo.HeightPx,
			PlacementRole:      livePlacement(mediaType),
			Unoptimized:        true,
			AuthorAttributions: attributions,
			GoogleMapsURI:      photo.GoogleMapsURI,
			FlagContentURI:     photo.FlagContentURI,
		})
	}
	sort.SliceStable(live, func(left, right int) bool {
		return livePriority(live[left].MediaType) < livePriority(live[right].MediaType)
	})
	for _, item := range live {
		if len(items) >= limit {
			break
		}
		items = append(items, item)
	}
	return items
}

func (service *Service) publicOwnedAssets(
	ctx context.Context,
	restaurantID uuid.UUID,
	limit int,
) []PublicMedia {
	assets, err := service.repository.ListPublic(ctx, restaurantID)
	if err != nil {
		service.log.WarnContext(ctx, "owned_restaurant_media_unavailable", "restaurant_id", restaurantID, "error", err)
		return []PublicMedia{}
	}
	items := make([]PublicMedia, 0, min(limit, len(assets)))
	for _, asset := range assets {
		if len(items) >= limit {
			break
		}
		if !IsWebsiteMediaType(asset.MediaType) {
			continue
		}
		url := ""
		if service.objects != nil {
			url = service.objects.PublicURL(asset.StorageKey)
		}
		if url == "" {
			continue
		}
		assetID := asset.ID
		items = append(items, PublicMedia{
			ID:              &assetID,
			URL:             url,
			SourceKind:      asset.SourceKind,
			MediaType:       asset.MediaType,
			Caption:         asset.Caption,
			AltText:         asset.AltText,
			Tags:            decodeTags(asset.Tags),
			QualityScore:    asset.QualityScore,
			HeroScore:       asset.HeroScore,
			WidthPx:         asset.WidthPx,
			HeightPx:        asset.HeightPx,
			Orientation:     asset.Orientation,
			SubjectPosition: asset.SubjectPosition,
			ContainsPeople:  asset.ContainsPeople,
			ContainsText:    asset.ContainsText,
			PlacementRole:   asset.PlacementRole,
			Unoptimized:     false,
		})
	}
	return items
}

func (service *Service) ListAdmin(ctx context.Context, restaurantID uuid.UUID) ([]PublicMedia, error) {
	assets, err := service.repository.ListAdmin(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	items := make([]PublicMedia, 0, len(assets))
	for _, asset := range assets {
		assetID := asset.ID
		items = append(items, PublicMedia{
			ID:               &assetID,
			URL:              service.objects.PublicURL(asset.StorageKey),
			SourceKind:       asset.SourceKind,
			MediaType:        asset.MediaType,
			Caption:          asset.Caption,
			AltText:          asset.AltText,
			Tags:             decodeTags(asset.Tags),
			QualityScore:     asset.QualityScore,
			HeroScore:        asset.HeroScore,
			WidthPx:          asset.WidthPx,
			HeightPx:         asset.HeightPx,
			Orientation:      asset.Orientation,
			SubjectPosition:  asset.SubjectPosition,
			ContainsPeople:   asset.ContainsPeople,
			ContainsText:     asset.ContainsText,
			PlacementRole:    asset.PlacementRole,
			ApprovalStatus:   asset.ApprovalStatus,
			RightsStatus:     asset.RightsStatus,
			VisionStatus:     asset.VisionStatus,
			VisionLastError:  asset.VisionLastError,
			VisionAnalyzedAt: asset.VisionAnalyzedAt,
			HiddenAt:         asset.HiddenAt,
		})
	}
	return items, nil
}

func (service *Service) CreateOwnedAsset(
	ctx context.Context,
	input CreateAssetInput,
	content []byte,
) (PublicMedia, error) {
	if service == nil || service.objects == nil || !service.objects.Configured() {
		return PublicMedia{}, ErrStorageUnavailable
	}
	if input.SourceKind != SourceOwnerUpload && input.SourceKind != SourceLicensed {
		return PublicMedia{}, fmt.Errorf("unsupported media source")
	}
	if !IsWebsiteMediaType(input.MediaType) {
		return PublicMedia{}, fmt.Errorf("unsupported website media type")
	}
	if input.RightsStatus != "owner_granted" && input.RightsStatus != "licensed" {
		return PublicMedia{}, fmt.Errorf("media rights confirmation is required")
	}
	if input.StorageKey == "" || len(content) == 0 {
		return PublicMedia{}, fmt.Errorf("media content is required")
	}
	if err := service.objects.Put(
		ctx,
		input.StorageKey,
		input.MimeType,
		bytes.NewReader(content),
		int64(len(content)),
	); err != nil {
		return PublicMedia{}, err
	}
	asset, err := service.repository.Create(ctx, input)
	if err != nil {
		if cleanupErr := service.objects.Delete(ctx, input.StorageKey); cleanupErr != nil {
			service.log.ErrorContext(ctx, "orphaned_restaurant_media_object", "storage_key", input.StorageKey, "error", cleanupErr)
		}
		return PublicMedia{}, err
	}
	assetID := asset.ID
	return PublicMedia{
		ID:             &assetID,
		URL:            service.objects.PublicURL(asset.StorageKey),
		SourceKind:     asset.SourceKind,
		MediaType:      asset.MediaType,
		Caption:        asset.Caption,
		AltText:        asset.AltText,
		WidthPx:        asset.WidthPx,
		HeightPx:       asset.HeightPx,
		PlacementRole:  asset.PlacementRole,
		ApprovalStatus: asset.ApprovalStatus,
		RightsStatus:   asset.RightsStatus,
		VisionStatus:   asset.VisionStatus,
	}, nil
}

func (service *Service) SetHidden(
	ctx context.Context,
	restaurantID, assetID uuid.UUID,
	hiddenBy *uuid.UUID,
) error {
	return service.repository.SetHidden(ctx, restaurantID, assetID, hiddenBy)
}

func decodeTags(raw json.RawMessage) []string {
	var tags []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &tags)
	}
	return tags
}

func normalizeGoogleMediaType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "food_photo", "food":
		return "food"
	case "interior", "ambience":
		return "interior"
	case "logo":
		return "logo"
	case "menu_document", "menu_list", "menu_ocr":
		return "menu_document"
	case "drink", "exterior", "team", "event", "other":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "other"
	}
}

func livePriority(mediaType string) int {
	switch mediaType {
	case "exterior":
		return 0
	case "interior":
		return 1
	case "food", "drink":
		return 2
	case "team", "event":
		return 3
	case "logo":
		return 4
	default:
		return 5
	}
}

func livePlacement(mediaType string) string {
	switch mediaType {
	case "exterior", "interior":
		return "ambience_gallery"
	case "food", "drink":
		return "food_gallery"
	case "logo":
		return "logo"
	default:
		return "gallery"
	}
}

func liveAltText(mediaType string) string {
	switch mediaType {
	case "exterior":
		return "Restaurant exterior"
	case "interior":
		return "Restaurant interior and atmosphere"
	case "food":
		return "Food served at the restaurant"
	case "drink":
		return "Drink served at the restaurant"
	case "logo":
		return "Restaurant logo"
	case "team":
		return "Restaurant team"
	case "event":
		return "Restaurant event"
	default:
		return "Restaurant photo"
	}
}
