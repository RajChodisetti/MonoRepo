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
