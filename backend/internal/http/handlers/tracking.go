package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

var trackingPixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21,
	0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

type TrackingHandler struct {
	repo        campaigns.Repository
	restaurants restaurants.Repository
	writeError  func(http.ResponseWriter, int, string, string)
}

func NewTrackingHandler(
	repo campaigns.Repository,
	restaurantRepo restaurants.Repository,
	writeError func(http.ResponseWriter, int, string, string),
) *TrackingHandler {
	return &TrackingHandler{repo: repo, restaurants: restaurantRepo, writeError: writeError}
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
	if err := handler.repo.InsertEvent(r.Context(), record.CampaignID, record.RestaurantID, campaigns.EventClicked, meta); err == nil && handler.restaurants != nil {
		_, _ = handler.restaurants.MarkShownInterest(r.Context(), record.RestaurantID)
	}

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

func (handler *TrackingHandler) UnsubscribeConfirm(w http.ResponseWriter, r *http.Request) {
	setUnsubscribeSecurityHeaders(w)
	record, err := handler.repo.GetTrackingToken(r.Context(), r.PathValue("token"))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.writeError(w, http.StatusNotFound, "not_found", "Unsubscribe link was not found.")
			return
		}
		handler.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred.")
		return
	}
	if record.TokenType != campaigns.TokenUnsubscribe || strings.TrimSpace(record.RecipientEmail) == "" {
		handler.writeError(w, http.StatusNotFound, "not_found", "Unsubscribe link was not found.")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Manage email preferences</title><style>body{font-family:system-ui,sans-serif;background:#f7f7f5;color:#171717;margin:0}main{max-width:34rem;margin:10vh auto;padding:2rem;background:white;border:1px solid #ddd;border-radius:1rem}button,a{font:inherit}button{min-height:44px;padding:.7rem 1rem;border:0;border-radius:.6rem;background:#171717;color:white;font-weight:700;cursor:pointer}a{color:#171717}p{line-height:1.55}</style></head><body><main><h1>Stop outreach emails?</h1><p>Confirm below and Tuvi Solutions will stop sending outreach emails to this address.</p><form method="post"><button type="submit">Opt out of emails</button></form><p><a href="https://tuvisolutions.com" rel="noreferrer">Visit Tuvi Solutions</a></p></main></body></html>`))
}

func (handler *TrackingHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	setUnsubscribeSecurityHeaders(w)
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

	recipientEmail := strings.ToLower(strings.TrimSpace(record.RecipientEmail))
	if recipientEmail == "" {
		handler.writeError(w, http.StatusInternalServerError, "unsubscribe_failed", "Your unsubscribe request could not be saved. Please retry.")
		return
	}
	reason := "unsubscribed via link"
	if !record.RecipientSnapshot {
		// Pre-000021 tokens have no immutable recipient history. Preserve the
		// formerly supported opt-out by suppressing the restaurant's current
		// normalized address; all newly created tokens use the exact snapshot.
		reason = "legacy unsubscribe fallback via current restaurant email"
	}
	if err := handler.repo.AddSuppression(r.Context(), recipientEmail, reason); err != nil {
		handler.writeError(w, http.StatusInternalServerError, "unsubscribe_failed", "Your unsubscribe request could not be saved. Please retry.")
		return
	}

	meta, _ := json.Marshal(map[string]any{"recipient_snapshot": record.RecipientSnapshot})
	_ = handler.repo.InsertEvent(r.Context(), record.CampaignID, record.RestaurantID, campaigns.EventUnsubscribed, meta)
	_, _ = handler.repo.Stop(r.Context(), record.CampaignID, "recipient unsubscribed")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Opt-out confirmed</title><style>body{font-family:system-ui,sans-serif;background:#f7f7f5;color:#171717;margin:0}main{max-width:34rem;margin:10vh auto;padding:2rem;background:white;border:1px solid #ddd;border-radius:1rem}a{color:#171717}p{line-height:1.55}</style></head><body><main><h1>You have opted out</h1><p>Tuvi Solutions will no longer send outreach emails to this address.</p><p><a href="https://tuvisolutions.com" rel="noreferrer">Visit Tuvi Solutions</a></p></main></body></html>`))
}

func setUnsubscribeSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}
