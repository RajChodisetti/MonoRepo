package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/http/handlers"
)

func TestTrackingClickRedirects(t *testing.T) {
	campaignID := uuid.New()
	restaurantID := uuid.New()
	token := "click-token"

	repo := &campaigns.Mock{
		Tokens: map[string]campaigns.TrackingToken{
			token: {
				Token:        token,
				CampaignID:   campaignID,
				RestaurantID: restaurantID,
				TokenType:    campaigns.TokenClick,
				TargetURL:    "http://localhost:3000/demo/test",
			},
		},
	}

	handler := handlers.NewTrackingHandler(repo, func(w http.ResponseWriter, status int, code, message string) {
		http.Error(w, message, status)
	})

	req := httptest.NewRequest(http.MethodGet, "/t/click/"+token, nil)
	req.SetPathValue("token", token)
	rec := httptest.NewRecorder()

	handler.Click(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); location != "http://localhost:3000/demo/test" {
		t.Fatalf("location = %q", location)
	}

	events, err := repo.ListEvents(context.Background(), campaignID)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].EventType != campaigns.EventClicked {
		t.Fatalf("events = %+v, want one clicked event", events)
	}
}
