package scrapejobs

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

var supportedCities = map[string]string{
	"adelaide":  "Adelaide",
	"brisbane":  "Brisbane",
	"melbourne": "Melbourne",
	"perth":     "Perth",
	"sydney":    "Sydney",
}

var supportedNiches = map[string]struct{}{
	"restaurant": {},
	"dentist":    {},
	"plumber":    {},
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) Trigger(ctx context.Context, principal auth.Principal, input CreateInput) (TriggerResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return TriggerResult{}, ErrForbidden
	}

	cityKey := normalizeCity(input.City)
	city, ok := supportedCities[cityKey]
	if !ok {
		return TriggerResult{}, ErrInvalidCity
	}

	niche := strings.ToLower(strings.TrimSpace(input.Niche))
	if niche == "" {
		niche = DefaultNiche
	}
	if _, ok := supportedNiches[niche]; !ok {
		return TriggerResult{}, ErrInvalidNiche
	}

	job, created, err := service.repo.CreateOrGetActive(ctx, city, cityKey, niche, principal.UserID)
	if err != nil {
		return TriggerResult{}, err
	}
	return TriggerResult{Created: created, Job: job}, nil
}

func (service *Service) Get(ctx context.Context, principal auth.Principal, id uuid.UUID) (Job, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Job{}, ErrForbidden
	}
	return service.repo.GetByID(ctx, id)
}

func (service *Service) List(ctx context.Context, principal auth.Principal, limit int) ([]Job, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, ErrForbidden
	}
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	return service.repo.ListRecent(ctx, limit)
}

func (service *Service) RetryFailed(ctx context.Context, principal auth.Principal, id uuid.UUID) (Job, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Job{}, ErrForbidden
	}
	return service.repo.RetryFailed(ctx, id)
}

func normalizeCity(value string) string {
	value = strings.TrimSpace(value)
	if comma := strings.IndexByte(value, ','); comma >= 0 {
		value = value[:comma]
	}
	return strings.ToLower(strings.TrimSpace(value))
}
