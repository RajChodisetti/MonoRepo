package email

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type mutableOutreachConfigLoader struct {
	cfg config.OutreachConfig
}

func (loader *mutableOutreachConfigLoader) Load(context.Context) (config.OutreachConfig, error) {
	return loader.cfg, nil
}

type reloadingQuotaStub struct{}

func (reloadingQuotaStub) SyncEmailAccounts(context.Context, []QuotaAccountConfig, time.Duration) error {
	return nil
}
func (reloadingQuotaStub) ReconcileStaleEmailDeliveries(context.Context) (int, error) {
	return 0, nil
}
func (reloadingQuotaStub) ClaimEmailDelivery(context.Context, []string, DeliveryContext, time.Duration) (DeliveryClaim, error) {
	return DeliveryClaim{}, nil
}
func (reloadingQuotaStub) CompleteEmailDelivery(context.Context, DeliveryClaim, string) error {
	return nil
}
func (reloadingQuotaStub) SkipEmailDelivery(context.Context, DeliveryClaim, bool, bool) error {
	return nil
}
func (reloadingQuotaStub) MarkEmailDeliveryUnknown(context.Context, DeliveryClaim, string) error {
	return nil
}
func (reloadingQuotaStub) NextEmailAccountAvailableAt(context.Context, []string) (*time.Time, error) {
	return nil, nil
}

type quotaTouchCounter struct {
	syncs      int
	reconciles int
	claims     int
	completes  int
	skips      int
	unknowns   int
	reads      int
	claim      DeliveryClaim
}

func (store *quotaTouchCounter) SyncEmailAccounts(context.Context, []QuotaAccountConfig, time.Duration) error {
	store.syncs++
	return nil
}
func (store *quotaTouchCounter) ReconcileStaleEmailDeliveries(context.Context) (int, error) {
	store.reconciles++
	return 0, nil
}
func (store *quotaTouchCounter) ClaimEmailDelivery(context.Context, []string, DeliveryContext, time.Duration) (DeliveryClaim, error) {
	store.claims++
	return store.claim, nil
}
func (store *quotaTouchCounter) CompleteEmailDelivery(context.Context, DeliveryClaim, string) error {
	store.completes++
	return nil
}
func (store *quotaTouchCounter) SkipEmailDelivery(context.Context, DeliveryClaim, bool, bool) error {
	store.skips++
	return nil
}
func (store *quotaTouchCounter) MarkEmailDeliveryUnknown(context.Context, DeliveryClaim, string) error {
	store.unknowns++
	return nil
}
func (store *quotaTouchCounter) NextEmailAccountAvailableAt(context.Context, []string) (*time.Time, error) {
	store.reads++
	return nil, nil
}

func (store *quotaTouchCounter) total() int {
	return store.syncs + store.reconciles + store.claims + store.completes + store.skips + store.unknowns + store.reads
}

func reloadingTestConfig() config.OutreachConfig {
	return config.OutreachConfig{
		BulkMax:          40,
		EmailsPerAccount: 40,
		GoogleWorkspaceAccounts: []config.GmailMailConfig{{
			AccountKey:   "manual",
			MailboxEmail: "manual@example.com",
			FromEmail:    "manual@example.com",
			ClientID:     "client",
			ClientSecret: "secret",
			RefreshToken: uuid.NewString(),
		}},
	}
}

func TestReloadingAccountPoolSeesAccountsAddedAfterConstruction(t *testing.T) {
	loader := &mutableOutreachConfigLoader{}
	pool := NewReloadingPersistentAccountPool(config.EmailConfig{}, loader, reloadingQuotaStub{})

	configured, err := pool.Configured(context.Background())
	if err != nil || configured {
		t.Fatalf("Configured() = %v, %v; want false, nil", configured, err)
	}
	loader.cfg.GoogleWorkspaceAccounts = []config.GmailMailConfig{{
		AccountKey: "added", MailboxEmail: "added@example.com", FromEmail: "added@example.com",
		ClientID: "client", ClientSecret: "secret", RefreshToken: uuid.NewString(),
	}}

	configured, err = pool.Configured(context.Background())
	if err != nil || !configured {
		t.Fatalf("Configured() after add = %v, %v; want true, nil", configured, err)
	}
}

func TestReloadingAccountPoolDirectSendsDoNotTouchScheduledQuota(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		send func(*ReloadingAccountPool) error
	}{
		{
			name: "rotating manual send",
			send: func(pool *ReloadingAccountPool) error {
				_, err := pool.SendDirect(context.Background(), SendRequest{})
				return err
			},
		},
		{
			name: "same mailbox reply",
			send: func(pool *ReloadingAccountPool) error {
				_, err := pool.SendDirectFrom(context.Background(), "manual", SendRequest{})
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			loader := &mutableOutreachConfigLoader{cfg: reloadingTestConfig()}
			quota := &quotaTouchCounter{}
			pool := NewReloadingPersistentAccountPool(config.EmailConfig{}, loader, quota)

			if err := test.send(pool); err == nil {
				t.Fatal("manual send error = nil, want local recipient validation error")
			}
			if quota.total() != 0 {
				t.Fatalf("scheduled quota method calls = %d, want 0; store = %#v", quota.total(), quota)
			}
		})
	}
}

func TestReloadingAccountPoolScheduledSendClaimsDurableQuota(t *testing.T) {
	t.Parallel()

	loader := &mutableOutreachConfigLoader{cfg: reloadingTestConfig()}
	attemptID := uuid.New()
	quota := &quotaTouchCounter{claim: DeliveryClaim{
		AttemptID:  attemptID,
		AccountKey: "missing-after-claim",
	}}
	pool := NewReloadingPersistentAccountPool(config.EmailConfig{}, loader, quota)

	result, err := pool.Send(context.Background(), SendRequest{
		To:       "lead@example.com",
		Subject:  "Scheduled outreach",
		TextBody: "Scheduled body",
		Delivery: &DeliveryContext{},
	})
	if err == nil {
		t.Fatal("scheduled send error = nil, want missing-provider error after durable claim")
	}
	if !result.QuotaManaged || result.DeliveryAttemptID != attemptID {
		t.Fatalf("scheduled result = %#v, want quota-managed claimed attempt %s", result, attemptID)
	}
	if quota.reconciles != 1 || quota.syncs != 1 || quota.claims != 1 {
		t.Fatalf(
			"scheduled quota calls = reconciles %d, syncs %d, claims %d; want 1 each",
			quota.reconciles,
			quota.syncs,
			quota.claims,
		)
	}
	if quota.unknowns != 1 || quota.completes != 0 || quota.skips != 0 {
		t.Fatalf("scheduled quota finalization = %#v, want one unknown after claimed provider mismatch", quota)
	}
}
