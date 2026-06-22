package user

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/repository"
)

type Mock struct {
	Users    map[uuid.UUID]User
	ByEmail  map[string]uuid.UUID
	CreateFn func(ctx context.Context, input CreateInput) (User, error)
}

func (mock *Mock) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	record, ok := mock.Users[id]
	if !ok {
		return User{}, repository.ErrNotFound
	}
	return record, nil
}

func (mock *Mock) GetByEmail(ctx context.Context, email string) (User, error) {
	id, ok := mock.ByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, repository.ErrNotFound
	}
	return mock.GetByID(ctx, id)
}

func (mock *Mock) Create(ctx context.Context, input CreateInput) (User, error) {
	if mock.CreateFn != nil {
		return mock.CreateFn(ctx, input)
	}
	return User{}, repository.ErrConflict
}

var _ Repository = (*Mock)(nil)
