package email

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type QuotaAccountConfig struct {
	Key              string
	Provider         string
	ProviderIdentity string
	FromEmail        string
	Position         int
	SendLimit        int
	SendWindow       time.Duration
	SendJitterMin    time.Duration
	SendJitterMax    time.Duration
}

type DeliveryContext struct {
	CampaignID                  uuid.UUID
	RestaurantID                uuid.UUID
	BulkJobID                   *uuid.UUID
	Step                        int
	Recipient                   string
	CampaignArtifactFingerprint string
}

type DeliveryClaim struct {
	AttemptID       uuid.UUID
	SendSequence    int64
	CampaignStep    int
	AccountKey      string
	AccountCycle    int64
	AccountSequence int
}

type QuotaStore interface {
	SyncEmailAccounts(ctx context.Context, accounts []QuotaAccountConfig, cooldown time.Duration) error
	ReconcileStaleEmailDeliveries(ctx context.Context) (int, error)
	ClaimEmailDelivery(
		ctx context.Context,
		accountKeys []string,
		delivery DeliveryContext,
		cooldown time.Duration,
	) (DeliveryClaim, error)
	CompleteEmailDelivery(ctx context.Context, claim DeliveryClaim, providerMessageID string) error
	SkipEmailDelivery(ctx context.Context, claim DeliveryClaim, skipped, redirected bool) error
	MarkEmailDeliveryUnknown(ctx context.Context, claim DeliveryClaim, errorCode string) error
	NextEmailAccountAvailableAt(ctx context.Context, accountKeys []string) (*time.Time, error)
}
