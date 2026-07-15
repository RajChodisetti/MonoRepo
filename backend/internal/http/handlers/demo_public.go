package handlers

import (
	"errors"
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type DemoPublicHandler struct {
	demoService *demos.Service
	writeJSON   func(http.ResponseWriter, int, any)
	writeError  func(http.ResponseWriter, int, string, string)
}

func NewDemoPublicHandler(
	demoService *demos.Service,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *DemoPublicHandler {
	return &DemoPublicHandler{
		demoService: demoService,
		writeJSON:   writeJSON,
		writeError:  writeError,
	}
}

func (handler *DemoPublicHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	slug := r.PathValue("slug")
	token := r.URL.Query().Get("token")

	payload, err := handler.demoService.ResolvePublicDemo(r.Context(), slug, token)
	if err != nil {
		if errors.Is(err, demos.ErrDemoNotFound) {
			handler.writeError(w, http.StatusNotFound, "demo_not_found", "Demo site was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}

	handler.writeJSON(w, http.StatusOK, payload)
}
