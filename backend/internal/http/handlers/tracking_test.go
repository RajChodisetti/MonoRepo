package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
)

func unsubscribeTestHandler(repo *campaigns.Mock) *TrackingHandler {
	return NewTrackingHandler(repo, nil, func(w http.ResponseWriter, status int, code, message string) {
		http.Error(w, fmt.Sprintf("%s: %s", code, message), status)
	})
}

func unsubscribeTestRepo() (*campaigns.Mock, string, string) {
	token := "unsubscribe-token"
	email := "owner@example.com"
	campaignID := uuid.New()
	restaurantID := uuid.New()
	return &campaigns.Mock{
		Campaigns: map[uuid.UUID]campaigns.Campaign{
			campaignID: {ID: campaignID, RestaurantID: restaurantID, Status: campaigns.StatusApproved},
		},
		Tokens: map[string]campaigns.TrackingToken{
			token: {
				Token: token, CampaignID: campaignID, RestaurantID: restaurantID,
				TokenType: campaigns.TokenUnsubscribe, RecipientEmail: email, RecipientSnapshot: true,
			},
		},
	}, token, email
}

func TestUnsubscribeGETOnlyShowsConfirmation(t *testing.T) {
	repo, token, email := unsubscribeTestRepo()
	handler := unsubscribeTestHandler(repo)
	request := httptest.NewRequest(http.MethodGet, "/t/unsubscribe/"+token, nil)
	request.SetPathValue("token", token)
	recorder := httptest.NewRecorder()

	handler.UnsubscribeConfirm(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "Opt out of emails") {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
	if suppressed, _ := repo.IsSuppressed(context.Background(), email); suppressed {
		t.Fatal("GET unsubscribe mutated suppression state")
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestUnsubscribePOSTSuppressesRecipient(t *testing.T) {
	repo, token, email := unsubscribeTestRepo()
	handler := unsubscribeTestHandler(repo)
	request := httptest.NewRequest(http.MethodPost, "/t/unsubscribe/"+token, nil)
	request.SetPathValue("token", token)
	recorder := httptest.NewRecorder()

	handler.Unsubscribe(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "You have opted out") {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
	if suppressed, _ := repo.IsSuppressed(context.Background(), email); !suppressed {
		t.Fatal("POST unsubscribe did not suppress recipient")
	}
}
