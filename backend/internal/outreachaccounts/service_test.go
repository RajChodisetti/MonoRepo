package outreachaccounts

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type memoryRepository struct {
	records []StoredAccount
}

func (repo *memoryRepository) List(context.Context) ([]StoredAccount, error) {
	return append([]StoredAccount(nil), repo.records...), nil
}

func (repo *memoryRepository) Get(_ context.Context, id uuid.UUID) (StoredAccount, error) {
	for _, record := range repo.records {
		if record.ID == id {
			return record, nil
		}
	}
	return StoredAccount{}, ErrNotFound
}

func (repo *memoryRepository) Create(_ context.Context, record StoredAccount) (StoredAccount, error) {
	for _, existing := range repo.records {
		if existing.AccountKey == record.AccountKey || existing.MailboxEmail == record.MailboxEmail {
			return StoredAccount{}, ErrDuplicate
		}
	}
	record.ID = uuid.New()
	record.CreatedAt = time.Now().UTC()
	record.UpdatedAt = record.CreatedAt
	repo.records = append(repo.records, record)
	return record, nil
}

func (repo *memoryRepository) Update(_ context.Context, record StoredAccount) (StoredAccount, error) {
	for index, existing := range repo.records {
		if existing.ID == record.ID {
			record.UpdatedAt = time.Now().UTC()
			repo.records[index] = record
			return record, nil
		}
	}
	return StoredAccount{}, ErrNotFound
}

func TestCredentialCipherRoundTripUsesAuthenticatedIdentity(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	payload := credentialPayload{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"}
	ciphertext, err := vault.encrypt("sales", "sales@example.com", payload)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte(payload.ClientSecret)) || bytes.Contains(ciphertext, []byte(payload.RefreshToken)) {
		t.Fatal("encrypted payload contains plaintext credential material")
	}
	decoded, err := vault.decrypt("sales", "sales@example.com", ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != payload {
		t.Fatalf("decrypt() = %#v, want %#v", decoded, payload)
	}
	if _, err := vault.decrypt("other", "sales@example.com", ciphertext); err == nil {
		t.Fatal("decrypt() accepted ciphertext for a different account identity")
	}
}

func TestServiceCreateAllowsExactEnvironmentOverrideAndRejectsPartialConflicts(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)
	principal := testAdmin()

	for name, input := range map[string]CreateInput{
		"mailbox only": {
			AccountKey:   "other",
			MailboxEmail: "ENV@example.com",
			Credentials:  CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
		},
		"key only": {
			AccountKey:   "env",
			MailboxEmail: "other@example.com",
			Credentials:  CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := service.Create(context.Background(), principal, input)
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Create() error = %v, want invalid partial environment identity", err)
			}
		})
	}

	created, err := service.Create(context.Background(), principal, CreateInput{
		AccountKey:   "env",
		MailboxEmail: "ENV@example.com",
		FromEmail:    "new-sender@example.com",
		Credentials:  CredentialsInput{ClientID: "new-client", ClientSecret: "new-secret", RefreshToken: "new-refresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Source != "database" || !created.OverridesEnvironment || created.MailboxEmail != "env@example.com" || !created.CredentialsStored {
		t.Fatalf("Create() = %#v", created)
	}
	if len(repo.records) != 1 || bytes.Contains(repo.records[0].CredentialCiphertext, []byte("new-refresh")) {
		t.Fatal("credentials were not stored as a single encrypted payload")
	}
}

func TestServiceCreateDatabaseOnlyAccountNeverReturnsSecrets(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)
	created, err := service.Create(context.Background(), testAdmin(), CreateInput{
		AccountKey:   "other",
		MailboxEmail: "other@example.com",
		Credentials:  CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Source != "database" || created.OverridesEnvironment || created.MailboxEmail != "other@example.com" || !created.CredentialsStored {
		t.Fatalf("Create() = %#v", created)
	}
	if len(repo.records) != 1 || bytes.Contains(repo.records[0].CredentialCiphertext, []byte("refresh")) {
		t.Fatal("credentials were not stored as a single encrypted payload")
	}
}

func TestServiceListShowsDatabaseOverrideOnce(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	override := storedFixture(t, vault, "env", "env@example.com", true, "database")
	repo := &memoryRepository{records: []StoredAccount{override}}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)

	result, err := service.List(context.Background(), testAdmin())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Accounts) != 1 {
		t.Fatalf("Accounts = %#v, want one deduplicated account", result.Accounts)
	}
	account := result.Accounts[0]
	if account.Source != "database" || !account.OverridesEnvironment || !account.Editable || !account.Effective {
		t.Fatalf("override view = %#v", account)
	}
	if account.DatabaseFallback || account.ShadowedByEnvironment {
		t.Fatalf("legacy environment-precedence flags unexpectedly set: %#v", account)
	}
}

func TestServiceLoadDatabaseOverridesEnvironmentAndAppendsDatabaseOnly(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	repo := &memoryRepository{records: []StoredAccount{
		storedFixture(t, vault, "env", "env@example.com", true, "override"),
		storedFixture(t, vault, "db", "db@example.com", true, "database"),
		storedFixture(t, vault, "disabled", "disabled@example.com", false, "disabled"),
	}}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)

	effective, err := service.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective.GoogleWorkspaceAccounts) != 2 {
		t.Fatalf("GoogleWorkspaceAccounts = %#v, want override plus one database account", effective.GoogleWorkspaceAccounts)
	}
	if effective.GoogleWorkspaceAccounts[0].AccountKey != "env" || effective.GoogleWorkspaceAccounts[0].ClientID != "override-client" || effective.GoogleWorkspaceAccounts[1].AccountKey != "db" {
		t.Fatalf("GoogleWorkspaceAccounts order = %#v", effective.GoogleWorkspaceAccounts)
	}
	if len(effective.InboundMailboxes) != 2 || effective.InboundMailboxes[0].ClientID != "override-client" || effective.InboundMailboxes[1].AccountKey != "db" {
		t.Fatalf("InboundMailboxes = %#v, want database override and sender exactly once", effective.InboundMailboxes)
	}
	if effective.InboundMailbox == nil || effective.InboundMailbox.ClientID != "override-client" {
		t.Fatalf("InboundMailbox = %#v, want database override", effective.InboundMailbox)
	}
}

