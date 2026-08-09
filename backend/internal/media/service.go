package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	resolved, err, _ := service.photoCalls.Do(restaurantID.String(), func() (any, error) {
		return service.photos.ListPhotoURLs(ctx, placeID, min(10, limit-len(items)))
	})
	if err != nil {
		service.log.WarnContext(ctx, "google_places_public_media_unavailable", "restaurant_id", restaurantID, "error", err)
		return items
	}
	photos, ok := resolved.([]placesprovider.Photo)
	if !ok {
		service.log.WarnContext(ctx, "google_places_public_media_invalid", "restaurant_id", restaurantID)
		return items
	}
	for _, photo := range photos {
		if len(items) >= limit {
			break
		}
		if strings.TrimSpace(photo.URL) == "" {
			continue
		}
		items = append(items, PublicMedia{
			URL:                photo.URL,
			SourceKind:         SourceGoogleLive,
			MediaType:          "other",
			AltText:            "Restaurant photo from Google Maps",
			WidthPx:            photo.WidthPx,
			HeightPx:           photo.HeightPx,
			PlacementRole:      "gallery",
			Unoptimized:        true,
			AuthorAttributions: googleAttributions(photo.AuthorAttributions),
			GoogleMapsURI:      photo.GoogleMapsURI,
			FlagContentURI:     photo.FlagContentURI,
		})
	}
	return items
}

// PreviewForRestaurant intentionally shares the public media policy.
func (service *Service) PreviewForRestaurant(
	ctx context.Context,
	restaurantID uuid.UUID,
	limit int,
) []PublicMedia {
	return service.PublicForRestaurant(ctx, restaurantID, limit)
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
			ApprovalStatus:  asset.ApprovalStatus,
			ReviewedAt:      asset.ReviewedAt,
			ReviewedBy:      asset.ReviewedBy,
			ReviewNote:      asset.ReviewNote,
			RightsStatus:    asset.RightsStatus,
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
			ReviewedAt:       asset.ReviewedAt,
			ReviewedBy:       asset.ReviewedBy,
			ReviewNote:       asset.ReviewNote,
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

func (service *Service) ReviewOwnedAsset(
	ctx context.Context,
	restaurantID, assetID, reviewedBy uuid.UUID,
	approvalStatus, note string,
) (PublicMedia, error) {
	approvalStatus = strings.ToLower(strings.TrimSpace(approvalStatus))
	if approvalStatus != "approved" && approvalStatus != "rejected" {
		return PublicMedia{}, fmt.Errorf("approval status must be approved or rejected")
	}
	asset, err := service.repository.SetApproval(
		ctx, restaurantID, assetID, reviewedBy, approvalStatus, strings.TrimSpace(note),
	)
	if err != nil {
		return PublicMedia{}, err
	}
	url := ""
	if service.objects != nil {
		url = service.objects.PublicURL(asset.StorageKey)
	}
	assetIDCopy := asset.ID
	return PublicMedia{
		ID:             &assetIDCopy,
		URL:            url,
		SourceKind:     asset.SourceKind,
		MediaType:      asset.MediaType,
		Caption:        asset.Caption,
		AltText:        asset.AltText,
		WidthPx:        asset.WidthPx,
		HeightPx:       asset.HeightPx,
		PlacementRole:  asset.PlacementRole,
		ApprovalStatus: asset.ApprovalStatus,
		ReviewedAt:     asset.ReviewedAt,
		ReviewedBy:     asset.ReviewedBy,
		ReviewNote:     asset.ReviewNote,
		RightsStatus:   asset.RightsStatus,
		HiddenAt:       asset.HiddenAt,
	}, nil
}

func decodeTags(raw json.RawMessage) []string {
	var tags []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &tags)
	}
	return tags
}

func googleAttributions(source []placesprovider.Attribution) []Attribution {
	attributions := make([]Attribution, 0, len(source))
	for _, attribution := range source {
		attributions = append(attributions, Attribution{
			DisplayName: attribution.DisplayName,
			URI:         attribution.URI,
			PhotoURI:    attribution.PhotoURI,
		})
	}
	return attributions
}
