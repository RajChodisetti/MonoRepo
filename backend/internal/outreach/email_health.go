package outreach

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func (repo *Postgres) SyncEmailHealthAccounts(
	ctx context.Context,
	accounts []emailprovider.HealthAccountConfig,
	interval time.Duration,
) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	if interval < 24*time.Hour {
		return fmt.Errorf("email health interval must be at least 24h")
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email health account sync: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	keys := make([]string, 0, len(accounts))
	for _, account := range accounts {
		key := strings.TrimSpace(account.Key)
		if key == "" || strings.TrimSpace(account.ProviderIdentity) == "" || strings.TrimSpace(account.FromEmail) == "" {
			return fmt.Errorf("email health account metadata is incomplete")
		}
		keys = append(keys, key)
		status := "pending"
		var nextCheckAt any = time.Now().UTC()
		if !account.Enabled {
			status = "disabled"
			nextCheckAt = nil
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO outreach_email_account_health (
				account_key, provider, provider_identity, from_email, enabled,
				health_status, next_check_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (provider_identity) DO UPDATE
			SET account_key = EXCLUDED.account_key,
			    provider = EXCLUDED.provider,
			    from_email = EXCLUDED.from_email,
			    enabled = EXCLUDED.enabled,
			    health_status = CASE
			      WHEN EXCLUDED.enabled = false THEN 'disabled'
			      WHEN outreach_email_account_health.health_status = 'disabled' THEN 'pending'
			      ELSE outreach_email_account_health.health_status
			    END,
			    next_check_at = CASE
			      WHEN EXCLUDED.enabled = false THEN NULL
			      WHEN outreach_email_account_health.next_check_at IS NULL THEN now()
			      ELSE outreach_email_account_health.next_check_at
			    END,
			    updated_at = now()`,
			key,
			account.Provider,
			strings.ToLower(strings.TrimSpace(account.ProviderIdentity)),
			strings.ToLower(strings.TrimSpace(account.FromEmail)),
			account.Enabled,
			status,
			nextCheckAt,
		)
		if err != nil {
			return fmt.Errorf("sync email health account %q: %w", key, err)
		}
	}

	if len(keys) == 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE outreach_email_account_health
			SET enabled = false, health_status = 'disabled', next_check_at = NULL, updated_at = now()
			WHERE enabled = true`); err != nil {
			return fmt.Errorf("disable removed email health accounts: %w", err)
		}
	} else if _, err := tx.Exec(ctx, `
		UPDATE outreach_email_account_health
		SET enabled = false, health_status = 'disabled', next_check_at = NULL, updated_at = now()
		WHERE NOT (account_key = ANY($1::text[]))`, keys); err != nil {
		return fmt.Errorf("disable removed email health accounts: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email health account sync: %w", err)
	}
	return nil
}

func (repo *Postgres) ClaimDueEmailHealthAccounts(
	ctx context.Context,
	accountKeys []string,
	interval time.Duration,
) ([]string, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	if len(accountKeys) == 0 {
		return nil, nil
	}
	if interval < 24*time.Hour {
		return nil, fmt.Errorf("email health interval must be at least 24h")
	}

	rows, err := repo.pool.Query(ctx, `
		WITH due AS (
		  SELECT account_key
		  FROM outreach_email_account_health
		  WHERE enabled = true
		    AND account_key = ANY($1::text[])
		    AND next_check_at <= clock_timestamp()
		  ORDER BY next_check_at ASC, account_key ASC
		  FOR UPDATE SKIP LOCKED
		), claimed AS (
		  UPDATE outreach_email_account_health AS health
		  SET health_status = 'checking',
		      next_check_at = clock_timestamp() + ($2 * interval '1 second'),
		      updated_at = clock_timestamp()
		  FROM due
		  WHERE health.account_key = due.account_key
		  RETURNING health.account_key
		)
		SELECT account_key FROM claimed ORDER BY account_key`,
		accountKeys,
		int64(interval/time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("claim due email health accounts: %w", err)
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("scan due email health account: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (repo *Postgres) RecordEmailHealthResult(
	ctx context.Context,
	accountKey string,
	healthy bool,
	providerMessageID string,
	safeError string,
) error {
	status := "failed"
	if healthy {
		status = "healthy"
		safeError = ""
	}
	result, err := repo.pool.Exec(ctx, `
		UPDATE outreach_email_account_health
		SET health_status = $2,
		    last_checked_at = now(),
		    provider_message_id = $3,
		    last_error = $4,
		    updated_at = now()
		WHERE account_key = $1 AND enabled = true`,
		strings.TrimSpace(accountKey), status, strings.TrimSpace(providerMessageID), strings.TrimSpace(safeError),
	)
	if err != nil {
		return fmt.Errorf("record email health result: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("email health account was not found")
	}
	return nil
}

func (repo *Postgres) ListEmailHealth(ctx context.Context) ([]emailprovider.HealthStatus, error) {
	if repo.pool == nil {
		return []emailprovider.HealthStatus{}, nil
	}
	rows, err := repo.pool.Query(ctx, `
		SELECT account_key, provider, provider_identity, from_email, enabled,
		       health_status, last_checked_at, next_check_at,
		       provider_message_id, last_error
		FROM outreach_email_account_health
		ORDER BY enabled DESC, provider ASC, from_email ASC`)
	if errors.Is(err, pgx.ErrNoRows) {
		return []emailprovider.HealthStatus{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list email health accounts: %w", err)
	}
	defer rows.Close()
	statuses := make([]emailprovider.HealthStatus, 0)
	for rows.Next() {
		var status emailprovider.HealthStatus
		if err := rows.Scan(
			&status.AccountKey,
			&status.Provider,
			&status.ProviderIdentity,
			&status.FromEmail,
			&status.Enabled,
			&status.Status,
			&status.LastCheckedAt,
			&status.NextCheckAt,
			&status.ProviderMessageID,
			&status.LastError,
		); err != nil {
			return nil, fmt.Errorf("scan email health account: %w", err)
		}
		statuses = append(statuses, status)
	}
	return statuses, rows.Err()
}

var _ emailprovider.HealthStore = (*Postgres)(nil)