func TestServiceLoadDisabledOverrideBlocksEnvironmentFallback(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	repo := &memoryRepository{records: []StoredAccount{
		storedFixture(t, vault, "env", "env@example.com", false, "disabled"),
	}}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)

	effective, err := service.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective.GoogleWorkspaceAccounts) != 0 || len(effective.InboundMailboxes) != 0 || effective.InboundMailbox != nil {
		t.Fatalf("Load() = %#v, disabled database override must fail closed", effective)
	}
}

func TestServiceLoadUnreadableOverrideBlocksEnvironmentFallback(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	override := storedFixture(t, vault, "env", "env@example.com", true, "override")
	override.CredentialCiphertext = []byte("not-valid-ciphertext")
	repo := &memoryRepository{records: []StoredAccount{override}}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)

	effective, err := service.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective.GoogleWorkspaceAccounts) != 0 || len(effective.InboundMailboxes) != 0 || effective.InboundMailbox != nil {
		t.Fatalf("Load() = %#v, unreadable database override must fail closed", effective)
	}
	listed, err := service.List(context.Background(), testAdmin())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Accounts) != 1 || listed.Accounts[0].Effective || !listed.Accounts[0].OverridesEnvironment {
		t.Fatalf("List() = %#v, unreadable override must be visible and unavailable", listed.Accounts)
	}
}

func TestServiceLoadDedicatedInboxOverrideDoesNotJoinSenderRotation(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	base := testBaseConfig()
	dedicated := config.GmailMailConfig{
		AccountKey: "inbound", MailboxEmail: "inbound@example.com", FromEmail: "inbound@example.com",
		ClientID: "env-inbound-client", ClientSecret: "env-inbound-secret", RefreshToken: "env-inbound-refresh",
	}
	base.InboundMailbox = &dedicated
	base.InboundMailboxes = append(base.InboundMailboxes, dedicated)
	repo := &memoryRepository{records: []StoredAccount{
		storedFixture(t, vault, "inbound", "inbound@example.com", true, "override"),
	}}
	service := NewService(repo, base, testEncryptionKey(), nil)

	effective, err := service.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective.GoogleWorkspaceAccounts) != 1 || effective.GoogleWorkspaceAccounts[0].AccountKey != "env" {
		t.Fatalf("GoogleWorkspaceAccounts = %#v, dedicated inbox override joined sender rotation", effective.GoogleWorkspaceAccounts)
	}
	if len(effective.InboundMailboxes) != 2 || effective.InboundMailboxes[1].ClientID != "override-client" {
		t.Fatalf("InboundMailboxes = %#v", effective.InboundMailboxes)
	}
	if effective.InboundMailbox == nil || effective.InboundMailbox.ClientID != "override-client" {
		t.Fatalf("InboundMailbox = %#v", effective.InboundMailbox)
	}
}

func TestServiceUpdateCanDisableAndReplaceCredentials(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)
	created, err := service.Create(context.Background(), testAdmin(), CreateInput{
		AccountKey: "support", MailboxEmail: "support@example.com",
		Credentials: CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	before := append([]byte(nil), repo.records[0].CredentialCiphertext...)
	disabled := false
	updated, err := service.Update(context.Background(), testAdmin(), *created.ID, UpdateInput{
		Enabled:     &disabled,
		Credentials: &CredentialsInput{ClientID: "new-client", ClientSecret: "new-secret", RefreshToken: "new-refresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Enabled || bytes.Equal(before, repo.records[0].CredentialCiphertext) {
		t.Fatalf("Update() = %#v, credential replacement or disable was not persisted", updated)
	}
}

func testBaseConfig() config.OutreachConfig {
	account := config.GmailMailConfig{
		AccountKey: "env", MailboxEmail: "env@example.com", FromEmail: "env@example.com",
		ClientID: "env-client", ClientSecret: "env-secret", RefreshToken: "env-refresh",
	}
	return config.OutreachConfig{
		GoogleWorkspaceAccounts: []config.GmailMailConfig{account},
		InboundEnabled:          true,
		InboundMailbox:          &account,
		InboundMailboxes:        []config.GmailMailConfig{account},
	}
}

func testEncryptionKey() string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
}

func storedFixture(t *testing.T, vault *credentialCipher, key, mailbox string, enabled bool, prefix string) StoredAccount {
	t.Helper()
	payload := credentialPayload{
		ClientID: prefix + "-client", ClientSecret: prefix + "-secret", RefreshToken: prefix + "-refresh",
	}
	ciphertext, err := vault.encrypt(key, mailbox, payload)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	return StoredAccount{
		ID: uuid.New(), AccountKey: key, MailboxEmail: mailbox, FromEmail: mailbox,
		Enabled: enabled, CredentialCiphertext: ciphertext, CreatedAt: now, UpdatedAt: now,
	}
}

func testAdmin() auth.Principal {
	return auth.Principal{UserID: uuid.New(), Role: auth.RoleInternalAdmin}
}
