package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/media"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	placesprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/places"
)

// RestaurantImagesAdminHandler manages admin visibility (hide/restore) of
// scraped restaurant photos. It is separate from RestaurantPublicHandler,
// which serves the same underlying tables read-only and unauthenticated to
// the public demo site.
type RestaurantImagesAdminHandler struct {
	profiles   profiles.Repository
	photos     placesprovider.PhotoResolver
	media      *media.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewRestaurantImagesAdminHandler(
	profilesRepo profiles.Repository,
	photoResolver placesprovider.PhotoResolver,
	mediaService *media.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *RestaurantImagesAdminHandler {
	return &RestaurantImagesAdminHandler{
		profiles:   profilesRepo,
		photos:     photoResolver,
		media:      mediaService,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

// ListGoogle resolves fresh Google Places media URLs for one restaurant. The
// response is deliberately no-store: these URLs are for admin inspection and
// may expire, while the API key remains only on the server.
func (handler *RestaurantImagesAdminHandler) ListGoogle(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	placeID, err := handler.profiles.GetGooglePlaceID(r.Context(), restaurantID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "place_not_found", "This restaurant has no Google Place ID.")
			return
		}
		handler.writeInternalError(w, err)
		return
	}
	if handler.photos == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "photos_unavailable", "Google Places photo viewing is not configured.")
		return
	}

	photos, err := handler.photos.ListPhotoURLs(r.Context(), placeID, 10)
	if err != nil {
		if errors.Is(err, placesprovider.ErrNotConfigured) {
			handler.writeError(w, http.StatusServiceUnavailable, "photos_unavailable", "Google Places photo viewing is not configured.")
			return
		}
		handler.writeError(w, http.StatusBadGateway, "photo_provider_error", "Google Places photos could not be refreshed.")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":      restaurantID,
		"google_place_id":    placeID,
		"photos":             photos,
		"refreshed_at":       time.Now().UTC(),
		"urls_are_temporary": true,
	})
}

func (handler *RestaurantImagesAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	menuImages, err := handler.profiles.ListMenuImagesAdmin(r.Context(), restaurantID)
	if err != nil {
		handler.writeInternalError(w, err)
		return
	}
	galleryImages, err := handler.profiles.ListGalleryImagesAdmin(r.Context(), restaurantID)
	if err != nil {
		handler.writeInternalError(w, err)
		return
	}

	ownedMedia := []media.PublicMedia{}
	if handler.media != nil {
		ownedMedia, err = handler.media.ListAdmin(r.Context(), restaurantID)
		if err != nil {
			handler.writeInternalError(w, err)
			return
		}
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":  restaurantID,
		"menu_images":    menuImages,
		"gallery_images": galleryImages,
		"owned_media":    ownedMedia,
	})
}

const maxRestaurantMediaUploadBytes = 15 * 1024 * 1024

func (handler *RestaurantImagesAdminHandler) Upload(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}
	if handler.media == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "media_storage_unavailable", "Restaurant media storage is unavailable.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRestaurantMediaUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(maxRestaurantMediaUploadBytes + (1 << 20)); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_upload", "The media upload is invalid or too large.")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "missing_file", "An image file is required.")
		return
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxRestaurantMediaUploadBytes+1))
	if err != nil || len(content) == 0 || len(content) > maxRestaurantMediaUploadBytes {
		handler.writeError(w, http.StatusBadRequest, "invalid_upload", "The image must be between 1 byte and 15 MB.")
		return
	}
	mimeType := http.DetectContentType(content)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
		handler.writeError(w, http.StatusBadRequest, "unsupported_image", "Only JPEG, PNG, and GIF images are supported.")
		return
	}
	imageConfig, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || imageConfig.Width < 1 || imageConfig.Height < 1 {
		handler.writeError(w, http.StatusBadRequest, "invalid_image", "The uploaded file is not a valid image.")
		return
	}
	if imageConfig.Width > 12000 || imageConfig.Height > 12000 || int64(imageConfig.Width)*int64(imageConfig.Height) > 80_000_000 {
		handler.writeError(w, http.StatusBadRequest, "image_dimensions_too_large", "The image dimensions are too large for website media.")
		return
	}
	mediaType := strings.ToLower(strings.TrimSpace(r.FormValue("media_type")))
	if !media.IsWebsiteMediaType(mediaType) {
		handler.writeError(w, http.StatusBadRequest, "invalid_media_type", "Menu photos are not allowed; choose a website media type.")
		return
	}
	rightsStatus := strings.ToLower(strings.TrimSpace(r.FormValue("rights_status")))
	sourceKind := media.SourceOwnerUpload
	if rightsStatus == "licensed" {
		sourceKind = media.SourceLicensed
	} else if rightsStatus != "owner_granted" {
		handler.writeError(w, http.StatusBadRequest, "rights_required", "Confirm owner-granted or licensed image rights.")
		return
	}
	placementRole := strings.ToLower(strings.TrimSpace(r.FormValue("placement_role")))
	if !validPlacementRole(placementRole) {
		placementRole = "gallery"
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(content))
	assetID := uuid.New()
	storageKey := filepath.ToSlash(filepath.Join(
		"restaurants",
		restaurantID.String(),
		"media",
		assetID.String()+extensionForMIME(mimeType),
	))
	created, err := handler.media.CreateOwnedAsset(r.Context(), media.CreateAssetInput{
		RestaurantID:    restaurantID,
		SourceKind:      sourceKind,
		StorageKey:      storageKey,
		MediaType:       mediaType,
		Caption:         strings.TrimSpace(r.FormValue("caption")),
		AltText:         strings.TrimSpace(r.FormValue("alt_text")),
		Orientation:     mediaOrientation(imageConfig.Width, imageConfig.Height),
		SubjectPosition: "center",
		PlacementRole:   placementRole,
		// An internal administrator must approve this asset before it can
		// become public website media.
		ApprovalStatus: "draft",
		RightsStatus:   rightsStatus,
		MimeType:       mimeType,
		WidthPx:        imageConfig.Width,
		HeightPx:       imageConfig.Height,
		ByteSize:       int64(len(content)),
		SHA256:         digest,
		CreatedBy:      principal.UserID,
	}, content)
	if err != nil {
		if errors.Is(err, media.ErrStorageUnavailable) {
			handler.writeError(w, http.StatusServiceUnavailable, "media_storage_unavailable", "Restaurant media storage is not configured.")
			return
		}
		handler.writeInternalError(w, err)
		return
	}
	w.Header().Set("Location", created.URL)
	handler.writeJSON(w, http.StatusCreated, map[string]any{
		"asset":             created,
		"original_filename": filepath.Base(header.Filename),
	})
}

