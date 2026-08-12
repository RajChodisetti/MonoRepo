package outreach_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type mockRepo struct {
	count         int
	leads         []outreach.EligibleLead
	activeSteps   []outreach.SequenceStep
	delivery      outreach.SequenceDelivery
	prepared      outreach.RenderedSequenceStep
	preparedHTML  string
	finalizations []outreach.SequenceDeliveryFinalization
	nextDue       *time.Time
}

func (repo *mockRepo) ListEligibleLeads(context.Context, int) ([]outreach.EligibleLead, error) {
	return repo.leads, nil
}

func (repo *mockRepo) CountEligibleLeads(context.Context) (int, error) { return repo.count, nil }

func (repo *mockRepo) ListActiveSequenceSteps(context.Context) ([]outreach.SequenceStep, error) {
	return repo.activeSteps, nil
}

func (repo *mockRepo) GetSequenceDelivery(context.Context, uuid.UUID, int) (outreach.SequenceDelivery, error) {
	return repo.delivery, nil
}

func (repo *mockRepo) PrepareSequenceDelivery(_ context.Context, _ uuid.UUID, step int, subject, bodyHTML, bodyText string) error {
	repo.prepared = outreach.RenderedSequenceStep{Position: step, Subject: subject, BodyText: bodyText}
	repo.preparedHTML = bodyHTML
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
		config.AppURLsConfig{
			PresentationSiteURL: "https://tuvisolutions.com/services/restaurants",
			PublicMarketingURL:  "https://tuvisolutions.com",
		},
		nil,
		nil,
	)
}

func internalAdminPrincipal() auth.Principal {
	return auth.Principal{UserID: uuid.New(), Role: auth.RoleInternalAdmin}
}

func eligibleSequenceRepo() *mockRepo {
	restaurantID := uuid.New()
	campaignID := uuid.New()
	consentAt := time.Now().UTC()
	step := outreach.SequenceStep{
		ID: uuid.New(), SequenceID: uuid.New(), Position: 1, Enabled: true,
		SubjectTemplate:  "A practical idea for {{restaurant_name}}",
		BodyTextTemplate: "{{greeting}}\n\nI had one practical idea for {{restaurant_name}}. Open to a quick note back?",
	}
	return &mockRepo{
		leads:       []outreach.EligibleLead{{CampaignID: campaignID, RestaurantID: restaurantID, Step: 1}},
		activeSteps: []outreach.SequenceStep{step},
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

func TestRunBulkSendFinalizesAcceptedSequenceWithSharedSignature(t *testing.T) {
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
	if !strings.Contains(provider.request.HTMLBody, "tuvi-solutions-logo-transparent.png") ||
		!strings.Contains(provider.request.HTMLBody, "Team Tuvi") {
		t.Fatalf("HTMLBody missing shared logo signature: %q", provider.request.HTMLBody)
	}
	if got := len(strings.FieldsFunc(provider.request.TextBody, func(r rune) bool { return r == '\n' })); got == 0 {
		t.Fatal("TextBody is empty")
	}
	for _, token := range []string{"Thanks & Regards,", "Team Tuvi", "Tuvi Solutions", "https://tuvisolutions.com"} {
		if !strings.Contains(provider.request.TextBody, token) {
			t.Fatalf("TextBody missing signature token %q", token)
		}
	}
	if repo.prepared.BodyText != provider.request.TextBody {
		t.Fatalf("prepared BodyText does not match sent TextBody")
	}
	if repo.preparedHTML != provider.request.HTMLBody {
		t.Fatalf("prepared HTMLBody does not match sent HTMLBody")
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

func TestSendTemplateTestEmailsSendsSavedSequenceExactlyWithSignature(t *testing.T) {
	repo := eligibleSequenceRepo()
	provider := &mockEmailProvider{result: emailprovider.SendResult{ProviderMessageID: "message-1"}}
	service := newSequenceService(t, repo, provider)

	result, err := service.SendTemplateTest(context.Background(), internalAdminPrincipal(), outreach.TemplateTestSendInput{
		RecipientEmail: "test@example.com",
		RestaurantName: "Signature Cafe",
		OwnerFirstName: "Casey",
	})
	if err != nil {
		t.Fatalf("SendTemplateTest() error = %v", err)
	}
	if result.RecipientEmail != "test@example.com" || len(result.Items) != 1 {
		t.Fatalf("result = %#v, want the one enabled sequence item", result)
	}
	if provider.request.To != "test@example.com" {
		t.Fatalf("last request To = %q", provider.request.To)
	}
	if provider.request.Subject != "A practical idea for Signature Cafe" {
		t.Fatalf("Subject = %q, want the rendered saved subject without a test prefix", provider.request.Subject)
	}
	if !strings.Contains(provider.request.HTMLBody, "tuvi-solutions-logo-transparent.png") ||
		!strings.Contains(provider.request.HTMLBody, "Team Tuvi") {
		t.Fatalf("HTMLBody missing shared logo signature: %q", provider.request.HTMLBody)
	}
	if !strings.Contains(provider.request.TextBody, "Team Tuvi") ||
		!strings.Contains(provider.request.TextBody, "https://tuvisolutions.com") {
		t.Fatalf("last request missing text signature: %q", provider.request.TextBody)
	}
}
