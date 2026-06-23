package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("inactive user")
	ErrInvalidInput       = errors.New("invalid input")
)

const minPasswordLength = 8

type Service struct {
	users  Repository
	tokens *TokenManager
}

func NewService(users Repository, tokens *TokenManager) *Service {
	return &Service{
		users:  users,
		tokens: tokens,
	}
}

type SignupInput struct {
	Email    string
	Password string
	FullName string
	Role     string
}

type LoginInput struct {
	Email    string
	Password string
}

type UserView struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	Role     string    `json:"role"`
}

type AuthResult struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
	User        UserView `json:"user"`
}

func (service *Service) Signup(ctx context.Context, input SignupInput) (AuthResult, error) {
	if err := validateSignupInput(input); err != nil {
		return AuthResult{}, err
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	record, err := service.users.Create(ctx, CreateInput{
		Email:        strings.TrimSpace(input.Email),
		PasswordHash: passwordHash,
		FullName:     strings.TrimSpace(input.FullName),
		Role:         input.Role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return AuthResult{}, fmt.Errorf("email already exists: %w", repository.ErrConflict)
		}
		return AuthResult{}, err
	}

	return service.issueAuthResult(record)
}

func (service *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	record, err := service.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if !record.IsActive {
		return AuthResult{}, ErrInactiveUser
	}

	if err := CheckPassword(record.PasswordHash, input.Password); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return service.issueAuthResult(record)
}

func (service *Service) issueAuthResult(record User) (AuthResult, error) {
	token, expiresAt, err := service.tokens.IssueToken(record.ID, record.Email, record.Role)
	if err != nil {
		return AuthResult{}, err
	}

	expiresIn := int64(time.Until(expiresAt).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	return AuthResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User: UserView{
			ID:       record.ID,
			Email:    record.Email,
			FullName: record.FullName,
			Role:     record.Role,
		},
	}, nil
}

func validateSignupInput(input SignupInput) error {
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("%w: email is invalid", ErrInvalidInput)
	}
	if len(input.Password) < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters", ErrInvalidInput, minPasswordLength)
	}
	if !ValidRole(strings.TrimSpace(input.Role)) {
		return fmt.Errorf("%w: role is invalid", ErrInvalidInput)
	}
	return nil
}