type reviewOwnedMediaRequest struct {
	ApprovalStatus string `json:"approval_status"`
	Note           string `json:"note"`
}

func (handler *RestaurantImagesAdminHandler) ReviewOwned(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}
	if handler.media == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "media_storage_unavailable", "Restaurant media storage is unavailable.")
		return
	}
	restaurantID, assetID, err := restaurantMediaIDsFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	var request reviewOwnedMediaRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "A valid media review decision is required.")
		return
	}
	asset, err := handler.media.ReviewOwnedAsset(
		r.Context(), restaurantID, assetID, principal.UserID, request.ApprovalStatus, request.Note,
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Image was not found.")
			return
		}
		if strings.Contains(err.Error(), "approval status") {
			handler.writeError(w, http.StatusBadRequest, "invalid_approval_status", "Approval status must be approved or rejected.")
			return
		}
		handler.writeInternalError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"asset": asset})
}

func (handler *RestaurantImagesAdminHandler) HideOwned(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}
	if handler.media == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "media_storage_unavailable", "Restaurant media storage is unavailable.")
		return
	}
	restaurantID, assetID, err := restaurantMediaIDsFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := handler.media.SetHidden(r.Context(), restaurantID, assetID, &principal.UserID); err != nil {
		handler.writeInternalError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"status": "hidden"})
}

func (handler *RestaurantImagesAdminHandler) RestoreOwned(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.PrincipalFromContext(r.Context()); !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}
	if handler.media == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "media_storage_unavailable", "Restaurant media storage is unavailable.")
		return
	}
	restaurantID, assetID, err := restaurantMediaIDsFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := handler.media.SetHidden(r.Context(), restaurantID, assetID, nil); err != nil {
		handler.writeInternalError(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{"status": "visible"})
}

func restaurantMediaIDsFromRequest(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("restaurant id must be a valid UUID")
	}
	assetID, err := uuid.Parse(r.PathValue("assetId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, errors.New("asset id must be a valid UUID")
	}
	return restaurantID, assetID, nil
}

func validPlacementRole(value string) bool {
	switch value {
	case "hero", "about", "gallery", "food_gallery", "ambience_gallery", "logo":
		return true
	default:
		return false
	}
}

func mediaOrientation(width, height int) string {
	if width == height {
		return "square"
	}
	if width > height {
		return "landscape"
	}
	return "portrait"
}

func extensionForMIME(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func (handler *RestaurantImagesAdminHandler) Hide(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		handler.writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
		return
	}

	restaurantID, imageID, kind, err := handler.parseImagePath(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if kind == "menu" {
		err = handler.profiles.HideMenuImage(r.Context(), restaurantID, imageID, principal.UserID)
	} else {
		err = handler.profiles.HideGalleryImage(r.Context(), restaurantID, imageID, principal.UserID)
	}
	if err != nil {
		handler.writeInternalError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{"status": "hidden"})
}

func (handler *RestaurantImagesAdminHandler) Unhide(w http.ResponseWriter, r *http.Request) {
	restaurantID, imageID, kind, err := handler.parseImagePath(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if kind == "menu" {
		err = handler.profiles.UnhideMenuImage(r.Context(), restaurantID, imageID)
	} else {
		err = handler.profiles.UnhideGalleryImage(r.Context(), restaurantID, imageID)
	}
	if err != nil {
		handler.writeInternalError(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{"status": "visible"})
}

func (handler *RestaurantImagesAdminHandler) parseImagePath(r *http.Request) (restaurantID, imageID uuid.UUID, kind string, err error) {
	restaurantID, err = restaurantIDFromRequest(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errors.New("restaurant id must be a valid UUID")
	}
	imageID, err = uuid.Parse(r.PathValue("imageId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, "", errors.New("image id must be a valid UUID")
	}
	kind = r.PathValue("kind")
	if kind != "menu" && kind != "gallery" {
		return uuid.Nil, uuid.Nil, "", errors.New("kind must be 'menu' or 'gallery'")
	}
	return restaurantID, imageID, kind, nil
}

func (handler *RestaurantImagesAdminHandler) writeInternalError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		handler.writeError(w, http.StatusNotFound, "not_found", "Image was not found.")
		return
	}
	handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
}
