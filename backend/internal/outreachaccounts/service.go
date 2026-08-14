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
	accounts := service.environmentViews()
	envByKey := make(map[string]int, len(accounts))
	envByMailbox := make(map[string]int, len(accounts))
	for index, account := range accounts {
		envByKey[account.AccountKey] = index
		envByMailbox[account.MailboxEmail] = index
	}
	for _, record := range stored {
		if index, exists := envByMailbox[record.MailboxEmail]; exists {
			accounts[index].DatabaseFallback = true
			continue
		}
		shadowed := false
		if _, exists := envByKey[record.AccountKey]; exists {
			shadowed = true
		}
		effective := record.Enabled && !shadowed && service.vault != nil
		if effective {
			if _, decryptErr := service.vault.decrypt(record.AccountKey, record.MailboxEmail, record.CredentialCiphertext); decryptErr != nil {
				effective = false
				service.log.ErrorContext(ctx, "database_outreach_credential_unavailable", "account_key", record.AccountKey, "error", decryptErr)
			}
		}
		id := record.ID
		createdAt := record.CreatedAt
		updatedAt := record.UpdatedAt
		accounts = append(accounts, Account{
			ID:                    &id,
			AccountKey:            record.AccountKey,
			MailboxEmail:          record.MailboxEmail,
			FromEmail:             record.FromEmail,
			Source:                "database",
			Enabled:               record.Enabled,
			Effective:             effective,
			Editable:              true,
			CredentialsStored:     len(record.CredentialCiphertext) > 0,
			ShadowedByEnvironment: shadowed,
			CreatedAt:             &createdAt,
			UpdatedAt:             &updatedAt,
		})
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
	if service.conflictsWithEnvironment(accountKey, mailbox) {
		return Account{}, fmt.Errorf("%w: mailbox or account key is already provided by protected environment configuration", ErrDuplicate)
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
	return storedAccountView(created, service.vault != nil, false), nil
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
	if input.FromEmail != nil {
		record.FromEmail, err = canonicalMailbox(*input.FromEmail)
		if err != nil {
			return Account{}, fmt.Errorf("%w: from_email %v", ErrInvalid, err)
		}
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
	}
	actor := principal.UserID
	record.UpdatedBy = &actor
	updated, err := service.repo.Update(ctx, record)
	if err != nil {
		return Account{}, err
	}
	return storedAccountView(updated, service.vault != nil, false), nil
}

// Load returns the effective runtime configuration. Environment accounts win
// by stable key or normalized mailbox identity, so the same mailbox is never
// registered twice when it exists in both sources.
func (service *Service) Load(ctx context.Context) (config.OutreachConfig, error) {
	stored, err := service.listStored(ctx)
	if err != nil {
		return config.OutreachConfig{}, err
	}
	effective := append([]config.GmailMailConfig(nil), service.base.GoogleWorkspaceAccounts...)
	seenKeys, seenMailboxes := accountIdentitySets(service.environmentAccounts())
	for _, record := range stored {
		if !record.Enabled || service.vault == nil {
			continue
		}
		if _, exists := seenKeys[record.AccountKey]; exists {
			continue
		}
		if _, exists := seenMailboxes[record.MailboxEmail]; exists {
			continue
		}
		payload, decryptErr := service.vault.decrypt(record.AccountKey, record.MailboxEmail, record.CredentialCiphertext)
		if decryptErr != nil {
			service.log.ErrorContext(ctx, "database_outreach_credential_unavailable", "account_key", record.AccountKey, "error", decryptErr)
			continue
		}
		effective = append(effective, config.GmailMailConfig{
			AccountKey: record.AccountKey, MailboxEmail: record.MailboxEmail, FromEmail: record.FromEmail,
			ClientID: payload.ClientID, ClientSecret: payload.ClientSecret, RefreshToken: payload.RefreshToken,
		})
		seenKeys[record.AccountKey] = struct{}{}
		seenMailboxes[record.MailboxEmail] = struct{}{}
	}
	result := service.base
	result.GoogleWorkspaceAccounts = effective
	if result.InboundEnabled {
		result.InboundMailboxes = append([]config.GmailMailConfig(nil), service.base.InboundMailboxes...)
		pollKeys, pollMailboxes := accountIdentitySets(result.InboundMailboxes)
		for _, account := range effective {
			if _, exists := pollKeys[account.AccountKey]; exists {
				continue
			}
			if _, exists := pollMailboxes[account.MailboxEmail]; exists {
				continue
			}
			result.InboundMailboxes = append(result.InboundMailboxes, account)
			pollKeys[account.AccountKey] = struct{}{}
			pollMailboxes[account.MailboxEmail] = struct{}{}
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
		if _, keyExists := seenKeys[inbound.AccountKey]; !keyExists {
			if _, mailboxExists := seenMailboxes[inbound.MailboxEmail]; !mailboxExists {
				accounts = append(accounts, *inbound)
			}
		}
	}
	return accounts
}

func (service *Service) environmentViews() []Account {
	configured := service.environmentAccounts()
	views := make([]Account, 0, len(configured))
	for _, account := range configured {
		views = append(views, Account{
			AccountKey: account.AccountKey, MailboxEmail: account.MailboxEmail, FromEmail: account.FromEmail,
			Source: "environment", Enabled: true, Effective: true, Editable: false, CredentialsStored: true,
		})
	}
	return views
}

func (service *Service) conflictsWithEnvironment(key, mailbox string) bool {
	keys, mailboxes := accountIdentitySets(service.environmentAccounts())
	_, keyExists := keys[key]
	_, mailboxExists := mailboxes[mailbox]
	return keyExists || mailboxExists
}

func accountIdentitySets(accounts []config.GmailMailConfig) (map[string]struct{}, map[string]struct{}) {
	keys := make(map[string]struct{}, len(accounts))
	mailboxes := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		keys[strings.TrimSpace(account.AccountKey)] = struct{}{}
		mailboxes[strings.ToLower(strings.TrimSpace(account.MailboxEmail))] = struct{}{}
	}
	return keys, mailboxes
}

func storedAccountView(record StoredAccount, encryptionReady, shadowed bool) Account {
	id := record.ID
	createdAt := record.CreatedAt
	updatedAt := record.UpdatedAt
	return Account{
		ID: &id, AccountKey: record.AccountKey, MailboxEmail: record.MailboxEmail, FromEmail: record.FromEmail,
		Source: "database", Enabled: record.Enabled, Effective: record.Enabled && encryptionReady && !shadowed,
		Editable: true, CredentialsStored: len(record.CredentialCiphertext) > 0,
		ShadowedByEnvironment: shadowed, CreatedAt: &createdAt, UpdatedAt: &updatedAt,
	}
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
