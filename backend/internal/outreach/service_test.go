package outreach_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type mockRepo struct {
	count         int
	leads         []outreach.EligibleLead
	delivery      outreach.SequenceDelivery
	prepared      outreach.RenderedSequenceStep
	finalizations []outreach.SequenceDeliveryFinalization
	nextDue       *time.Time
}

func (repo *mockRepo) ListEligibleLeads(context.Context, int) ([]outreach.EligibleLead, error) {
	return repo.leads, nil
}

func (repo *mockRepo) CountEligibleLeads(context.Context) (int, error) { return repo.count, nil }

func (repo *mockRepo) GetSequenceDelivery(context.Context, uuid.UUID, int) (outreach.SequenceDelivery, error) {
	return repo.delivery, nil
}

func (repo *mockRepo) PrepareSequenceDelivery(_ context.Context, _ uuid.UUID, step int, subject, bodyText string) error {
	repo.prepared = outreach.RenderedSequenceStep{Position: step, Subject: subject, BodyText: bodyText}
	return nil
}

func (repo *mockRepo) FinalizeSequenceDelivery(_ context.Context, finalization outreach.SequenceDeliveryFinalization) error {
	repo.finalizations = append(repo.finalizations, finalization)
	return nil
}

func (repo *mockRepo) NextSequenceDueAt(context.Context) (*time.Time, error) {
	return repo.nextDue, nil
}

type mockEmailProvider struct {
	request emailprovider.SendRequest
	result  emailprovider.SendResult
	err     error
}

func (provider *mockEmailProvider) Send(_ context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error) {
	provider.request = req
	return provider.result, provider.err
}

func testAccountPool(t *testing.T, provider emailprovider.Provider) *emailprovider.AccountPool {
	t.Helper()
	pool, err := emailprovider.NewAccountPool([]emailprovider.Provider{provider}, 50, 150)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	return pool
}

func newSequenceService(t *testing.T, repo *mockRepo, provider *mockEmailProvider) *outreach.Service {
	t.Helper()
	campaignRepo := &campaigns.Mock{}
	campaignService := campaigns.NewService(campaignRepo, nil, nil, nil, config.AppURLsConfig{
		PublicBaseURL: "https://api.example.com",
	})
	return outreach.NewService(
		repo,
		nil,
		campaignRepo,
		campaignService,
		nil,
		outreach.DemoTokenResolver{},
		testAccountPool(t, provider),
		nil,
		config.EmailConfig{Provider: "fake"},
		config.OutreachConfig{BulkMax: 150},
		nil,
		nil,
	)
}

func eligibleSequenceRepo() *mockRepo {
	restaurantID := uuid.New()
	campaignID := uuid.New()
	consentAt := time.Now().UTC()
	step := outreach.SequenceStep{
		ID: uuid.New(), SequenceID: uuid.New(), Position: 1, Enabled: true,
		SubjectTemplate:  "A practical idea for {{restaurant_name}}",
		BodyTextTemplate: "{{greeting}}\n\nLearn more: {{website_url}}\n\nOpt out: {{unsubscribe_url}}",
	}
	return &mockRepo{
		leads: []outreach.EligibleLead{{CampaignID: campaignID, RestaurantID: restaurantID, Step: 1}},
		delivery: outreach.SequenceDelivery{
			CampaignID: campaignID, RestaurantID: restaurantID,
			RestaurantName: "Test Cafe", RecipientEmail: "owner@example.com",
			LifecycleStatus: restaurants.StatusLead,
			ConsentBasis:    "inferred_business", ConsentSource: "business_contact_import",
			ConsentRecordedAt: &consentAt, ConsentEvidence: []byte(`{"policy":"inferred_business"}`),
			SequenceStatus: outreach.SequenceStatusApproved, Step: step,
		},
	}
}

func TestRunBulkSendFinalizesAcceptedSequenceAndSendsPlainText(t *testing.T) {
	repo := eligibleSequenceRepo()
	provider := &mockEmailProvider{result: emailprovider.SendResult{ProviderMessageID: "mock"}}
	service := newSequenceService(t, repo, provider)

	summary, err := service.RunBulkSend(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("RunBulkSend() error = %v", err)
	}
	if summary.Sent != 1 || summary.Attempted != 1 {
		t.Fatalf("summary = %#v, want one sent attempt", summary)
	}
	if provider.request.HTMLBody != "" {
		t.Fatalf("HTMLBody = %q, want empty", provider.request.HTMLBody)
	}
	if got := len(strings.FieldsFunc(provider.request.TextBody, func(r rune) bool { return r == '\n' })); got == 0 {
		t.Fatal("TextBody is empty")
	}
	if strings.Count(provider.request.TextBody, "https://") != 2 {
		t.Fatalf("TextBody = %q, want exactly two links", provider.request.TextBody)
	}
	if len(repo.finalizations) != 1 || repo.finalizations[0].Outcome != "sent" || repo.finalizations[0].Step != 1 {
		t.Fatalf("finalizations = %#v, want confirmed step 1", repo.finalizations)
	}
}

func TestRunBulkSendSkipDoesNotAdvanceSequence(t *testing.T) {
	repo := eligibleSequenceRepo()
	provider := &mockEmailProvider{result: emailprovider.SendResult{Skipped: true}}
	service := newSequenceService(t, repo, provider)

	summary, err := service.RunBulkSend(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("RunBulkSend() error = %v", err)
	}
	if summary.Skipped != 1 || len(repo.finalizations) != 1 || repo.finalizations[0].Outcome != "skipped" {
		t.Fatalf("summary/finalizations = %#v / %#v", summary, repo.finalizations)
	}
}

func TestRunBulkSendProviderFailureRetainsSequenceAsUnknown(t *testing.T) {
	repo := eligibleSequenceRepo()
	provider := &mockEmailProvider{err: errors.New("provider outcome unknown")}
	service := newSequenceService(t, repo, provider)

	_, err := service.RunBulkSend(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("RunBulkSend() error = nil, want provider failure")
	}
	if len(repo.finalizations) != 1 || repo.finalizations[0].Outcome != "unknown" || repo.finalizations[0].Step != 1 {
		t.Fatalf("finalizations = %#v, want unknown without advancement", repo.finalizations)
	}
}
