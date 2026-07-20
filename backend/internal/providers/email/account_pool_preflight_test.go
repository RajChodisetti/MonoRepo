package email

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type preflightQuotaStore struct {
	claims        int
	registrations []QuotaAccountConfig
}

func (store *preflightQuotaStore) SyncEmailAccounts(_ context.Context, accounts []QuotaAccountConfig, _ time.Duration) error {
	store.registrations = append([]QuotaAccountConfig(nil), accounts...)
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

func TestDurableAccountPoolSendDirectSkipsQuotaClaim(t *testing.T) {
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

	result, err := pool.SendDirect(context.Background(), SendRequest{
		To:       "owner@example.com",
		Subject:  "Manual send",
		TextBody: "hello",
	})
	if err != nil {
		t.Fatalf("SendDirect() error = %v, want nil", err)
	}
	if result.QuotaManaged {
		t.Fatal("SendDirect() result was quota managed")
	}
	if quota.claims != 0 {
		t.Fatalf("quota claims = %d, want 0", quota.claims)
	}
	if provider.sends != 1 {
		t.Fatalf("provider sends = %d, want 1", provider.sends)
	}
}

func TestPersistentAccountPoolRegistersDurablePacingPolicy(t *testing.T) {
	t.Parallel()

	quota := &preflightQuotaStore{}
	pool, err := buildAccountPool(
		context.Background(),
		config.EmailConfig{},
		config.OutreachConfig{
			BulkMax:          40,
			EmailsPerAccount: 40,
			SendWindow:       8 * time.Hour,
			SendJitterMin:    2 * time.Minute,
			SendJitterMax:    5 * time.Minute,
			AccountCooldown:  24 * time.Hour,
			GoogleWorkspaceAccounts: []config.GmailMailConfig{{
				AccountKey:   "workspace-sales-1",
				MailboxEmail: "sales1@example.com",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RefreshToken: "refresh-token",
			}},
		},
		quota,
	)
	if err != nil {
		t.Fatalf("buildAccountPool() error = %v", err)
	}
	if pool == nil || !pool.Durable() {
		t.Fatal("buildAccountPool() did not return a durable pool")
	}
	if len(quota.registrations) != 1 {
		t.Fatalf("registrations = %d, want 1", len(quota.registrations))
	}
	registration := quota.registrations[0]
	if registration.SendLimit != 40 || registration.SendWindow != 8*time.Hour || registration.SendJitterMin != 2*time.Minute || registration.SendJitterMax != 5*time.Minute {
		t.Fatalf("registration pacing = %#v", registration)
	}
}
