package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/repository"
)

func TestMockGetByEmail(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	mock := &Mock{
		Users: map[uuid.UUID]User{
			userID: {
				ID:    userID,
				Email: "admin@example.com",
				Role:  auth.RoleAdmin,
			},
		},
		ByEmail: map[string]uuid.UUID{
			"admin@example.com": userID,
		},
	}

	record, err := mock.GetByEmail(context.Background(), "admin@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if record.Role != auth.RoleAdmin {
		t.Fatalf("Role = %q, want %q", record.Role, auth.RoleAdmin)
	}
}

func TestMockGetByEmailNotFound(t *testing.T) {
	t.Parallel()

	mock := &Mock{
		Users:   map[uuid.UUID]User{},
		ByEmail: map[string]uuid.UUID{},
	}

	_, err := mock.GetByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("GetByEmail() error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestMockCreateUsesCallback(t *testing.T) {
	t.Parallel()

	mock := &Mock{
		CreateFn: func(ctx context.Context, input CreateInput) (User, error) {
			return User{
				ID:        uuid.New(),
				Email:     input.Email,
				FullName:  input.FullName,
				Role:      input.Role,
				IsActive:  true,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		},
	}

	record, err := mock.Create(context.Background(), CreateInput{
		Email:        "dev@example.com",
		PasswordHash: "hash",
		FullName:     "Dev User",
		Role:         auth.RoleDeveloper,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if record.Email != "dev@example.com" {
		t.Fatalf("Email = %q, want dev@example.com", record.Email)
	}
}
