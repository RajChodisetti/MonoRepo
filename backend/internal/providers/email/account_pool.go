package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrAccountsExhausted = errors.New("all outreach email accounts are unavailable")

type accountProvider struct {
	key      string
	provider Provider
}

type AccountPool struct {
	accounts          []accountProvider
	limitPerAccount   int
	currentIndex      int
	sentOnCurrent     int
	totalSent         int
	maxTotal          int
	quota             QuotaStore
	cooldown          time.Duration
	replyToForAttempt func(uuid.UUID) string
}

// NewAccountPool retains the in-memory constructor for isolated/local callers.
// Production outreach is wired through NewPersistentAccountPoolFromConfig.
func NewAccountPool(providers []Provider, limitPerAccount, maxTotal int) (*AccountPool, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("account pool requires at least one email provider")
	}
	if limitPerAccount < 1 {
		return nil, fmt.Errorf("limit per account must be at least 1")
	}
	if maxTotal < 1 {
		return nil, fmt.Errorf("max total sends must be at least 1")
	}
	accounts := make([]accountProvider, 0, len(providers))
	for index, provider := range providers {
		accounts = append(accounts, accountProvider{
			key:      fmt.Sprintf("in-memory-%d", index+1),
			provider: provider,
		})
	}
	return &AccountPool{
		accounts:        accounts,
		limitPerAccount: limitPerAccount,
		maxTotal:        maxTotal,
	}, nil
}

func newPersistentAccountPool(
	accounts []accountProvider,
	limitPerAccount int,
	maxTotal int,
	cooldown time.Duration,
	quota QuotaStore,
) (*AccountPool, error) {
	if quota == nil {
		return nil, fmt.Errorf("persistent account pool requires a quota store")
	}
	if cooldown <= 0 {
		cooldown = 24 * time.Hour
	}
	pool, err := newAccountPoolProviders(accounts, limitPerAccount, maxTotal)
	if err != nil {
		return nil, err
	}
	pool.quota = quota
	pool.cooldown = cooldown
	return pool, nil
}

func newAccountPoolProviders(accounts []accountProvider, limitPerAccount, maxTotal int) (*AccountPool, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("account pool requires at least one email provider")
	}
	if limitPerAccount < 1 {
		return nil, fmt.Errorf("limit per account must be at least 1")
	}
	if maxTotal < 1 {
		return nil, fmt.Errorf("max total sends must be at least 1")
	}
	return &AccountPool{
		accounts:        accounts,
		limitPerAccount: limitPerAccount,
		maxTotal:        maxTotal,
	}, nil
}

func (pool *AccountPool) Durable() bool {
	return pool != nil && pool.quota != nil
}

func (pool *AccountPool) NextAvailableAt(ctx context.Context) (*time.Time, error) {
	if pool == nil || !pool.Durable() {
		return nil, nil
	}
	if _, err := pool.quota.ReconcileStaleEmailDeliveries(ctx); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(pool.accounts))
	for _, account := range pool.accounts {
		keys = append(keys, account.key)
	}
	return pool.quota.NextEmailAccountAvailableAt(ctx, keys)
}

// Exhausted is meaningful only for the legacy in-memory pool. Durable pools
// discover availability atomically while claiming a PostgreSQL quota slot.
func (pool *AccountPool) Exhausted() bool {
	if pool == nil {
		return true
	}
	if pool.Durable() {
		return false
	}
	return pool.currentIndex >= len(pool.accounts) || pool.totalSent >= pool.maxTotal
}

func (pool *AccountPool) TotalSent() int {
	return pool.totalSent
}

func (pool *AccountPool) CurrentAccountIndex() int {
	if pool.currentIndex >= len(pool.accounts) {
		return len(pool.accounts)
	}
	return pool.currentIndex + 1
}

// Reset is retained for source compatibility with legacy/local pools. Durable
// pools intentionally ignore it because their allowance is PostgreSQL-backed.
func (pool *AccountPool) Reset() {
	if pool == nil || pool.Durable() {
		return
	}
	pool.currentIndex = 0
	pool.sentOnCurrent = 0
	pool.totalSent = 0
}

