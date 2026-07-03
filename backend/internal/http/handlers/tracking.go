package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var trackingPixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21,
	0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

type TrackingHandler struct {
	repo       campaigns.Repository
	writeError func(http.ResponseWriter, int, string, string)
}

func NewTrackingHandler(
	repo campaigns.Repository,
	writeError func(http.ResponseWriter, int, string, string),
) *TrackingHandler {
	return &TrackingHandler{repo: repo, writeError: writeError}
}

func (handler *TrackingHandler) Click(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	record, err := handler.repo.GetTrackingToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Tracking link was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}
	if record.TokenType != campaigns.TokenClick {
		handler.writeError(w, http.StatusNotFound, "not_found", "Tracking link was not found.")
		return
	}

	meta, _ := json.Marshal(map[string]string{"token": token, "target_url": record.TargetURL})
	_ = handler.repo.InsertEvent(r.Context(), record.CampaignID, record.RestaurantID, campaigns.EventClicked, meta)

	target := strings.TrimSpace(record.TargetURL)
	if target == "" {
		handler.writeError(w, http.StatusBadRequest, "invalid_request", "Tracking link has no destination.")
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (handler *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	record, err := handler.repo.GetTrackingToken(r.Context(), token)
	if err != nil {
		w.Header().Set("Content-Type", "image/gif")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(trackingPixelGIF)
		return
	}
	if record.TokenType != campaigns.TokenOpen {
		w.Header().Set("Content-Type", "image/gif")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(trackingPixelGIF)
		return
	}

	meta, _ := json.Marshal(map[string]string{"token": token})
	_ = handler.repo.InsertEvent(r.Context(), record.CampaignID, record.RestaurantID, campaigns.EventOpened, meta)

	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(trackingPixelGIF)
}

func (handler *TrackingHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	record, err := handler.repo.GetTrackingToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Unsubscribe link was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}
	if record.TokenType != campaigns.TokenUnsubscribe {
		handler.writeError(w, http.StatusNotFound, "not_found", "Unsubscribe link was not found.")
		return
	}

	sendCtx, err := handler.repo.GetSendContext(r.Context(), record.CampaignID)
	if err == nil && strings.TrimSpace(sendCtx.RestaurantEmail) != "" {
		_ = handler.repo.AddSuppression(r.Context(), sendCtx.RestaurantEmail, "unsubscribed via link")
	}

	meta, _ := json.Marshal(map[string]string{"token": token})
	_ = handler.repo.InsertEvent(r.Context(), record.CampaignID, record.RestaurantID, campaigns.EventUnsubscribed, meta)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><h1>Unsubscribed</h1><p>You will no longer receive outreach emails from Tuvi Solutions.</p></body></html>`))
}
