package email

import (
	"context"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type healthTestProvider struct {
	requests []SendRequest
}

func (provider *healthTestProvider) Send(_ context.Context, request SendRequest) (SendResult, error) {
	provider.requests = append(provider.requests, request)
	return SendResult{ProviderMessageID: "gmail-health-message"}, nil
}

type healthTestStore struct {
	due       []string
	recorded  bool
	healthy   bool
	messageID string
}

func (store *healthTestStore) SyncEmailHealthAccounts(context.Context, []HealthAccountConfig, time.Duration) error {
	return nil
}

func (store *healthTestStore) ClaimDueEmailHealthAccounts(context.Context, []string, time.Duration) ([]string, error) {
	return store.due, nil
}

func (store *healthTestStore) RecordEmailHealthResult(_ context.Context, _ string, healthy bool, messageID, _ string) error {
	store.recorded = true
	store.healthy = healthy
	store.messageID = messageID
	return nil
}

func (store *healthTestStore) ListEmailHealth(context.Context) ([]HealthStatus, error) {
	return nil, nil
}

func TestHealthServiceSendsDueGmailCheck(t *testing.T) {
	provider := &healthTestProvider{}
	store := &healthTestStore{due: []string{"gmail-1"}}
	service := &HealthService{
		emailCfg: config.EmailConfig{DisableSending: true},
		cfg: config.OutreachConfig{
			EmailHealthEnabled:   true,
			EmailHealthRecipient: "rajchodisetti@gmail.com",
			EmailHealthInterval:  24 * time.Hour,
		},
		store: store,
		accounts: map[string]healthAccount{
			"gmail-1": {key: "gmail-1", from: "sales@example.com", provider: provider},
		},
		ordered: []string{"gmail-1"},
	}

	if err := service.RunDue(context.Background()); err != nil {
		t.Fatalf("RunDue() error = %v", err)
	}
	if len(provider.requests) != 1 || provider.requests[0].To != "rajchodisetti@gmail.com" {
		t.Fatalf("requests = %+v, want one health check to configured recipient", provider.requests)
	}
	if !store.recorded || !store.healthy || store.messageID != "gmail-health-message" {
		t.Fatalf("recorded=%v healthy=%v messageID=%q", store.recorded, store.healthy, store.messageID)
	}
}
