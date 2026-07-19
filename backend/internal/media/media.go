package media

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

const (
	SourceOwnerUpload = "owner_upload"
	SourceLicensed    = "licensed"
	SourceGoogleLive  = "google_places_live"
)

var ErrStorageUnavailable = errors.New("restaurant media storage is unavailable")

type Attribution struct {
	DisplayName string `json:"display_name"`
	URI         string `json:"uri,omitempty"`
	PhotoURI    string `json:"photo_uri,omitempty"`
}

// PublicMedia is safe for a restaurant website. Menu-document images are not
// represented by this type and are filtered at every service boundary.
type PublicMedia struct {
	ID                 *uuid.UUID    `json:"id,omitempty"`
	URL                string        `json:"url"`
	SourceKind         string        `json:"source_kind"`
	MediaType          string        `json:"media_type"`
	Caption            string        `json:"caption,omitempty"`
	AltText            string        `json:"alt_text,omitempty"`
	Tags               []string      `json:"tags,omitempty"`
	QualityScore       *float64      `json:"quality_score,omitempty"`
	HeroScore          *float64      `json:"hero_score,omitempty"`
	WidthPx            int           `json:"width_px,omitempty"`
	HeightPx           int           `json:"height_px,omitempty"`
	Orientation        string        `json:"orientation,omitempty"`
	SubjectPosition    string        `json:"subject_position,omitempty"`
	ContainsPeople     bool          `json:"contains_people,omitempty"`
	ContainsText       bool          `json:"contains_text,omitempty"`
	PlacementRole      string        `json:"placement_role,omitempty"`
	ApprovalStatus     string        `json:"approval_status,omitempty"`
	RightsStatus       string        `json:"rights_status,omitempty"`
	VisionStatus       string        `json:"vision_status,omitempty"`
	VisionLastError    string        `json:"vision_last_error,omitempty"`
	VisionAnalyzedAt   *time.Time    `json:"vision_analyzed_at,omitempty"`
	HiddenAt           *time.Time    `json:"hidden_at,omitempty"`
	Unoptimized        bool          `json:"unoptimized"`
	AuthorAttributions []Attribution `json:"author_attributions,omitempty"`
	GoogleMapsURI      string        `json:"google_maps_uri,omitempty"`
	FlagContentURI     string        `json:"flag_content_uri,omitempty"`
}

type Asset struct {
	ID               uuid.UUID       `json:"id"`
	RestaurantID     uuid.UUID       `json:"restaurant_id"`
	SourceKind       string          `json:"source_kind"`
	StorageKey       string          `json:"storage_key"`
	MediaType        string          `json:"media_type"`
	Caption          string          `json:"caption"`
	AltText          string          `json:"alt_text"`
	Tags             json.RawMessage `json:"tags"`
	QualityScore     *float64        `json:"quality_score,omitempty"`
	HeroScore        *float64        `json:"hero_score,omitempty"`
	Orientation      string          `json:"orientation"`
	SubjectPosition  string          `json:"subject_position"`
	ContainsPeople   bool            `json:"contains_people"`
	ContainsText     bool            `json:"contains_text"`
	PlacementRole    string          `json:"placement_role"`
	ApprovalStatus   string          `json:"approval_status"`
	RightsStatus     string          `json:"rights_status"`
	MimeType         string          `json:"mime_type"`
	WidthPx          int             `json:"width_px"`
	HeightPx         int             `json:"height_px"`
	ByteSize         int64           `json:"byte_size"`
	SHA256           string          `json:"sha256"`
	SortOrder        int             `json:"sort_order"`
	VisionStatus     string          `json:"vision_status"`
	VisionAttempts   int             `json:"vision_attempts"`
	VisionLastError  string          `json:"vision_last_error"`
	VisionResult     json.RawMessage `json:"vision_result"`
	VisionAnalyzedAt *time.Time      `json:"vision_analyzed_at,omitempty"`
	HiddenAt         *time.Time      `json:"hidden_at,omitempty"`
	HiddenBy         *uuid.UUID      `json:"hidden_by,omitempty"`
	CreatedBy        *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type CreateAssetInput struct {
	RestaurantID    uuid.UUID
	SourceKind      string
	StorageKey      string
	MediaType       string
	Caption         string
	AltText         string
	Orientation     string
	SubjectPosition string
	PlacementRole   string
	ApprovalStatus  string
	RightsStatus    string
	MimeType        string
	WidthPx         int
	HeightPx        int
	ByteSize        int64
	SHA256          string
	CreatedBy       uuid.UUID
}

type ClassificationHint struct {
	SourceIndex       int
	SourceFingerprint string
	MediaType         string
	Confidence        float64
	PublicEligible    bool
}

type Repository interface {
	ListPublic(ctx context.Context, restaurantID uuid.UUID) ([]Asset, error)
	ListAdmin(ctx context.Context, restaurantID uuid.UUID) ([]Asset, error)
	ListClassificationHints(ctx context.Context, restaurantID uuid.UUID) ([]ClassificationHint, error)
	Create(ctx context.Context, input CreateAssetInput) (Asset, error)
	SetHidden(ctx context.Context, restaurantID, assetID uuid.UUID, hiddenBy *uuid.UUID) error
}

type ObjectStore interface {
	Configured() bool
	Put(ctx context.Context, key, contentType string, body io.Reader, size int64) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

func IsWebsiteMediaType(value string) bool {
	switch value {
	case "exterior", "interior", "food", "drink", "logo", "team", "event", "other":
		return true
	default:
		return false
	}
}
