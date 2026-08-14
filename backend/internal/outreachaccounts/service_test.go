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

func TestServiceCreateRejectsEnvironmentDuplicateAndNeverReturnsSecrets(t *testing.T) {
	repo := &memoryRepository{}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)
	principal := testAdmin()

	_, err := service.Create(context.Background(), principal, CreateInput{
		AccountKey:   "other",
		MailboxEmail: "ENV@example.com",
		Credentials:  CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
	})
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("Create() error = %v, want duplicate environment mailbox", err)
	}

	created, err := service.Create(context.Background(), principal, CreateInput{
		AccountKey:   "support",
		MailboxEmail: "Support@Example.com",
		Credentials:  CredentialsInput{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Source != "database" || created.MailboxEmail != "support@example.com" || !created.CredentialsStored {
		t.Fatalf("Create() = %#v", created)
	}
	if len(repo.records) != 1 || bytes.Contains(repo.records[0].CredentialCiphertext, []byte("refresh")) {
		t.Fatal("credentials were not stored as a single encrypted payload")
	}
}

func TestServiceLoadMergesEnvironmentAndEnabledDatabaseAccounts(t *testing.T) {
	vault, err := newCredentialCipher(testEncryptionKey())
	if err != nil {
		t.Fatal(err)
	}
	encrypted := func(key, mailbox string) []byte {
		value, encryptErr := vault.encrypt(key, mailbox, credentialPayload{ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"})
		if encryptErr != nil {
			t.Fatal(encryptErr)
		}
		return value
	}
	repo := &memoryRepository{records: []StoredAccount{
		{ID: uuid.New(), AccountKey: "db", MailboxEmail: "db@example.com", FromEmail: "db@example.com", Enabled: true, CredentialCiphertext: encrypted("db", "db@example.com")},
		{ID: uuid.New(), AccountKey: "shadow-key", MailboxEmail: "env@example.com", FromEmail: "env@example.com", Enabled: true, CredentialCiphertext: encrypted("shadow-key", "env@example.com")},
		{ID: uuid.New(), AccountKey: "disabled", MailboxEmail: "disabled@example.com", FromEmail: "disabled@example.com", Enabled: false, CredentialCiphertext: encrypted("disabled", "disabled@example.com")},
	}}
	service := NewService(repo, testBaseConfig(), testEncryptionKey(), nil)

	effective, err := service.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(effective.GoogleWorkspaceAccounts) != 2 {
		t.Fatalf("GoogleWorkspaceAccounts = %#v, want environment plus one database account", effective.GoogleWorkspaceAccounts)
	}
	if effective.GoogleWorkspaceAccounts[0].AccountKey != "env" || effective.GoogleWorkspaceAccounts[1].AccountKey != "db" {
		t.Fatalf("GoogleWorkspaceAccounts order = %#v", effective.GoogleWorkspaceAccounts)
	}
	if len(effective.InboundMailboxes) != 2 || effective.InboundMailboxes[1].AccountKey != "db" {
		t.Fatalf("InboundMailboxes = %#v, want database sender included once", effective.InboundMailboxes)
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

func testAdmin() auth.Principal {
	return auth.Principal{UserID: uuid.New(), Role: auth.RoleInternalAdmin}
}