func (pool *AccountPool) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if pool == nil {
		return SendResult{}, ErrAccountsExhausted
	}
	if pool.Durable() {
		return pool.sendDurable(ctx, req)
	}
	return pool.sendInMemory(ctx, req)
}

// SendDirect uses the configured account providers without the durable quota
// claim. It is intended for bounded, internal-admin manual sends where the
// caller owns the preview/confirmation and restaurant contact recording.
func (pool *AccountPool) SendDirect(ctx context.Context, req SendRequest) (SendResult, error) {
	if pool == nil || len(pool.accounts) == 0 {
		return SendResult{}, ErrAccountsExhausted
	}
	if pool.currentIndex >= len(pool.accounts) {
		pool.currentIndex = 0
	}
	account := pool.accounts[pool.currentIndex]
	pool.currentIndex = (pool.currentIndex + 1) % len(pool.accounts)
	result, err := account.provider.Send(ctx, req)
	if result.AccountKey == "" {
		result.AccountKey = account.key
	}
	return result, err
}

// SendDirectFrom sends a bounded manual message through one explicitly selected
// configured account. Inbox replies remain in the mailbox that captured the
// thread instead of rotating to another outreach identity.
func (pool *AccountPool) SendDirectFrom(ctx context.Context, accountKey string, req SendRequest) (SendResult, error) {
	if pool == nil {
		return SendResult{}, ErrAccountsExhausted
	}
	accountKey = strings.TrimSpace(accountKey)
	for _, account := range pool.accounts {
		if account.key != accountKey {
			continue
		}
		result, err := account.provider.Send(ctx, req)
		if result.AccountKey == "" {
			result.AccountKey = account.key
		}
		return result, err
	}
	return SendResult{}, fmt.Errorf("outreach email account %q is not configured", accountKey)
}

func (pool *AccountPool) sendInMemory(ctx context.Context, req SendRequest) (SendResult, error) {
	if pool.Exhausted() {
		return SendResult{}, ErrAccountsExhausted
	}

	for pool.currentIndex < len(pool.accounts) {
		if pool.sentOnCurrent >= pool.limitPerAccount {
			pool.currentIndex++
			pool.sentOnCurrent = 0
			continue
		}
		if pool.totalSent >= pool.maxTotal {
			return SendResult{}, ErrAccountsExhausted
		}

		result, err := pool.accounts[pool.currentIndex].provider.Send(ctx, req)
		if err != nil {
			return SendResult{}, err
		}

		pool.sentOnCurrent++
		pool.totalSent++
		return result, nil
	}

	return SendResult{}, ErrAccountsExhausted
}

