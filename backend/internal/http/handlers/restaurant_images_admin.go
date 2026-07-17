package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
)

// RestaurantImagesAdminHandler manages admin visibility (hide/restore) of
// scraped restaurant photos. It is separate from RestaurantPublicHandler,
// which serves the same underlying tables read-only and unauthenticated to
// the public demo site.
type RestaurantImagesAdminHandler struct {
	profiles   profiles.Repository
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewRestaurantImagesAdminHandler(
	profilesRepo profiles.Repository,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *RestaurantImagesAdminHandler {
	return &RestaurantImagesAdminHandler{
		profiles:   profilesRepo,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
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

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":  restaurantID,
		"menu_images":    menuImages,
		"gallery_images": galleryImages,
	})
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
