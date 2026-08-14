package outreachaccounts

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(context.Context) ([]StoredAccount, error)
	Get(context.Context, uuid.UUID) (StoredAccount, error)
	Create(context.Context, StoredAccount) (StoredAccount, error)
	Update(context.Context, StoredAccount) (StoredAccount, error)
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) List(ctx context.Context) ([]StoredAccount, error) {
	if repo == nil || repo.pool == nil {
		return []StoredAccount{}, nil
	}
	rows, err := repo.pool.Query(ctx, `
		SELECT id, account_key, mailbox_email, from_email, credential_ciphertext,
		       encryption_version, enabled, created_by, updated_by, created_at, updated_at
		FROM outreach_email_credentials
		ORDER BY created_at ASC, account_key ASC`)
	if err != nil {
		return nil, fmt.Errorf("list database outreach email accounts: %w", err)
	}
	defer rows.Close()
	accounts := make([]StoredAccount, 0)
	for rows.Next() {
		account, scanErr := scanStoredAccount(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (repo *Postgres) Get(ctx context.Context, id uuid.UUID) (StoredAccount, error) {
	if repo == nil || repo.pool == nil {
		return StoredAccount{}, ErrNotFound
	}
	account, err := scanStoredAccount(repo.pool.QueryRow(ctx, `
		SELECT id, account_key, mailbox_email, from_email, credential_ciphertext,
		       encryption_version, enabled, created_by, updated_by, created_at, updated_at
		FROM outreach_email_credentials
		WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredAccount{}, ErrNotFound
	}
	return account, err
}

func (repo *Postgres) Create(ctx context.Context, account StoredAccount) (StoredAccount, error) {
	if repo == nil || repo.pool == nil {
		return StoredAccount{}, fmt.Errorf("database pool is not configured")
	}
	created, err := scanStoredAccount(repo.pool.QueryRow(ctx, `
		INSERT INTO outreach_email_credentials (
		  account_key, mailbox_email, from_email, credential_ciphertext,
		  encryption_version, enabled, created_by, updated_by
		) VALUES ($1, $2, $3, $4, 1, $5, $6, $6)
		RETURNING id, account_key, mailbox_email, from_email, credential_ciphertext,
		          encryption_version, enabled, created_by, updated_by, created_at, updated_at`,
		account.AccountKey, account.MailboxEmail, account.FromEmail, account.CredentialCiphertext,
		account.Enabled, account.CreatedBy,
	))
	if isUniqueViolation(err) {
		return StoredAccount{}, ErrDuplicate
	}
	return created, err
}

func (repo *Postgres) Update(ctx context.Context, account StoredAccount) (StoredAccount, error) {
	if repo == nil || repo.pool == nil {
		return StoredAccount{}, fmt.Errorf("database pool is not configured")
	}
	updated, err := scanStoredAccount(repo.pool.QueryRow(ctx, `
		UPDATE outreach_email_credentials
		SET from_email = $2,
		    credential_ciphertext = $3,
		    enabled = $4,
		    updated_by = $5,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, account_key, mailbox_email, from_email, credential_ciphertext,
		          encryption_version, enabled, created_by, updated_by, created_at, updated_at`,
		account.ID, account.FromEmail, account.CredentialCiphertext, account.Enabled, account.UpdatedBy,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredAccount{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return StoredAccount{}, ErrDuplicate
	}
	return updated, err
}

type accountScanner interface {
	Scan(...any) error
}

func scanStoredAccount(row accountScanner) (StoredAccount, error) {
	var account StoredAccount
	err := row.Scan(
		&account.ID,
		&account.AccountKey,
		&account.MailboxEmail,
		&account.FromEmail,
		&account.CredentialCiphertext,
		&account.EncryptionVersion,
		&account.Enabled,
		&account.CreatedBy,
		&account.UpdatedBy,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return StoredAccount{}, fmt.Errorf("scan database outreach email account: %w", err)
	}
	return account, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var _ Repository = (*Postgres)(nil)
