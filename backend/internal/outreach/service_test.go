package outreach_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type mockRepo struct {
	count int
}

func (repo *mockRepo) ListEligibleLeads(ctx context.Context, limit int) ([]outreach.EligibleLead, error) {
	return nil, nil
}

func (repo *mockRepo) CountEligibleLeads(ctx context.Context) (int, error) {
	return repo.count, nil
}

type mockEnqueuer struct {
	jobID string
	err   error
}

func (enqueuer *mockEnqueuer) EnqueueBulkSend(ctx context.Context, triggeredBy uuid.UUID) (string, error) {
	if enqueuer.err != nil {
		return "", enqueuer.err
	}
	return enqueuer.jobID, nil
}

type mockEmailProvider struct{}

func (provider *mockEmailProvider) Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	return emailprovider.SendResult{ProviderMessageID: "mock"}, nil
}

func testAccountPool(t *testing.T) *emailprovider.AccountPool {
	t.Helper()
	pool, err := emailprovider.NewAccountPool([]emailprovider.Provider{&mockEmailProvider{}}, 50, 150)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	return pool
}

func TestTriggerBulkSendRequiresConfiguredAccounts(t *testing.T) {
	service := outreach.NewService(
		&mockRepo{count: 3},
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		nil,
		config.EmailConfig{},
		config.OutreachConfig{BulkMax: 150},
		&mockEnqueuer{jobID: "job-1"},
		nil,
	)

	_, err := service.TriggerBulkSend(context.Background(), auth.Principal{
		Role:   auth.RoleInternalAdmin,
		UserID: uuid.New(),
	})
	if !errors.Is(err, outreach.ErrNotConfigured) {
		t.Fatalf("TriggerBulkSend() error = %v, want ErrNotConfigured", err)
	}
}

func TestTriggerBulkSendEnqueuesJob(t *testing.T) {
	service := outreach.NewService(
		&mockRepo{count: 7},
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		testAccountPool(t),
		config.EmailConfig{},
		config.OutreachConfig{
			BulkMax:      150,
			ZohoAccounts: []config.ZohoMailConfig{{AccountID: "1", ClientID: "a", ClientSecret: "b", RefreshToken: "c"}},
		},
		&mockEnqueuer{jobID: "job-123"},
		nil,
	)

	result, err := service.TriggerBulkSend(context.Background(), auth.Principal{
		Role:   auth.RoleInternalAdmin,
		UserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("TriggerBulkSend() error = %v, want nil", err)
	}
	if result.JobID != "job-123" {
		t.Fatalf("JobID = %q, want job-123", result.JobID)
	}
	if result.PendingEligibleCount != 7 {
		t.Fatalf("PendingEligibleCount = %d, want 7", result.PendingEligibleCount)
	}
}
