package restaurants

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestRestaurantOwnerCannotAccessAnotherRestaurant(t *testing.T) {
	ownerID := uuid.New()
	ownedRestaurantID := uuid.New()
	otherRestaurantID := uuid.New()

	service := NewService(
		&Mock{
			Restaurants: map[uuid.UUID]Restaurant{
				ownedRestaurantID: {ID: ownedRestaurantID, Name: "Owned Cafe"},
				otherRestaurantID: {ID: otherRestaurantID, Name: "Other Cafe"},
			},
		},
		&MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedRestaurantID: true},
			},
		},
	)

	principal := auth.Principal{
		UserID: ownerID,
		Email:  "owner@example.com",
		Role:   auth.RoleRestaurantOwner,
	}

	if err := service.CanAccessRestaurant(context.Background(), principal, ownedRestaurantID); err != nil {
		t.Fatalf("CanAccessRestaurant(owned) error = %v, want nil", err)
	}

	if err := service.CanAccessRestaurant(context.Background(), principal, otherRestaurantID); err == nil {
		t.Fatal("CanAccessRestaurant(other) error = nil, want forbidden")
	} else if !isForbidden(err) {
		t.Fatalf("CanAccessRestaurant(other) error = %v, want ErrForbidden", err)
	}
}

func TestInternalAdminCanAccessAnyRestaurant(t *testing.T) {
	restaurantID := uuid.New()
	service := NewService(
		&Mock{
			Restaurants: map[uuid.UUID]Restaurant{
				restaurantID: {ID: restaurantID, Name: "Any Cafe", Status: StatusLead, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
		},
		&MembershipMock{},
	)

	principal := auth.Principal{
		UserID: uuid.New(),
		Email:  "admin@example.com",
		Role:   auth.RoleInternalAdmin,
	}

	if err := service.CanAccessRestaurant(context.Background(), principal, restaurantID); err != nil {
		t.Fatalf("CanAccessRestaurant() error = %v, want nil", err)
	}

	record, err := service.GetRestaurant(context.Background(), principal, restaurantID)
	if err != nil {
		t.Fatalf("GetRestaurant() error = %v", err)
	}
	if record.Name != "Any Cafe" {
		t.Fatalf("Name = %q, want Any Cafe", record.Name)
	}
}

func TestRestaurantOwnerListAccessibleRestaurants(t *testing.T) {
	ownerID := uuid.New()
	ownedRestaurantID := uuid.New()
	otherRestaurantID := uuid.New()

	service := NewService(
		&Mock{
			Restaurants: map[uuid.UUID]Restaurant{
				ownedRestaurantID: {ID: ownedRestaurantID, Name: "Owned Cafe"},
				otherRestaurantID: {ID: otherRestaurantID, Name: "Other Cafe"},
			},
		},
		&MembershipMock{
			Members: map[uuid.UUID]map[uuid.UUID]bool{
				ownerID: {ownedRestaurantID: true},
			},
		},
	)

	principal := auth.Principal{
		UserID: ownerID,
		Email:  "owner@example.com",
		Role:   auth.RoleRestaurantOwner,
	}

	records, err := service.ListAccessibleRestaurants(context.Background(), principal, ListFilter{})
	if err != nil {
		t.Fatalf("ListAccessibleRestaurants() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0].ID != ownedRestaurantID {
		t.Fatalf("records[0].ID = %s, want %s", records[0].ID, ownedRestaurantID)
	}
}

func TestInternalAdminListsAllRestaurants(t *testing.T) {
	restaurantA := uuid.New()
	restaurantB := uuid.New()

	service := NewService(
		&Mock{
			Restaurants: map[uuid.UUID]Restaurant{
				restaurantA: {ID: restaurantA, Name: "A", Status: StatusLead},
				restaurantB: {ID: restaurantB, Name: "B", Status: StatusLead},
			},
		},
		&MembershipMock{},
	)

	principal := auth.Principal{
		UserID: uuid.New(),
		Email:  "admin@example.com",
		Role:   auth.RoleInternalAdmin,
	}

	records, err := service.ListAccessibleRestaurants(context.Background(), principal, ListFilter{})
	if err != nil {
		t.Fatalf("ListAccessibleRestaurants() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}
}

func isForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func TestUpdateRestaurantSetsEmailedWhenContacted(t *testing.T) {
	restaurantID := uuid.New()
	service := NewService(
		&Mock{
			Restaurants: map[uuid.UUID]Restaurant{
				restaurantID: {
					ID:     restaurantID,
					Name:   "Lead Cafe",
					Email:  "lead@example.com",
					Status: StatusLead,
				},
			},
		},
		&MembershipMock{},
	)

	principal := auth.Principal{
		UserID: uuid.New(),
		Email:  "admin@example.com",
		Role:   auth.RoleInternalAdmin,
	}

	contacted := true
	record, err := service.UpdateRestaurant(context.Background(), principal, restaurantID, UpdateInput{
		IsContacted: &contacted,
	})
	if err != nil {
		t.Fatalf("UpdateRestaurant() error = %v", err)
	}
	if !record.IsContacted {
		t.Fatal("IsContacted = false, want true")
	}
	if record.Status != StatusEmailed {
		t.Fatalf("Status = %q, want %q", record.Status, StatusEmailed)
	}
}

func TestListFilterRejectsInvalidStatus(t *testing.T) {
	service := NewService(&Mock{}, &MembershipMock{})
	principal := auth.Principal{Role: auth.RoleInternalAdmin}

	_, err := service.ListAccessibleRestaurants(context.Background(), principal, ListFilter{
		Status: "not-a-status",
	})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("error = %v, want ErrInvalidStatus", err)
	}
}
