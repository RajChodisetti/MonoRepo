package outreachaccounts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var accountKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

type Service struct {
	repo  Repository
	base  config.OutreachConfig
	vault *credentialCipher
	log   *slog.Logger
}

func NewService(repo Repository, base config.OutreachConfig, encryptionKey string, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	vault, err := newCredentialCipher(encryptionKey)
	if err != nil && !errors.Is(err, ErrEncryptionUnavailable) {
		log.Error("outreach_credential_encryption_unavailable", "error", err)
	}
	return &Service{repo: repo, base: base, vault: vault, log: log}
}

func (service *Service) List(ctx context.Context, principal auth.Principal) (ListResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return ListResult{}, ErrForbidden
	}
	stored, err := service.listStored(ctx)
	if err != nil {
		return ListResult{}, err
	}
	resolved := service.resolveStored(ctx, stored)
	accounts := make([]Account, 0, len(service.environmentAccounts())+len(stored))
	used := make(map[uuid.UUID]struct{}, len(stored))
	for _, environment := range service.environmentAccounts() {
		if record, found := findStoredConflict(stored, environment); found {
			used[record.ID] = struct{}{}
			_, effective := resolved[record.ID]
			accounts = append(accounts, storedAccountView(record, effective, true))
			continue
		}
		accounts = append(accounts, environmentAccountView(environment))
	}
	for _, record := range stored {
		if _, exists := used[record.ID]; exists {
			continue
		}
		_, effective := resolved[record.ID]
		accounts = append(accounts, storedAccountView(record, effective, false))
	}
	return ListResult{Accounts: accounts, EncryptionReady: service.vault != nil}, nil
}

func (service *Service) Create(ctx context.Context, principal auth.Principal, input CreateInput) (Account, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Account{}, ErrForbidden
	}
	if service.vault == nil {
		return Account{}, ErrEncryptionUnavailable
	}
	accountKey := strings.ToLower(strings.TrimSpace(input.AccountKey))
	if !accountKeyPattern.MatchString(accountKey) {
		return Account{}, fmt.Errorf("%w: account_key must contain 2-64 lowercase letters, numbers, underscores, or hyphens", ErrInvalid)
	}
	mailbox, err := canonicalMailbox(input.MailboxEmail)
	if err != nil {
		return Account{}, fmt.Errorf("%w: mailbox_email %v", ErrInvalid, err)
	}
	fromEmail := strings.TrimSpace(input.FromEmail)
	if fromEmail == "" {
		fromEmail = mailbox
	}
	fromEmail, err = canonicalMailbox(fromEmail)
	if err != nil {
		return Account{}, fmt.Errorf("%w: from_email %v", ErrInvalid, err)
	}
	payload := credentialPayload{
		ClientID: strings.TrimSpace(input.Credentials.ClientID), ClientSecret: strings.TrimSpace(input.Credentials.ClientSecret), RefreshToken: strings.TrimSpace(input.Credentials.RefreshToken),
	}
	if err := validateCredentialPayload(payload); err != nil {
		return Account{}, err
	}
	overridesEnvironment, err := service.validateEnvironmentIdentity(accountKey, mailbox)
	if err != nil {
		return Account{}, err
	}
	ciphertext, err := service.vault.encrypt(accountKey, mailbox, payload)
	if err != nil {
		return Account{}, err
	}
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	actor := principal.UserID
	created, err := service.repo.Create(ctx, StoredAccount{
		AccountKey: accountKey, MailboxEmail: mailbox, FromEmail: fromEmail,
		CredentialCiphertext: ciphertext, EncryptionVersion: 1, Enabled: enabled,
		CreatedBy: &actor, UpdatedBy: &actor,
	})
	if err != nil {
		return Account{}, err
	}
	return storedAccountView(created, created.Enabled, overridesEnvironment), nil
}

