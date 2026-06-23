package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

func TestSignupIssuesTokenWithRoleFromRepository(t *testing.T) {
	userID := uuid.New()
	repo := &Mock{
		CreateFn: func(ctx context.Context, input CreateInput) (User, error) {
			return User{
				ID:           userID,
				Email:        input.Email,
				PasswordHash: input.PasswordHash,
				FullName:     input.FullName,
				Role:         input.Role,
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	service := NewService(repo, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	result, err := service.Signup(context.Background(), SignupInput{
		Email:    "owner@example.com",
		Password: "password123",
		FullName: "Owner",
		Role:     RoleRestaurantOwner,
	})
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if result.User.Role != RoleRestaurantOwner {
		t.Fatalf("User.Role = %q, want %q", result.User.Role, RoleRestaurantOwner)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestSignupRejectsInvalidEmail(t *testing.T) {
	service := NewService(&Mock{}, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	_, err := service.Signup(context.Background(), SignupInput{
		Email:    "not-an-email",
		Password: "password123",
		FullName: "Owner",
		Role:     RoleRestaurantOwner,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Signup() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestSignupRejectsShortPassword(t *testing.T) {
	service := NewService(&Mock{}, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	_, err := service.Signup(context.Background(), SignupInput{
		Email:    "owner@example.com",
		Password: "short",
		FullName: "Owner",
		Role:     RoleRestaurantOwner,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Signup() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestLoginReturnsTokenForValidCredentials(t *testing.T) {
	userID := uuid.New()
	password := "password123"
	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	repo := &Mock{
		Users: map[uuid.UUID]User{
			userID: {
				ID:           userID,
				Email:        "owner@example.com",
				PasswordHash: passwordHash,
				FullName:     "Owner",
				Role:         RoleRestaurantOwner,
				IsActive:     true,
			},
		},
		ByEmail: map[string]uuid.UUID{
			"owner@example.com": userID,
		},
	}

	service := NewService(repo, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	result, err := service.Login(context.Background(), LoginInput{
		Email:    "owner@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestLoginRejectsInactiveUser(t *testing.T) {
	userID := uuid.New()
	passwordHash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	repo := &Mock{
		Users: map[uuid.UUID]User{
			userID: {
				ID:           userID,
				Email:        "inactive@example.com",
				PasswordHash: passwordHash,
				FullName:     "Inactive",
				Role:         RoleRestaurantOwner,
				IsActive:     false,
			},
		},
		ByEmail: map[string]uuid.UUID{
			"inactive@example.com": userID,
		},
	}

	service := NewService(repo, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	_, err = service.Login(context.Background(), LoginInput{
		Email:    "inactive@example.com",
		Password: "password123",
	})
	if !errors.Is(err, ErrInactiveUser) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInactiveUser)
	}
}

func TestSignupReturnsConflictForDuplicateEmail(t *testing.T) {
	repo := &Mock{
		CreateFn: func(ctx context.Context, input CreateInput) (User, error) {
			return User{}, repository.ErrConflict
		},
	}

	service := NewService(repo, NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour))
	_, err := service.Signup(context.Background(), SignupInput{
		Email:    "owner@example.com",
		Password: "password123",
		FullName: "Owner",
		Role:     RoleRestaurantOwner,
	})
	if !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("Signup() error = %v, want %v", err, repository.ErrConflict)
	}
	if !strings.Contains(err.Error(), "email already exists") {
		t.Fatalf("Signup() error = %q, want duplicate email message", err.Error())
	}
}
