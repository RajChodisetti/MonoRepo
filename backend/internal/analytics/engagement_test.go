package analytics

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type engagementTestRepository struct {
	created      bool
	demoSiteID   *uuid.UUID
	restaurantID uuid.UUID
	templateID   string
}

func (repo *engagementTestRepository) CreateSession(
	_ context.Context,
	demoSiteID *uuid.UUID,
	restaurantID uuid.UUID,
	templateID, _ string,
) (Session, error) {
	repo.created = true
	repo.demoSiteID = demoSiteID
	repo.restaurantID = restaurantID
	repo.templateID = templateID
	return Session{ID: uuid.New(), RestaurantID: restaurantID, TemplateID: templateID}, nil
}

func (*engagementTestRepository) TouchSession(context.Context, uuid.UUID, string, int, bool) error {
	return nil
}

func (*engagementTestRepository) AddTranscript(context.Context, uuid.UUID, string, string, string) error {
	return nil
}

func (*engagementTestRepository) ListSessions(context.Context, uuid.UUID) ([]Session, error) {
	return nil, nil
}

func TestStartAdminPreviewRecordsSelectedTemplate(t *testing.T) {
	repo := &engagementTestRepository{}
	service := NewService(nil, repo)
	restaurantID := uuid.New()

	result, err := service.StartAdminPreview(context.Background(), restaurantID, "3")
	if err != nil {
		t.Fatalf("StartAdminPreview() error = %v", err)
	}
	if result.SessionID == uuid.Nil || result.SessionToken == "" {
		t.Fatalf("StartAdminPreview() result = %+v, want capability", result)
	}
	if !repo.created || repo.demoSiteID != nil || repo.restaurantID != restaurantID || repo.templateID != "3" {
		t.Fatalf("created session = %+v, want restaurant admin Elysian preview", repo)
	}
}

func TestStartAdminPreviewRejectsUnknownTemplate(t *testing.T) {
	repo := &engagementTestRepository{}
	service := NewService(nil, repo)

	if _, err := service.StartAdminPreview(context.Background(), uuid.New(), "4"); err == nil {
		t.Fatal("StartAdminPreview() error = nil, want invalid template")
	}
	if repo.created {
		t.Fatal("invalid template must not create a session")
	}
}