func (service *Service) Update(ctx context.Context, principal auth.Principal, id uuid.UUID, input UpdateInput) (Account, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Account{}, ErrForbidden
	}
	if input.Enabled == nil && input.FromEmail == nil && input.Credentials == nil {
		return Account{}, fmt.Errorf("%w: at least one editable field is required", ErrInvalid)
	}
	record, err := service.repo.Get(ctx, id)
	if err != nil {
		return Account{}, err
	}
	wasEnabled := record.Enabled
	if input.FromEmail != nil {
		record.FromEmail, err = canonicalMailbox(*input.FromEmail)
		if err != nil {
			return Account{}, fmt.Errorf("%w: from_email %v", ErrInvalid, err)
		}
		record.clearAuthQuarantine = true
	}
	if input.Enabled != nil {
		record.Enabled = *input.Enabled
	}
	if input.Credentials != nil {
		if service.vault == nil {
			return Account{}, ErrEncryptionUnavailable
		}
		payload := credentialPayload{
			ClientID: strings.TrimSpace(input.Credentials.ClientID), ClientSecret: strings.TrimSpace(input.Credentials.ClientSecret), RefreshToken: strings.TrimSpace(input.Credentials.RefreshToken),
		}
		if err := validateCredentialPayload(payload); err != nil {
			return Account{}, err
		}
		record.CredentialCiphertext, err = service.vault.encrypt(record.AccountKey, record.MailboxEmail, payload)
		if err != nil {
			return Account{}, err
		}
		record.clearAuthQuarantine = true
	}
	if input.Enabled != nil && !wasEnabled && record.Enabled {
		record.clearAuthQuarantine = true
	}
	actor := principal.UserID
	record.UpdatedBy = &actor
	updated, err := service.repo.Update(ctx, record)
	if err != nil {
		return Account{}, err
	}
	_, effective := service.resolveStored(ctx, []StoredAccount{updated})[updated.ID]
	return storedAccountView(updated, effective, service.exactEnvironmentIdentity(updated.AccountKey, updated.MailboxEmail)), nil
}

// Load returns the effective runtime configuration. Database records replace an
// environment account with the same stable key and normalized mailbox. A
// disabled or unreadable database override fails closed instead of falling back
// to the environment secret.
func (service *Service) Load(ctx context.Context) (config.OutreachConfig, error) {
	stored, err := service.listStored(ctx)
	if err != nil {
		return config.OutreachConfig{}, err
	}
	resolved := service.resolveStored(ctx, stored)
	effective, senderOverrides := mergeEnvironmentAccounts(service.base.GoogleWorkspaceAccounts, stored, resolved)

	// A dedicated environment inbox is not a bulk sender. Its database override
	// must retain that role instead of being appended to the sender rotation.
	dedicatedOverrides := make(map[uuid.UUID]struct{})
	if inbound := service.base.InboundMailbox; inbound != nil && !containsAccount(service.base.GoogleWorkspaceAccounts, *inbound) {
		if record, found := findStoredConflict(stored, *inbound); found {
			dedicatedOverrides[record.ID] = struct{}{}
		}
	}
	seenKeys, seenMailboxes := accountIdentitySets(effective)
	for _, record := range stored {
		if _, used := senderOverrides[record.ID]; used {
			continue
		}
		if _, dedicated := dedicatedOverrides[record.ID]; dedicated {
			continue
		}
		account, active := resolved[record.ID]
		if !active || accountIdentitySeen(seenKeys, seenMailboxes, account) {
			continue
		}
		effective = append(effective, account)
		addAccountIdentity(seenKeys, seenMailboxes, account)
	}
	result := service.base
	result.GoogleWorkspaceAccounts = effective
	if result.InboundEnabled {
		result.InboundMailboxes, _ = mergeEnvironmentAccounts(service.base.InboundMailboxes, stored, resolved)
		pollKeys, pollMailboxes := accountIdentitySets(result.InboundMailboxes)
		for _, account := range effective {
			if accountIdentitySeen(pollKeys, pollMailboxes, account) {
				continue
			}
			result.InboundMailboxes = append(result.InboundMailboxes, account)
			addAccountIdentity(pollKeys, pollMailboxes, account)
		}
		if inbound := service.base.InboundMailbox; inbound != nil {
			if record, found := findStoredConflict(stored, *inbound); found {
				if account, active := resolved[record.ID]; active {
					result.InboundMailbox = &account
				} else {
					result.InboundMailbox = nil
				}
			} else {
				copy := *inbound
				result.InboundMailbox = &copy
			}
		}
	}
	return result, nil
}

