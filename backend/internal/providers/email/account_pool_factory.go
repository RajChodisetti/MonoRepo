package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func NewAccountPoolFromConfig(emailCfg config.EmailConfig, outreachCfg config.OutreachConfig) (*AccountPool, error) {
	return buildAccountPool(context.Background(), emailCfg, outreachCfg, nil)
}

func NewPersistentAccountPoolFromConfig(
	ctx context.Context,
	emailCfg config.EmailConfig,
	outreachCfg config.OutreachConfig,
	quota QuotaStore,
) (*AccountPool, error) {
	return buildAccountPool(ctx, emailCfg, outreachCfg, quota)
}

func buildAccountPool(
	ctx context.Context,
	emailCfg config.EmailConfig,
	outreachCfg config.OutreachConfig,
	quota QuotaStore,
) (*AccountPool, error) {
	// Reconcile abandoned provider-boundary claims even while new sending is
	// disabled. Disabling outreach must not leave legacy `sending` rows looking
	// safely retryable after their delivery lease expires.
	if quota != nil {
		if _, err := quota.ReconcileStaleEmailDeliveries(ctx); err != nil {
			return nil, fmt.Errorf("reconcile stale outreach email deliveries: %w", err)
		}
	}
	if redirect := strings.TrimSpace(emailCfg.RedirectTo); redirect != "" {
		if _, err := canonicalMailbox(redirect); err != nil {
			return nil, fmt.Errorf("outreach redirect recipient: %w", err)
		}
	}

	accountCount := len(outreachCfg.GoogleWorkspaceAccounts)
	if accountCount == 0 {
		return nil, fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON must contain at least one Gmail outreach account")
	}

	keyedProviders := make([]accountProvider, 0, accountCount)
	registrations := make([]QuotaAccountConfig, 0, accountCount)
	seenKeys := make(map[string]struct{}, accountCount)
	seenIdentities := make(map[string]struct{}, accountCount)
	for index, account := range outreachCfg.GoogleWorkspaceAccounts {
		mailboxEmail, err := canonicalMailbox(account.MailboxEmail)
		if err != nil {
			return nil, fmt.Errorf("outreach Google Workspace account %d mailbox_email: %w", index+1, err)
		}
		fromEmail := strings.TrimSpace(account.FromEmail)
		if fromEmail == "" {
			fromEmail = mailboxEmail
		}
		fromEmail, err = canonicalMailbox(fromEmail)
		if err != nil {
			return nil, fmt.Errorf("outreach Google Workspace account %d from_email: %w", index+1, err)
		}
		account.MailboxEmail = mailboxEmail
		account.FromEmail = fromEmail
		provider, err := NewGmail(emailCfg, account)
		if err != nil {
			return nil, fmt.Errorf("outreach Google Workspace account %d: %w", index+1, err)
		}
		accountKey := strings.TrimSpace(account.AccountKey)
		if accountKey == "" {
			accountKey = mailboxEmail
		}
		if accountKey == "" {
			return nil, fmt.Errorf("outreach Google Workspace account %d requires a stable key or mailbox_email", index+1)
		}
		if _, exists := seenKeys[accountKey]; exists {
			return nil, fmt.Errorf("outreach Google Workspace account %d duplicates stable key %q", index+1, accountKey)
		}
		seenKeys[accountKey] = struct{}{}
		identity := "gmail|" + mailboxEmail
		if _, exists := seenIdentities[identity]; exists {
			return nil, fmt.Errorf("outreach Google Workspace account %d duplicates mailbox %q", index+1, mailboxEmail)
		}
		seenIdentities[identity] = struct{}{}
		keyedProviders = append(keyedProviders, accountProvider{key: accountKey, provider: provider})
		registrations = append(registrations, QuotaAccountConfig{
			Key:              accountKey,
			Provider:         "gmail",
			ProviderIdentity: identity,
			FromEmail:        fromEmail,
			Position:         len(registrations),
			SendLimit:        outreachCfg.EmailsPerAccount,
			SendWindow:       outreachCfg.SendWindow,
			SendJitterMin:    outreachCfg.SendJitterMin,
			SendJitterMax:    outreachCfg.SendJitterMax,
		})
	}

	maxTotal := outreachCfg.BulkMax
	if maxTotal <= 0 {
		maxTotal = accountCount * outreachCfg.EmailsPerAccount
	}
	if quota == nil {
		pool, err := newAccountPoolProviders(keyedProviders, outreachCfg.EmailsPerAccount, maxTotal)
		if err != nil {
			return nil, err
		}
		attachReplyTo(pool, outreachCfg)
		return pool, nil
	}

	cooldown := outreachCfg.AccountCooldown
	if cooldown <= 0 {
		cooldown = 24 * time.Hour
	}
	if err := quota.SyncEmailAccounts(ctx, registrations, cooldown); err != nil {
		return nil, fmt.Errorf("sync outreach email accounts: %w", err)
	}
	pool, err := newPersistentAccountPool(keyedProviders, outreachCfg.EmailsPerAccount, maxTotal, cooldown, quota)
	if err != nil {
		return nil, err
	}
	attachReplyTo(pool, outreachCfg)
	return pool, nil
}

func attachReplyTo(pool *AccountPool, outreachCfg config.OutreachConfig) {
	if pool == nil {
		return
	}
	localPart := outreachCfg.InboundLocalPart
	domain := outreachCfg.InboundDomain
	if localPart == "" || domain == "" {
		return
	}
	pool.replyToForAttempt = func(attemptID uuid.UUID) string {
		return ReplyToAddress(localPart, domain, attemptID)
	}
}
