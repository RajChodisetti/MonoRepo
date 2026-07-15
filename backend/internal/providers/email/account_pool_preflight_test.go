package email

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type preflightQuotaStore struct {
	claims int
}

func (store *preflightQuotaStore) SyncEmailAccounts(context.Context, []QuotaAccountConfig, time.Duration) error {
	return nil
}

func (store *preflightQuotaStore) ReconcileStaleEmailDeliveries(context.Context) (int, error) {
	return 0, nil
}

func (store *preflightQuotaStore) ClaimEmailDelivery(context.Context, []string, DeliveryContext, time.Duration) (DeliveryClaim, error) {
	store.claims++
	return DeliveryClaim{AttemptID: uuid.New(), AccountKey: "account-1"}, nil
}

func (store *preflightQuotaStore) CompleteEmailDelivery(context.Context, DeliveryClaim, string) error {
	return nil
}

func (store *preflightQuotaStore) SkipEmailDelivery(context.Context, DeliveryClaim, bool, bool) error {
	return nil
}

func (store *preflightQuotaStore) MarkEmailDeliveryUnknown(context.Context, DeliveryClaim, string) error {
	return nil
}

func (store *preflightQuotaStore) NextEmailAccountAvailableAt(context.Context, []string) (*time.Time, error) {
	return nil, nil
}

type preflightProvider struct {
	sends int
}

func (provider *preflightProvider) Send(context.Context, SendRequest) (SendResult, error) {
	provider.sends++
	return SendResult{ProviderMessageID: "message-1"}, nil
}

func TestDurableAccountPoolValidatesBeforeClaim(t *testing.T) {
	quota := &preflightQuotaStore{}
	provider := &preflightProvider{}
	pool, err := newPersistentAccountPool(
		[]accountProvider{{key: "account-1", provider: provider}},
		40,
		40,
		24*time.Hour,
		quota,
	)
	if err != nil {
		t.Fatalf("newPersistentAccountPool() error = %v", err)
	}

	result, err := pool.Send(context.Background(), SendRequest{
		To:       "owner@example.com",
		Subject:  "safe\r\nBcc: attacker@example.com",
		TextBody: "hello",
		Delivery: &DeliveryContext{
			CampaignID:   uuid.New(),
			RestaurantID: uuid.New(),
		},
	})
	if err == nil || !strings.Contains(err.Error(), "newline") {
		t.Fatalf("Send() error = %v, want newline rejection", err)
	}
	if !result.QuotaManaged {
		t.Fatal("Send() result was not marked quota managed")
	}
	if quota.claims != 0 {
		t.Fatalf("quota claims = %d, want 0", quota.claims)
	}
	if provider.sends != 0 {
		t.Fatalf("provider sends = %d, want 0", provider.sends)
	}
}