func (service *Service) listStored(ctx context.Context) ([]StoredAccount, error) {
	if service == nil || service.repo == nil {
		return []StoredAccount{}, nil
	}
	return service.repo.List(ctx)
}

func (service *Service) environmentAccounts() []config.GmailMailConfig {
	accounts := append([]config.GmailMailConfig(nil), service.base.GoogleWorkspaceAccounts...)
	seenKeys, seenMailboxes := accountIdentitySets(accounts)
	if inbound := service.base.InboundMailbox; inbound != nil {
		if _, keyExists := seenKeys[normalizedAccountKey(inbound.AccountKey)]; !keyExists {
			if _, mailboxExists := seenMailboxes[normalizedMailbox(inbound.MailboxEmail)]; !mailboxExists {
				accounts = append(accounts, *inbound)
			}
		}
	}
	return accounts
}

func (service *Service) validateEnvironmentIdentity(key, mailbox string) (bool, error) {
	partial := false
	for _, account := range service.environmentAccounts() {
		keyMatch := normalizedAccountKey(account.AccountKey) == normalizedAccountKey(key)
		mailboxMatch := normalizedMailbox(account.MailboxEmail) == normalizedMailbox(mailbox)
		if keyMatch && mailboxMatch {
			return true, nil
		}
		partial = partial || keyMatch || mailboxMatch
	}
	if partial {
		return false, fmt.Errorf("%w: replacing an environment account must keep its existing account_key and mailbox_email together", ErrInvalid)
	}
	return false, nil
}

func (service *Service) exactEnvironmentIdentity(key, mailbox string) bool {
	for _, account := range service.environmentAccounts() {
		if sameAccountIdentity(account, config.GmailMailConfig{AccountKey: key, MailboxEmail: mailbox}) {
			return true
		}
	}
	return false
}

func accountIdentitySets(accounts []config.GmailMailConfig) (map[string]struct{}, map[string]struct{}) {
	keys := make(map[string]struct{}, len(accounts))
	mailboxes := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		addAccountIdentity(keys, mailboxes, account)
	}
	return keys, mailboxes
}

func storedAccountView(record StoredAccount, effective, overridesEnvironment bool) Account {
	id := record.ID
	createdAt := record.CreatedAt
	updatedAt := record.UpdatedAt
	return Account{
		ID: &id, AccountKey: record.AccountKey, MailboxEmail: record.MailboxEmail, FromEmail: record.FromEmail,
		Source: "database", Enabled: record.Enabled, Effective: effective,
		Editable: true, CredentialsStored: len(record.CredentialCiphertext) > 0,
		OverridesEnvironment: overridesEnvironment, CreatedAt: &createdAt, UpdatedAt: &updatedAt,
	}
}

func environmentAccountView(account config.GmailMailConfig) Account {
	return Account{
		AccountKey: account.AccountKey, MailboxEmail: account.MailboxEmail, FromEmail: account.FromEmail,
		Source: "environment", Enabled: true, Effective: true, Editable: false, CredentialsStored: true,
	}
}

