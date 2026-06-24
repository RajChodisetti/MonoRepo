package restaurants

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

type Mock struct {
	Restaurants map[uuid.UUID]Restaurant
}

func (mock *Mock) GetByID(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	record, ok := mock.Restaurants[id]
	if !ok {
		return Restaurant{}, repository.ErrNotFound
	}
	return record, nil
}

func (mock *Mock) List(ctx context.Context, filter ListFilter) ([]Restaurant, error) {
	return mock.filterRecords(nil, filter), nil
}

func (mock *Mock) ListByIDs(ctx context.Context, ids []uuid.UUID, filter ListFilter) ([]Restaurant, error) {
	return mock.filterRecords(ids, filter), nil
}

func (mock *Mock) filterRecords(ids []uuid.UUID, filter ListFilter) []Restaurant {
	records := make([]Restaurant, 0, len(mock.Restaurants))
	for _, record := range mock.Restaurants {
		if len(ids) > 0 {
			matched := false
			for _, id := range ids {
				if id == record.ID {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if MatchesFilter(record, filter) {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}

func (mock *Mock) Create(ctx context.Context, input CreateInput) (Restaurant, error) {
	now := time.Now()
	record := Restaurant{
		ID:        uuid.New(),
		Name:      input.Name,
		Email:     input.Email,
		Status:    StatusLead,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if mock.Restaurants == nil {
		mock.Restaurants = make(map[uuid.UUID]Restaurant)
	}
	mock.Restaurants[record.ID] = record
	return record, nil
}

func (mock *Mock) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Restaurant, error) {
	record, ok := mock.Restaurants[id]
	if !ok {
		return Restaurant{}, repository.ErrNotFound
	}
	updated := ApplyUpdateInput(record, input)
	updated.UpdatedAt = time.Now()
	mock.Restaurants[id] = updated
	return updated, nil
}

func (mock *Mock) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (Restaurant, error) {
	statusValue := status
	return mock.Update(ctx, id, UpdateInput{Status: &statusValue})
}

func (mock *Mock) Archive(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	return mock.UpdateStatus(ctx, id, StatusArchived)
}

func (mock *Mock) MarkShownInterest(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	record, ok := mock.Restaurants[id]
	if !ok {
		return Restaurant{}, repository.ErrNotFound
	}
	shownInterest := true
	status := StatusAfterShownInterest(record.Status)
	return mock.Update(ctx, id, UpdateInput{
		ShownInterest: &shownInterest,
		Status:        &status,
	})
}

type MembershipMock struct {
	Members map[uuid.UUID]map[uuid.UUID]bool
}

func (mock *MembershipMock) HasMembership(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error) {
	restaurants, ok := mock.Members[userID]
	if !ok {
		return false, nil
	}
	return restaurants[restaurantID], nil
}

func (mock *MembershipMock) ListRestaurantIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	restaurants, ok := mock.Members[userID]
	if !ok {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(restaurants))
	for restaurantID, allowed := range restaurants {
		if allowed {
			ids = append(ids, restaurantID)
		}
	}
	return ids, nil
}

func (mock *MembershipMock) ListMembersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]Member, error) {
	var members []Member
	for userID, restaurants := range mock.Members {
		if !restaurants[restaurantID] {
			continue
		}
		members = append(members, Member{
			ID:           uuid.New(),
			RestaurantID: restaurantID,
			UserID:       userID,
			MemberRole:   "owner",
			CreatedAt:    time.Now(),
		})
	}
	return members, nil
}

func (mock *MembershipMock) AddMember(ctx context.Context, restaurantID, userID uuid.UUID, memberRole string) (Member, error) {
	if mock.Members == nil {
		mock.Members = make(map[uuid.UUID]map[uuid.UUID]bool)
	}
	if mock.Members[userID] == nil {
		mock.Members[userID] = make(map[uuid.UUID]bool)
	}
	mock.Members[userID][restaurantID] = true
	return Member{
		ID:           uuid.New(),
		RestaurantID: restaurantID,
		UserID:       userID,
		MemberRole:   memberRole,
		CreatedAt:    time.Now(),
	}, nil
}

var (
	_ Repository           = (*Mock)(nil)
	_ MembershipRepository = (*MembershipMock)(nil)
)
