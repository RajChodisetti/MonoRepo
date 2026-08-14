package email

import (
	"context"
	"fmt"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type OutreachConfigLoader interface {
	Load(context.Context) (config.OutreachConfig, error)
}

// ReloadingAccountPool resolves environment plus database-managed accounts for
// every operation. Admin changes therefore take effect without restarting the
// API or worker, while quota ownership remains in PostgreSQL.
type ReloadingAccountPool struct {
	emailCfg config.EmailConfig
	loader   OutreachConfigLoader
	quota    QuotaStore
}

func NewReloadingPersistentAccountPool(
	emailCfg config.EmailConfig,
	loader OutreachConfigLoader,
	quota QuotaStore,
) *ReloadingAccountPool {
	return &ReloadingAccountPool{emailCfg: emailCfg, loader: loader, quota: quota}
}

func (pool *ReloadingAccountPool) Durable() bool {
	return pool != nil && pool.quota != nil
}

func (pool *ReloadingAccountPool) Configured(ctx context.Context) (bool, error) {
	if pool == nil || pool.loader == nil || pool.quota == nil {
		return false, nil
	}
	outreachCfg, err := pool.loader.Load(ctx)
	if err != nil {
		return false, fmt.Errorf("load outreach email accounts: %w", err)
	}
	return len(outreachCfg.GoogleWorkspaceAccounts) > 0, nil
}

func (pool *ReloadingAccountPool) Exhausted() bool {
	return pool == nil || pool.loader == nil || pool.quota == nil
}

func (pool *ReloadingAccountPool) NextAvailableAt(ctx context.Context) (*time.Time, error) {
	current, err := pool.current(ctx)
	if err != nil {
		return nil, err
	}
	return current.NextAvailableAt(ctx)
}

func (pool *ReloadingAccountPool) Send(ctx context.Context, request SendRequest) (SendResult, error) {
	current, err := pool.current(ctx)
	if err != nil {
		return SendResult{}, err
	}
	return current.Send(ctx, request)
}

func (pool *ReloadingAccountPool) SendDirect(ctx context.Context, request SendRequest) (SendResult, error) {
	current, err := pool.current(ctx)
	if err != nil {
		return SendResult{}, err
	}
	return current.SendDirect(ctx, request)
}

func (pool *ReloadingAccountPool) SendDirectFrom(ctx context.Context, accountKey string, request SendRequest) (SendResult, error) {
	current, err := pool.current(ctx)
	if err != nil {
		return SendResult{}, err
	}
	return current.SendDirectFrom(ctx, accountKey, request)
}

func (pool *ReloadingAccountPool) current(ctx context.Context) (*AccountPool, error) {
	if pool == nil || pool.loader == nil || pool.quota == nil {
		return nil, fmt.Errorf("outreach email account pool is not configured")
	}
	outreachCfg, err := pool.loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load outreach email accounts: %w", err)
	}
	return buildAccountPool(ctx, pool.emailCfg, outreachCfg, pool.quota)
}

var _ AccountPoolProvider = (*ReloadingAccountPool)(nil)
