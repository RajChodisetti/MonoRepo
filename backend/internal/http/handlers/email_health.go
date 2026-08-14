package handlers

import (
	"net/http"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type EmailHealthHandler struct {
	service    emailprovider.HealthMonitor
	cfg        config.OutreachConfig
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewEmailHealthHandler(
	service emailprovider.HealthMonitor,
	cfg config.OutreachConfig,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *EmailHealthHandler {
	return &EmailHealthHandler{service: service, cfg: cfg, writeJSON: writeJSON, writeError: writeError}
}

func (handler *EmailHealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	if handler.service == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "email_health_unavailable", "Email account health is not configured.")
		return
	}
	accounts, err := handler.service.List(r.Context())
	if err != nil {
		handler.writeError(w, http.StatusInternalServerError, "email_health_unavailable", "Email account health is unavailable.")
		return
	}
	intervalHours := int(handler.cfg.EmailHealthInterval / time.Hour)
	handler.writeJSON(w, http.StatusOK, map[string]any{
		"enabled":        handler.cfg.EmailHealthEnabled,
		"recipient":      handler.cfg.EmailHealthRecipient,
		"interval_hours": intervalHours,
		"accounts":       accounts,
	})
}
