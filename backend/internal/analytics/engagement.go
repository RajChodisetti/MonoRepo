package analytics

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

const (
	maxSessionDuration  = 24 * time.Hour
	maxTranscriptLength = 4000
)

var (
	ErrSessionNotFound = errors.New("demo session not found")
	ErrInvalidEvent    = errors.New("invalid demo engagement event")
)

type Transcript struct {
	ID         uuid.UUID `json:"id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Session struct {
	ID              uuid.UUID    `json:"id"`
	DemoSiteID      *uuid.UUID   `json:"demo_site_id,omitempty"`
	RestaurantID    uuid.UUID    `json:"restaurant_id"`
	TemplateID      string       `json:"template_id"`
	StartedAt       time.Time    `json:"started_at"`
	LastSeenAt      time.Time    `json:"last_seen_at"`
	EndedAt         *time.Time   `json:"ended_at,omitempty"`
	DurationSeconds int          `json:"duration_seconds"`
	Transcript      []Transcript `json:"transcript"`
}

type StartResult struct {
	SessionID    uuid.UUID `json:"session_id"`
	SessionToken string    `json:"session_token"`
}

type Repository interface {
	CreateSession(ctx context.Context, demoSiteID *uuid.UUID, restaurantID uuid.UUID, templateID, sessionTokenHash string) (Session, error)
	TouchSession(ctx context.Context, sessionID uuid.UUID, sessionToken string, activeSeconds int, ended bool) error
	AddTranscript(ctx context.Context, sessionID uuid.UUID, sessionToken, role, content string) error
	ListSessions(ctx context.Context, restaurantID uuid.UUID) ([]Session, error)
}

type Service struct {
	demos *demos.Service
	repo  Repository
}

func NewService(demoService *demos.Service, repo Repository) *Service {
	return &Service{demos: demoService, repo: repo}
}

func (service *Service) StartSession(ctx context.Context, slug, demoToken, templateID string) (StartResult, error) {
	if service == nil || service.demos == nil || service.repo == nil {
		return StartResult{}, fmt.Errorf("demo engagement is not configured")
	}
	site, err := service.demos.ResolvePublicSite(ctx, slug, demoToken)
	if err != nil {
		return StartResult{}, demos.ErrDemoNotFound
	}
	token, err := generateSessionToken()
	if err != nil {
		return StartResult{}, err
	}
	hash, err := demos.HashDemoToken(token)
	if err != nil {
		return StartResult{}, fmt.Errorf("hash demo session token: %w", err)
	}
	templateID, err = validateTemplateID(templateID)
	if err != nil {
		return StartResult{}, err
	}
	session, err := service.repo.CreateSession(ctx, &site.ID, site.RestaurantID, templateID, hash)
	if err != nil {
		return StartResult{}, err
	}
	return StartResult{SessionID: session.ID, SessionToken: token}, nil
}

func (service *Service) StartAdminPreview(ctx context.Context, restaurantID uuid.UUID, templateID string) (StartResult, error) {
	if service == nil || service.repo == nil || restaurantID == uuid.Nil {
		return StartResult{}, fmt.Errorf("demo engagement is not configured")
	}
	templateID, err := validateTemplateID(templateID)
	if err != nil {
		return StartResult{}, err
	}
	token, err := generateSessionToken()
	if err != nil {
		return StartResult{}, err
	}
	hash, err := demos.HashDemoToken(token)
	if err != nil {
		return StartResult{}, fmt.Errorf("hash demo session token: %w", err)
	}
	session, err := service.repo.CreateSession(ctx, nil, restaurantID, templateID, hash)
	if err != nil {
		return StartResult{}, err
	}
	return StartResult{SessionID: session.ID, SessionToken: token}, nil
}

func (service *Service) Touch(ctx context.Context, sessionID uuid.UUID, sessionToken string, activeSeconds int, ended bool) error {
	if sessionID == uuid.Nil || strings.TrimSpace(sessionToken) == "" || activeSeconds < 0 || activeSeconds > int(maxSessionDuration/time.Second) {
		return ErrSessionNotFound
	}
	return service.repo.TouchSession(ctx, sessionID, strings.TrimSpace(sessionToken), activeSeconds, ended)
}

func (service *Service) AddTranscript(ctx context.Context, sessionID uuid.UUID, sessionToken, role, content string) error {
	role = strings.ToLower(strings.TrimSpace(role))
	content = strings.TrimSpace(content)
	if role != "user" && role != "assistant" && role != "system" {
		return ErrInvalidEvent
	}
	if content == "" || len(content) > maxTranscriptLength {
		return ErrInvalidEvent
	}
	return service.repo.AddTranscript(ctx, sessionID, strings.TrimSpace(sessionToken), role, content)
}

func (service *Service) ListSessions(ctx context.Context, restaurantID uuid.UUID) ([]Session, error) {
	if restaurantID == uuid.Nil {
		return nil, ErrSessionNotFound
	}
	return service.repo.ListSessions(ctx, restaurantID)
}

func generateSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := cryptorand.Read(buf); err != nil {
		return "", fmt.Errorf("generate demo session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func validateTemplateID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value != "1" && value != "2" && value != "3" {
		return "", ErrInvalidEvent
	}
	return value, nil
}
