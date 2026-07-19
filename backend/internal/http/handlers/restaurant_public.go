package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/media"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
)

type RestaurantPublicHandler struct {
	profiles   profiles.Repository
	media      *media.Service
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewRestaurantPublicHandler(
	profilesRepo profiles.Repository,
	mediaService *media.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *RestaurantPublicHandler {
	return &RestaurantPublicHandler{
		profiles:   profilesRepo,
		media:      mediaService,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

func (handler *RestaurantPublicHandler) GetSiteImagesByID(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_id", "Restaurant id must be a valid UUID.")
		return
	}

	payload, err := handler.profiles.GetSiteImages(r.Context(), restaurantID)
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":  payload.RestaurantID,
		"gallery_images": payload.GalleryImages,
	})
}

func (handler *RestaurantPublicHandler) GetSiteImagesByPlaceID(w http.ResponseWriter, r *http.Request) {
	placeID := r.PathValue("place_id")
	if placeID == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_place_id", "Google place id is required.")
		return
	}

	payload, err := handler.profiles.GetSiteImagesByPlaceID(r.Context(), placeID)
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":  payload.RestaurantID,
		"gallery_images": payload.GalleryImages,
	})
}

func (handler *RestaurantPublicHandler) ListSiteRestaurants(w http.ResponseWriter, r *http.Request) {
	payload, err := handler.profiles.ListSiteRestaurants(r.Context())
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{
		"count":       len(payload),
		"restaurants": payload,
	})
}

func (handler *RestaurantPublicHandler) GetSiteContentByIndex(w http.ResponseWriter, r *http.Request) {
	indexStr := r.PathValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		handler.writeError(w, http.StatusBadRequest, "invalid_index", "Site index must be a non-negative integer.")
		return
	}

	payload, err := handler.profiles.GetSiteContentByIndex(r.Context(), index)
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}

	handler.writeSiteContent(w, r, payload)
}

func (handler *RestaurantPublicHandler) GetSiteContentByID(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_id", "Restaurant id must be a valid UUID.")
		return
	}
	payload, err := handler.profiles.GetSiteContentByID(r.Context(), restaurantID)
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}
	handler.writeSiteContent(w, r, payload)
}

func (handler *RestaurantPublicHandler) GetSiteContentByPlaceID(w http.ResponseWriter, r *http.Request) {
	placeID := r.PathValue("place_id")
	if placeID == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_place_id", "Google place id is required.")
		return
	}

	payload, err := handler.profiles.GetSiteContentByPlaceID(r.Context(), placeID)
	if err != nil {
		handler.writeNotFoundOrInternal(w, err)
		return
	}

	handler.writeSiteContent(w, r, payload)
}

func (handler *RestaurantPublicHandler) writeSiteContent(
	w http.ResponseWriter,
	r *http.Request,
	payload profiles.SiteContent,
) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	mediaItems := []media.PublicMedia{}
	if handler.media != nil && payload.RestaurantID != uuid.Nil {
		if r.URL.Query().Get("preview_media") == "google_live" {
			mediaItems = handler.media.PreviewForRestaurant(r.Context(), payload.RestaurantID, 10)
		} else {
			mediaItems = handler.media.PublicForRestaurant(r.Context(), payload.RestaurantID, 10)
		}
	}
	handler.writeJSON(w, http.StatusOK, struct {
		profiles.SiteContent
		Media []media.PublicMedia `json:"media"`
	}{
		SiteContent: payload,
		Media:       mediaItems,
	})
}

func (handler *RestaurantPublicHandler) writeNotFoundOrInternal(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		handler.writeError(w, http.StatusNotFound, "not_found", "Restaurant images were not found.")
		return
	}
	handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
}
