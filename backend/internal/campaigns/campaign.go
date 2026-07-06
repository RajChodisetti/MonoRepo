package campaigns

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	StatusDraft    = "draft"
	StatusApproved = "approved"
	StatusSending  = "sending"
	StatusSent     = "sent"
	StatusStopped  = "stopped"

	TypeOutreach = "outreach"

	EventSent         = "sent"
	EventOpened       = "opened"
	EventClicked      = "clicked"
	EventFailed       = "failed"
	EventUnsubscribed = "unsubscribed"

	TokenClick       = "click"
	TokenOpen        = "open"
	TokenUnsubscribe = "unsubscribe"
)

var (
	ErrNotEligible    = ErrEligibility("campaign is not eligible to send")
	ErrAlreadyStopped = ErrEligibility("campaign is stopped")
)

type ErrEligibility string

func (err ErrEligibility) Error() string { return string(err) }

type Campaign struct {
	ID            uuid.UUID
	RestaurantID  uuid.UUID
	DemoSiteID    uuid.UUID
	CampaignType  string
	Status        string
	CurrentStep   int
	Subject       string
	BodyHTML      string
	BodyText      string
	DemoToken     string
	ApprovedAt    *time.Time
	ApprovedBy    *uuid.UUID
	LastSentAt    *time.Time
	StoppedReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Event struct {
	ID           uuid.UUID
	CampaignID   uuid.UUID
	RestaurantID uuid.UUID
	EventType    string
	Metadata     json.RawMessage
	EventTime    time.Time
}

type TrackingToken struct {
	Token        string
	CampaignID   uuid.UUID
	RestaurantID uuid.UUID
	DemoSiteID   *uuid.UUID
	TokenType    string
	TargetURL    string
	ExpiresAt    *time.Time
	CreatedAt    time.Time
}

type CreateInput struct {
	RestaurantID uuid.UUID
	DemoSiteID   uuid.UUID
	DemoToken    string
	CampaignType string
}

type SendContext struct {
	RestaurantEmail string
	RestaurantName  string
	ReviewStatus    string
	DemoStatus      string
	DemoSlug        string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput, draft DraftContent) (Campaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (Campaign, error)
	ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Campaign, error)
	Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (Campaign, error)
	MarkSending(ctx context.Context, id uuid.UUID, step int) (Campaign, error)
	MarkSent(ctx context.Context, id uuid.UUID, step int) (Campaign, error)
	Stop(ctx context.Context, id uuid.UUID, reason string) (Campaign, error)
	ListEvents(ctx context.Context, campaignID uuid.UUID) ([]Event, error)
	InsertEvent(ctx context.Context, campaignID, restaurantID uuid.UUID, eventType string, metadata json.RawMessage) error
	CreateTrackingToken(ctx context.Context, token TrackingToken) error
	GetTrackingToken(ctx context.Context, token string) (TrackingToken, error)
	IsSuppressed(ctx context.Context, email string) (bool, error)
	AddSuppression(ctx context.Context, email, reason string) error
	GetSendContext(ctx context.Context, campaignID uuid.UUID) (SendContext, error)
	GetRestaurantContext(ctx context.Context, restaurantID uuid.UUID) (SendContext, error)
	MarkRestaurantEmailed(ctx context.Context, restaurantID uuid.UUID) error
}
