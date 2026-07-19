package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	profilesrepo "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
)

type RestaurantSiteAdminHandler struct {
	profiles     profiles.Repository
	publicWebURL string
	writeJSON    func(http.ResponseWriter, int, any)
	writeError   func(http.ResponseWriter, int, string, string)
}

func NewRestaurantSiteAdminHandler(
	profilesRepo profiles.Repository,
	publicWebURL string,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *RestaurantSiteAdminHandler {
	return &RestaurantSiteAdminHandler{
		profiles:     profilesRepo,
		publicWebURL: strings.TrimRight(publicWebURL, "/"),
		writeJSON:    writeJSON,
		writeError:   writeError,
	}
}

func (handler *RestaurantSiteAdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := restaurantIDFromRequest(r)
	if err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Restaurant id must be a valid UUID.")
		return
	}

	summary, err := handler.profiles.GetSiteRestaurantByID(r.Context(), restaurantID)
	if err != nil {
		if errors.Is(err, profilesrepo.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "A generated website is not available for this restaurant.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}

	templates := []map[string]any{
		{"id": "1", "name": "Cinematic", "url": generatedSiteURL(handler.publicWebURL, summary.ID.String(), summary.Index, "1")},
		{"id": "2", "name": "Aurora", "url": generatedSiteURL(handler.publicWebURL, summary.ID.String(), summary.Index, "2")},
		{"id": "3", "name": "Elysian reservations", "url": generatedSiteURL(handler.publicWebURL, summary.ID.String(), summary.Index, "3")},
		{"id": "4", "name": "Italian Villa experimental", "url": generatedSiteURL(handler.publicWebURL, summary.ID.String(), summary.Index, "4")},
	}
	handler.writeJSON(w, http.StatusOK, map[string]any{
		"restaurant_id":   summary.ID,
		"restaurant_name": summary.Name,
		"google_place_id": summary.PlaceID,
		"site_index":      summary.Index,
		"templates":       templates,
		"shareable":       false,
	})
}

func generatedSiteURL(baseURL, restaurantID string, index int, templateID string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	query := parsed.Query()
	query.Set("id", strconv.Itoa(index))
	query.Set("restaurant_id", strings.TrimSpace(restaurantID))
	query.Set("template", templateID)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
