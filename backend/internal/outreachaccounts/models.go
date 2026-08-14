package outreachaccounts

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrForbidden             = errors.New("outreach email account access is forbidden")
	ErrInvalid               = errors.New("outreach email account is invalid")
	ErrDuplicate             = errors.New("outreach email account already exists")
	ErrNotFound              = errors.New("outreach email account was not found")
	ErrEnvironmentManaged    = errors.New("environment-managed outreach email accounts are read-only")
	ErrEncryptionUnavailable = errors.New("outreach credential encryption is unavailable")
)

type StoredAccount struct {
	ID                   uuid.UUID
	AccountKey           string
	MailboxEmail         string
	FromEmail            string
	CredentialCiphertext []byte
	EncryptionVersion    int16
	Enabled              bool
	CreatedBy            *uuid.UUID
	UpdatedBy            *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Account struct {
	ID                    *uuid.UUID `json:"id,omitempty"`
	AccountKey            string     `json:"account_key"`
	MailboxEmail          string     `json:"mailbox_email"`
	FromEmail             string     `json:"from_email"`
	Source                string     `json:"source"`
	Enabled               bool       `json:"enabled"`
	Effective             bool       `json:"effective"`
	Editable              bool       `json:"editable"`
	CredentialsStored     bool       `json:"credentials_stored"`
	DatabaseFallback      bool       `json:"database_fallback,omitempty"`
	ShadowedByEnvironment bool       `json:"shadowed_by_environment,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}

type ListResult struct {
	Accounts        []Account `json:"accounts"`
	EncryptionReady bool      `json:"encryption_ready"`
}

type CredentialsInput struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

type CreateInput struct {
	AccountKey   string           `json:"account_key"`
	MailboxEmail string           `json:"mailbox_email"`
	FromEmail    string           `json:"from_email"`
	Credentials  CredentialsInput `json:"credentials"`
	Enabled      *bool            `json:"enabled,omitempty"`
}

type UpdateInput struct {
	FromEmail   *string           `json:"from_email,omitempty"`
	Credentials *CredentialsInput `json:"credentials,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}
