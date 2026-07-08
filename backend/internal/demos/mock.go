package demos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	Sites map[string]Site
}

func (mock *Mock) GetBySlug(ctx context.Context, slug string) (Site, error) {
	record, ok := mock.Sites[slug]
	if !ok {
		return Site{}, repository.ErrNotFound
	}
	return record, nil
}

func (mock *Mock) GetByID(ctx context.Context, id uuid.UUID) (Site, error) {
	if mock.Sites == nil {
		return Site{}, repository.ErrNotFound
	}
	for _, record := range mock.Sites {
		if record.ID == id {
			return record, nil
		}
	}
	return Site{}, repository.ErrNotFound
}

func (mock *Mock) Create(ctx context.Context, input CreateInput) (Site, error) {
	now := time.Now()
	record := Site{
		ID:            uuid.New(),
		RestaurantID:  input.RestaurantID,
		Slug:          input.Slug,
		TokenHash:     input.TokenHash,
		Status:        input.Status,
		PublicPayload: input.PublicPayload,
		ExpiresAt:     input.ExpiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if mock.Sites == nil {
		mock.Sites = make(map[string]Site)
	}
	mock.Sites[input.Slug] = record
	return record, nil
}

func (mock *Mock) UpdateTokenHash(ctx context.Context, id uuid.UUID, tokenHash string) error {
	if mock.Sites == nil {
		return repository.ErrNotFound
	}
	for slug, record := range mock.Sites {
		if record.ID == id {
			record.TokenHash = tokenHash
			record.UpdatedAt = time.Now().UTC()
			mock.Sites[slug] = record
			return nil
		}
	}
	return repository.ErrNotFound
}

var _ Repository = (*Mock)(nil)

func DefaultPublicPayload() json.RawMessage {
	return json.RawMessage(`{
		"restaurant_name": "Sample Cafe",
		"cuisine": "Thai",
		"hero": "Welcome to Sample Cafe",
		"hours": {"monday": "9am-9pm"},
		"address": "123 Main St",
		"phone": "+1-555-0100",
		"menu_sections": [{"name": "Mains", "items": ["Pad Thai"]}],
		"reservation_cta": "Book a table",
		"ai_receptionist_cta": "Call our AI assistant"
	}`)
}
