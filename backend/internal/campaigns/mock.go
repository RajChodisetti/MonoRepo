package campaigns

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	mu        sync.Mutex
	Campaigns map[uuid.UUID]Campaign
	Events    []Event
	Tokens    map[string]TrackingToken
	Suppressed map[string]struct{}
	SendContexts map[uuid.UUID]SendContext
	RestaurantContexts map[uuid.UUID]SendContext
}

func (mock *Mock) Create(ctx context.Context, input CreateInput, draft DraftContent) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.Campaigns == nil {
		mock.Campaigns = make(map[uuid.UUID]Campaign)
	}
	now := time.Now().UTC()
	record := Campaign{
		ID:           uuid.New(),
		RestaurantID: input.RestaurantID,
		DemoSiteID:   input.DemoSiteID,
		CampaignType: input.CampaignType,
		Status:       StatusDraft,
		Subject:      draft.Subject,
		BodyHTML:     draft.BodyHTML,
		BodyText:     draft.BodyText,
		DemoToken:    input.DemoToken,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	mock.Campaigns[record.ID] = record
	return record, nil
}

func (mock *Mock) GetByID(ctx context.Context, id uuid.UUID) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Campaigns[id]
	if !ok {
		return Campaign{}, repository.ErrNotFound
	}
	return record, nil
}

func (mock *Mock) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	var records []Campaign
	for _, record := range mock.Campaigns {
		if record.RestaurantID == restaurantID {
			records = append(records, record)
		}
	}
	return records, nil
}

func (mock *Mock) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Campaigns[id]
	if !ok {
		return Campaign{}, repository.ErrNotFound
	}
	now := time.Now().UTC()
	record.Status = StatusApproved
	record.ApprovedAt = &now
	record.ApprovedBy = &approvedBy
	record.UpdatedAt = now
	mock.Campaigns[id] = record
	return record, nil
}

func (mock *Mock) MarkSending(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Campaigns[id]
	if !ok {
		return Campaign{}, repository.ErrNotFound
	}
	record.Status = StatusSending
	record.CurrentStep = step
	record.UpdatedAt = time.Now().UTC()
	mock.Campaigns[id] = record
	return record, nil
}

func (mock *Mock) MarkSent(ctx context.Context, id uuid.UUID, step int) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Campaigns[id]
	if !ok {
		return Campaign{}, repository.ErrNotFound
	}
	now := time.Now().UTC()
	record.Status = StatusSent
	record.CurrentStep = step
	record.LastSentAt = &now
	record.UpdatedAt = now
	mock.Campaigns[id] = record
	return record, nil
}

func (mock *Mock) Stop(ctx context.Context, id uuid.UUID, reason string) (Campaign, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Campaigns[id]
	if !ok {
		return Campaign{}, repository.ErrNotFound
	}
	record.Status = StatusStopped
	record.StoppedReason = reason
	record.UpdatedAt = time.Now().UTC()
	mock.Campaigns[id] = record
	return record, nil
}

func (mock *Mock) ListEvents(ctx context.Context, campaignID uuid.UUID) ([]Event, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	var records []Event
	for _, event := range mock.Events {
		if event.CampaignID == campaignID {
			records = append(records, event)
		}
	}
	return records, nil
}

func (mock *Mock) InsertEvent(ctx context.Context, campaignID, restaurantID uuid.UUID, eventType string, metadata json.RawMessage) error {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	mock.Events = append(mock.Events, Event{
		ID:           uuid.New(),
		CampaignID:   campaignID,
		RestaurantID: restaurantID,
		EventType:    eventType,
		Metadata:     metadata,
		EventTime:    time.Now().UTC(),
	})
	return nil
}

func (mock *Mock) CreateTrackingToken(ctx context.Context, token TrackingToken) error {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.Tokens == nil {
		mock.Tokens = make(map[string]TrackingToken)
	}
	mock.Tokens[token.Token] = token
	return nil
}

func (mock *Mock) GetTrackingToken(ctx context.Context, token string) (TrackingToken, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	record, ok := mock.Tokens[token]
	if !ok {
		return TrackingToken{}, repository.ErrNotFound
	}
	return record, nil
}

func (mock *Mock) IsSuppressed(ctx context.Context, email string) (bool, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.Suppressed == nil {
		return false, nil
	}
	_, ok := mock.Suppressed[email]
	return ok, nil
}

func (mock *Mock) AddSuppression(ctx context.Context, email, reason string) error {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.Suppressed == nil {
		mock.Suppressed = make(map[string]struct{})
	}
	mock.Suppressed[email] = struct{}{}
	return nil
}

func (mock *Mock) GetSendContext(ctx context.Context, campaignID uuid.UUID) (SendContext, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.SendContexts == nil {
		return SendContext{}, repository.ErrNotFound
	}
	ctxData, ok := mock.SendContexts[campaignID]
	if !ok {
		return SendContext{}, repository.ErrNotFound
	}
	return ctxData, nil
}

func (mock *Mock) GetRestaurantContext(ctx context.Context, restaurantID uuid.UUID) (SendContext, error) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.RestaurantContexts == nil {
		return SendContext{}, repository.ErrNotFound
	}
	ctxData, ok := mock.RestaurantContexts[restaurantID]
	if !ok {
		return SendContext{}, repository.ErrNotFound
	}
	return ctxData, nil
}

func (mock *Mock) MarkRestaurantEmailed(ctx context.Context, restaurantID uuid.UUID) error {
	return nil
}

var _ Repository = (*Mock)(nil)
