package email

import (
	"context"
	"fmt"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type ReloadingHealthService struct {
	emailCfg config.EmailConfig
	loader   OutreachConfigLoader
	store    HealthStore
}

func NewReloadingHealthService(
	emailCfg config.EmailConfig,
	loader OutreachConfigLoader,
	store HealthStore,
) *ReloadingHealthService {
	return &ReloadingHealthService{emailCfg: emailCfg, loader: loader, store: store}
}

func (service *ReloadingHealthService) RunDue(ctx context.Context) error {
	current, err := service.current(ctx)
	if err != nil {
		return err
	}
	return current.RunDue(ctx)
}

func (service *ReloadingHealthService) List(ctx context.Context) ([]HealthStatus, error) {
	current, err := service.current(ctx)
	if err != nil {
		return nil, err
	}
	return current.List(ctx)
}

func (service *ReloadingHealthService) current(ctx context.Context) (*HealthService, error) {
	if service == nil || service.loader == nil || service.store == nil {
		return nil, fmt.Errorf("outreach email health is not configured")
	}
	outreachCfg, err := service.loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load outreach email accounts for health: %w", err)
	}
	return NewHealthServiceFromConfig(ctx, service.emailCfg, outreachCfg, service.store)
}

var _ HealthMonitor = (*ReloadingHealthService)(nil)