func (pool *AccountPool) sendDurable(ctx context.Context, req SendRequest) (SendResult, error) {
	managedResult := SendResult{QuotaManaged: true}
	if req.Delivery == nil {
		return managedResult, fmt.Errorf("quota-managed outreach send requires delivery context")
	}
	recipient, err := validateQuotaManagedSendRequest(req)
	if err != nil {
		return managedResult, err
	}

	keys := make([]string, 0, len(pool.accounts))
	providers := make(map[string]Provider, len(pool.accounts))
	for _, account := range pool.accounts {
		keys = append(keys, account.key)
		providers[account.key] = account.provider
	}

	delivery := *req.Delivery
	// Bind the database claim to the actual recipient passed to the provider;
	// callers cannot accidentally validate one address and send to another.
	delivery.Recipient = recipient
	claim, err := pool.quota.ClaimEmailDelivery(ctx, keys, delivery, pool.cooldown)
	if err != nil {
		return managedResult, err
	}
	if strings.TrimSpace(req.ReplyTo) == "" && pool.replyToForAttempt != nil {
		if replyTo := pool.replyToForAttempt(claim.AttemptID); replyTo != "" {
			req.ReplyTo = replyTo
		}
	}

	result := SendResult{
		QuotaManaged:      true,
		DeliveryAttemptID: claim.AttemptID,
		SendSequence:      claim.SendSequence,
		AccountKey:        claim.AccountKey,
		AccountCycle:      claim.AccountCycle,
		AccountSequence:   claim.AccountSequence,
	}
	provider, ok := providers[claim.AccountKey]
	if !ok {
		finalizeErr := pool.markUnknown(ctx, claim, "provider_configuration_missing")
		if finalizeErr != nil {
			return result, fmt.Errorf("email account %q is not configured; record unknown delivery: %w", claim.AccountKey, finalizeErr)
		}
		result.Finalized = true
		return result, fmt.Errorf("email account %q is not configured", claim.AccountKey)
	}

	providerResult, sendErr := provider.Send(ctx, req)
	result.ProviderMessageID = providerResult.ProviderMessageID
	result.ProviderThreadID = providerResult.ProviderThreadID
	result.RFCMessageID = providerResult.RFCMessageID
	result.FromEmail = providerResult.FromEmail
	result.ReplyTo = providerResult.ReplyTo
	if result.ReplyTo == "" {
		result.ReplyTo = strings.TrimSpace(req.ReplyTo)
	}
	result.RedirectedTo = providerResult.RedirectedTo
	result.Skipped = providerResult.Skipped

	if sendErr != nil {
		finalizeErr := pool.markUnknown(ctx, claim, "provider_send_unknown")
		if finalizeErr != nil {
			return result, fmt.Errorf("provider send failed: %v; record unknown delivery: %w", sendErr, finalizeErr)
		}
		result.Finalized = true
		return result, sendErr
	}

	finalizeCtx, cancel := durableFinalizeContext(ctx)
	defer cancel()
	if result.Skipped || result.RedirectedTo != "" {
		if err := pool.quota.SkipEmailDelivery(finalizeCtx, claim, result.Skipped, result.RedirectedTo != ""); err != nil {
			if unknownErr := pool.markUnknown(ctx, claim, "delivery_finalization_unknown"); unknownErr != nil {
				return result, fmt.Errorf("record skipped outreach delivery: %v; record unknown delivery: %w", err, unknownErr)
			}
			result.Finalized = true
			return result, fmt.Errorf("record skipped outreach delivery: %w", err)
		}
		result.Finalized = true
		return result, nil
	}

	if err := pool.quota.CompleteEmailDelivery(finalizeCtx, claim, result.ProviderMessageID); err != nil {
		if unknownErr := pool.markUnknown(ctx, claim, "delivery_finalization_unknown"); unknownErr != nil {
			return result, fmt.Errorf("finalize accepted outreach delivery: %v; record unknown delivery: %w", err, unknownErr)
		}
		result.Finalized = true
		return result, fmt.Errorf("finalize accepted outreach delivery: %w", err)
	}
	result.Finalized = true
	return result, nil
}

func validateQuotaManagedSendRequest(req SendRequest) (string, error) {
	recipient, err := canonicalMailbox(req.To)
	if err != nil {
		return "", fmt.Errorf("quota-managed email recipient: %w", err)
	}
	if _, err := cleanHeaderValue(req.Subject, "subject"); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.ReplyTo) != "" {
		if _, err := canonicalMailbox(req.ReplyTo); err != nil {
			return "", fmt.Errorf("quota-managed email reply-to: %w", err)
		}
	}
	if strings.TrimSpace(req.HTMLBody) == "" && strings.TrimSpace(req.TextBody) == "" {
		return "", fmt.Errorf("quota-managed email requires an HTML or text body")
	}
	return recipient, nil
}

func (pool *AccountPool) markUnknown(ctx context.Context, claim DeliveryClaim, code string) error {
	finalizeCtx, cancel := durableFinalizeContext(ctx)
	defer cancel()
	return pool.quota.MarkEmailDeliveryUnknown(finalizeCtx, claim, code)
}

func durableFinalizeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
}

var _ Provider = (*AccountPool)(nil)
