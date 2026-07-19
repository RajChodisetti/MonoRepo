package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type HealthAccountConfig struct {
	Key              string
	Provider         string
	ProviderIdentity string
	FromEmail        string
	Enabled          bool
}

type HealthStatus struct {
	AccountKey        string     `json:"account_key"`
	Provider          string     `json:"provider"`
	ProviderIdentity  string     `json:"provider_identity"`
	FromEmail         string     `json:"from_email"`
	Enabled           bool       `json:"enabled"`
	Status            string     `json:"status"`
	LastCheckedAt     *time.Time `json:"last_checked_at,omitempty"`
	NextCheckAt       *time.Time `json:"next_check_at,omitempty"`
	ProviderMessageID string     `json:"provider_message_id,omitempty"`
	LastError         string     `json:"last_error,omitempty"`
}

type HealthStore interface {
	SyncEmailHealthAccounts(ctx context.Context, accounts []HealthAccountConfig, interval time.Duration) error
	ClaimDueEmailHealthAccounts(ctx context.Context, accountKeys []string, interval time.Duration) ([]string, error)
	RecordEmailHealthResult(ctx context.Context, accountKey string, healthy bool, providerMessageID, safeError string) error
	ListEmailHealth(ctx context.Context) ([]HealthStatus, error)
}

type healthAccount struct {
	key      string
	from     string
	provider Provider
}

type HealthService struct {
	emailCfg config.EmailConfig
	cfg      config.OutreachConfig
	store    HealthStore
	accounts map[string]healthAccount
	ordered  []string
}

func NewHealthServiceFromConfig(
	ctx context.Context,
	emailCfg config.EmailConfig,
	outreachCfg config.OutreachConfig,
	store HealthStore,
) (*HealthService, error) {
	service := &HealthService{
		emailCfg: emailCfg,
		cfg:      outreachCfg,
		store:    store,
		accounts: make(map[string]healthAccount, len(outreachCfg.GoogleWorkspaceAccounts)),
	}

	registrations := make([]HealthAccountConfig, 0, len(outreachCfg.GoogleWorkspaceAccounts))
	for index, account := range outreachCfg.GoogleWorkspaceAccounts {
		mailbox, err := canonicalMailbox(account.MailboxEmail)
		if err != nil {
			return nil, fmt.Errorf("Gmail health account %d mailbox: %w", index+1, err)
		}
		from := strings.TrimSpace(account.FromEmail)
		if from == "" {
			from = mailbox
		}
		from, err = canonicalMailbox(from)
		if err != nil {
			return nil, fmt.Errorf("Gmail health account %d from address: %w", index+1, err)
		}
		key := strings.TrimSpace(account.AccountKey)
		if key == "" {
			key = mailbox
		}
		provider, err := NewGmail(emailCfg, account)
		if err != nil {
			return nil, fmt.Errorf("Gmail health account %d: %w", index+1, err)
		}
		service.accounts[key] = healthAccount{key: key, from: from, provider: provider}
		service.ordered = append(service.ordered, key)
		registrations = append(registrations, HealthAccountConfig{
			Key:              key,
			Provider:         "gmail",
			ProviderIdentity: "gmail|" + mailbox,
			FromEmail:        from,
			Enabled:          outreachCfg.EmailHealthEnabled,
		})
	}

	if store != nil {
		if err := store.SyncEmailHealthAccounts(ctx, registrations, outreachCfg.EmailHealthInterval); err != nil {
			return nil, fmt.Errorf("sync Gmail health accounts: %w", err)
		}
	}
	return service, nil
}

func (service *HealthService) RunDue(ctx context.Context) error {
	if service == nil || service.store == nil || len(service.accounts) == 0 {
		return nil
	}
	if !service.cfg.EmailHealthEnabled {
		return nil
	}

	due, err := service.store.ClaimDueEmailHealthAccounts(ctx, service.ordered, service.cfg.EmailHealthInterval)
	if err != nil {
		return fmt.Errorf("claim due Gmail health accounts: %w", err)
	}
	for _, key := range due {
		account, ok := service.accounts[key]
		if !ok {
			continue
		}
		checkedAt := time.Now().UTC()
		result, sendErr := account.provider.Send(ctx, SendRequest{
			To:      service.cfg.EmailHealthRecipient,
			Subject: fmt.Sprintf("[Tuvi] Gmail sender health check — %s", account.from),
			TextBody: fmt.Sprintf(
				"Gmail sender %s passed its Tuvi outreach health check at %s. This mailbox is active and ready for approved restaurant outreach.",
				account.from,
				checkedAt.Format(time.RFC3339),
			),
		})
		if sendErr != nil {
			if recordErr := service.store.RecordEmailHealthResult(ctx, key, false, "", sanitizeHealthError(sendErr)); recordErr != nil {
				return fmt.Errorf("record Gmail health failure for %q: %w", key, recordErr)
			}
			continue
		}
		if err := service.store.RecordEmailHealthResult(ctx, key, true, result.ProviderMessageID, ""); err != nil {
			return fmt.Errorf("record Gmail health success for %q: %w", key, err)
		}
	}
	return nil
}

func (service *HealthService) List(ctx context.Context) ([]HealthStatus, error) {
	if service == nil || service.store == nil {
		return []HealthStatus{}, nil
	}
	return service.store.ListEmailHealth(ctx)
}

func sanitizeHealthError(err error) string {
	message := strings.TrimSpace(err.Error())
	message = redactEmailAddresses(message)
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
