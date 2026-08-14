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
	attemptID     uuid.UUID
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
	if store.attemptID == uuid.Nil {
		store.attemptID = uuid.New()
	}
	return DeliveryClaim{AttemptID: store.attemptID, AccountKey: "account-1"}, nil
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
	last  SendRequest
}

func (provider *preflightProvider) Send(_ context.Context, req SendRequest) (SendResult, error) {
	provider.sends++
	provider.last = req
	return SendResult{ProviderMessageID: "message-1", ReplyTo: req.ReplyTo}, nil
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

func TestAccountPoolSendDirectFromUsesRequestedMailbox(t *testing.T) {
	first := &preflightProvider{}
	second := &preflightProvider{}
	pool, err := newAccountPoolProviders(
		[]accountProvider{
			{key: "first", provider: first},
			{key: "reply-mailbox", provider: second},
		},
		40,
		80,
	)
	if err != nil {
		t.Fatalf("newAccountPoolProviders() error = %v", err)
	}

	result, err := pool.SendDirectFrom(context.Background(), "reply-mailbox", SendRequest{
		To:       "owner@example.com",
		Subject:  "Re: hello",
		TextBody: "Thanks",
	})
	if err != nil {
		t.Fatalf("SendDirectFrom() error = %v", err)
	}
	if first.sends != 0 || second.sends != 1 {
		t.Fatalf("provider sends = %d/%d, want 0/1", first.sends, second.sends)
	}
	if result.AccountKey != "reply-mailbox" {
		t.Fatalf("AccountKey = %q, want reply-mailbox", result.AccountKey)
	}
}

func TestAccountPoolSendDirectFromSupportsDirectOnlyInbox(t *testing.T) {
	bulk := &preflightProvider{}
	inbox := &preflightProvider{}
	pool, err := newAccountPoolProviders(
		[]accountProvider{{key: "bulk", provider: bulk}},
		40,
		40,
	)
	if err != nil {
		t.Fatalf("newAccountPoolProviders() error = %v", err)
	}
	if err := pool.addDirectAccount(accountProvider{key: "inbound", provider: inbox}); err != nil {
		t.Fatalf("addDirectAccount() error = %v", err)
	}

	result, err := pool.SendDirectFrom(context.Background(), "inbound", SendRequest{
		To:       "owner@example.com",
		Subject:  "Re: hello",
		TextBody: "Thanks",
	})
	if err != nil {
		t.Fatalf("SendDirectFrom() error = %v", err)
	}
	if bulk.sends != 0 || inbox.sends != 1 || result.AccountKey != "inbound" {
		t.Fatalf("provider sends/account = %d/%d/%q", bulk.sends, inbox.sends, result.AccountKey)
	}
	if len(pool.accounts) != 1 {
		t.Fatalf("quota-managed accounts = %d, want direct-only inbox excluded", len(pool.accounts))
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

func TestDurableAccountPoolSetsPlusAddressReplyToAfterClaim(t *testing.T) {
	quota := &preflightQuotaStore{attemptID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")}
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
	pool.replyToForAttempt = func(id uuid.UUID) string {
		return ReplyToAddress("outreach", "tuvisolutions.com", id)
	}

	result, err := pool.Send(context.Background(), SendRequest{
		To:       "owner@example.com",
		Subject:  "Your restaurant demo",
		TextBody: "hello",
		Delivery: &DeliveryContext{
			CampaignID:   uuid.New(),
			RestaurantID: uuid.New(),
		},
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	want := "outreach+aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa@tuvisolutions.com"
	if provider.last.ReplyTo != want {
		t.Fatalf("provider ReplyTo = %q, want %q", provider.last.ReplyTo, want)
	}
	if result.ReplyTo != want {
		t.Fatalf("result ReplyTo = %q, want %q", result.ReplyTo, want)
	}
	if result.DeliveryAttemptID != quota.attemptID {
		t.Fatalf("DeliveryAttemptID = %s, want %s", result.DeliveryAttemptID, quota.attemptID)
	}
}