func (service *Service) resolveStored(ctx context.Context, stored []StoredAccount) map[uuid.UUID]config.GmailMailConfig {
	resolved := make(map[uuid.UUID]config.GmailMailConfig, len(stored))
	if service.vault == nil {
		return resolved
	}
	for _, record := range stored {
		if !record.Enabled {
			continue
		}
		payload, err := service.vault.decrypt(record.AccountKey, record.MailboxEmail, record.CredentialCiphertext)
		if err != nil {
			service.log.ErrorContext(ctx, "database_outreach_credential_unavailable", "account_key", record.AccountKey, "error", err)
			continue
		}
		resolved[record.ID] = config.GmailMailConfig{
			AccountKey: record.AccountKey, MailboxEmail: record.MailboxEmail, FromEmail: record.FromEmail,
			ClientID: payload.ClientID, ClientSecret: payload.ClientSecret, RefreshToken: payload.RefreshToken,
		}
	}
	return resolved
}

func mergeEnvironmentAccounts(environment []config.GmailMailConfig, stored []StoredAccount, resolved map[uuid.UUID]config.GmailMailConfig) ([]config.GmailMailConfig, map[uuid.UUID]struct{}) {
	accounts := make([]config.GmailMailConfig, 0, len(environment))
	used := make(map[uuid.UUID]struct{})
	for _, account := range environment {
		if record, found := findStoredConflict(stored, account); found {
			used[record.ID] = struct{}{}
			if replacement, active := resolved[record.ID]; active {
				accounts = append(accounts, replacement)
			}
			continue
		}
		accounts = append(accounts, account)
	}
	return accounts, used
}

func findStoredConflict(stored []StoredAccount, account config.GmailMailConfig) (StoredAccount, bool) {
	for _, record := range stored {
		if normalizedAccountKey(record.AccountKey) == normalizedAccountKey(account.AccountKey) ||
			normalizedMailbox(record.MailboxEmail) == normalizedMailbox(account.MailboxEmail) {
			return record, true
		}
	}
	return StoredAccount{}, false
}

func containsAccount(accounts []config.GmailMailConfig, target config.GmailMailConfig) bool {
	for _, account := range accounts {
		if sameAccountIdentity(account, target) {
			return true
		}
	}
	return false
}

func sameAccountIdentity(left, right config.GmailMailConfig) bool {
	return normalizedAccountKey(left.AccountKey) == normalizedAccountKey(right.AccountKey) &&
		normalizedMailbox(left.MailboxEmail) == normalizedMailbox(right.MailboxEmail)
}

func accountIdentitySeen(keys, mailboxes map[string]struct{}, account config.GmailMailConfig) bool {
	_, keyExists := keys[normalizedAccountKey(account.AccountKey)]
	_, mailboxExists := mailboxes[normalizedMailbox(account.MailboxEmail)]
	return keyExists || mailboxExists
}

func addAccountIdentity(keys, mailboxes map[string]struct{}, account config.GmailMailConfig) {
	keys[normalizedAccountKey(account.AccountKey)] = struct{}{}
	mailboxes[normalizedMailbox(account.MailboxEmail)] = struct{}{}
}

func normalizedAccountKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizedMailbox(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func canonicalMailbox(value string) (string, error) {
	value = strings.TrimSpace(value)
	address, err := mail.ParseAddress(value)
	if err != nil || address == nil || address.Name != "" || strings.TrimSpace(address.Address) == "" {
		return "", fmt.Errorf("must be one valid email address")
	}
	canonical := strings.ToLower(strings.TrimSpace(address.Address))
	if len(canonical) > 320 || strings.ContainsAny(canonical, "\r\n") {
		return "", fmt.Errorf("must be one valid email address")
	}
	return canonical, nil
}

func validateCredentialPayload(payload credentialPayload) error {
	values := []struct {
		name  string
		value string
		max   int
	}{
		{name: "client_id", value: payload.ClientID, max: 512},
		{name: "client_secret", value: payload.ClientSecret, max: 2048},
		{name: "refresh_token", value: payload.RefreshToken, max: 4096},
	}
	for _, field := range values {
		value := strings.TrimSpace(field.value)
		if value == "" || len(value) > field.max || strings.ContainsAny(value, "\x00\r\n") {
			return fmt.Errorf("%w: credentials.%s is required and must be a single safe value", ErrInvalid, field.name)
		}
	}
	return nil
}
