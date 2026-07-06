package campaigns

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type Service struct {
	repo         Repository
	demos        demos.Repository
	access       *restaurants.Service
	enqueuer     SendJobEnqueuer
	publicBase   string
	publicWebURL string
}

func NewService(
	repo Repository,
	demosRepo demos.Repository,
	access *restaurants.Service,
	enqueuer SendJobEnqueuer,
	urls config.AppURLsConfig,
) *Service {
	return &Service{
		repo:         repo,
		demos:        demosRepo,
		access:       access,
		enqueuer:     enqueuer,
		publicBase:   strings.TrimRight(urls.PublicBaseURL, "/"),
		publicWebURL: strings.TrimRight(urls.PublicWebURL, "/"),
	}
}

func (service *Service) CreateDraft(ctx context.Context, principal auth.Principal, input CreateInput) (Campaign, error) {
	if err := service.access.CanAccessRestaurant(ctx, principal, input.RestaurantID); err != nil {
		return Campaign{}, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return Campaign{}, restaurants.ErrForbidden
	}

	demoSite, err := service.demos.GetByID(ctx, input.DemoSiteID)
	if err != nil {
		return Campaign{}, err
	}
	if demoSite.RestaurantID != input.RestaurantID {
		return Campaign{}, repository.ErrNotFound
	}

	profileCtx, err := service.repo.GetRestaurantContext(ctx, input.RestaurantID)
	if err != nil {
		return Campaign{}, err
	}

	campaignType := strings.TrimSpace(input.CampaignType)
	if campaignType == "" {
		campaignType = TypeOutreach
	}

	draft := BuildDraft(DraftInput{
		RestaurantName: profileCtx.RestaurantName,
		DemoWebURL:     service.publicWebURL,
		DemoSlug:       demoSite.Slug,
		DemoToken:      strings.TrimSpace(input.DemoToken),
	})

	return service.repo.Create(ctx, CreateInput{
		RestaurantID: input.RestaurantID,
		DemoSiteID:   input.DemoSiteID,
		DemoToken:    strings.TrimSpace(input.DemoToken),
		CampaignType: campaignType,
	}, draft)
}

func (service *Service) ListByRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) ([]Campaign, error) {
	if err := service.access.CanAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return nil, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, restaurants.ErrForbidden
	}
	return service.repo.ListByRestaurant(ctx, restaurantID)
}

func (service *Service) GetByID(ctx context.Context, principal auth.Principal, campaignID uuid.UUID) (Campaign, []Event, error) {
	campaign, err := service.repo.GetByID(ctx, campaignID)
	if err != nil {
		return Campaign{}, nil, err
	}
	if err := service.access.CanAccessRestaurant(ctx, principal, campaign.RestaurantID); err != nil {
		return Campaign{}, nil, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return Campaign{}, nil, restaurants.ErrForbidden
	}
	events, err := service.repo.ListEvents(ctx, campaignID)
	if err != nil {
		return Campaign{}, nil, err
	}
	return campaign, events, nil
}

func (service *Service) Approve(ctx context.Context, principal auth.Principal, campaignID uuid.UUID) (Campaign, error) {
	campaign, err := service.repo.GetByID(ctx, campaignID)
	if err != nil {
		return Campaign{}, err
	}
	if err := service.access.CanAccessRestaurant(ctx, principal, campaign.RestaurantID); err != nil {
		return Campaign{}, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return Campaign{}, restaurants.ErrForbidden
	}
	if campaign.Status == StatusStopped {
		return Campaign{}, ErrAlreadyStopped
	}
	return service.repo.Approve(ctx, campaignID, principal.UserID)
}

func (service *Service) Stop(ctx context.Context, principal auth.Principal, campaignID uuid.UUID, reason string) (Campaign, error) {
	campaign, err := service.repo.GetByID(ctx, campaignID)
	if err != nil {
		return Campaign{}, err
	}
	if err := service.access.CanAccessRestaurant(ctx, principal, campaign.RestaurantID); err != nil {
		return Campaign{}, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return Campaign{}, restaurants.ErrForbidden
	}
	if strings.TrimSpace(reason) == "" {
		reason = "stopped by admin"
	}
	return service.repo.Stop(ctx, campaignID, reason)
}

func (service *Service) SendStep(ctx context.Context, principal auth.Principal, campaignID uuid.UUID, step int) (Campaign, error) {
	campaign, err := service.repo.GetByID(ctx, campaignID)
	if err != nil {
		return Campaign{}, err
	}
	if err := service.access.CanAccessRestaurant(ctx, principal, campaign.RestaurantID); err != nil {
		return Campaign{}, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return Campaign{}, restaurants.ErrForbidden
	}
	if campaign.Status == StatusStopped {
		return Campaign{}, ErrAlreadyStopped
	}

	sendCtx, err := service.repo.GetSendContext(ctx, campaignID)
	if err != nil {
		return Campaign{}, err
	}
	suppressed, err := service.repo.IsSuppressed(ctx, sendCtx.RestaurantEmail)
	if err != nil {
		return Campaign{}, err
	}
	if err := CheckEligibility(EligibilityInput{
		RestaurantEmail: sendCtx.RestaurantEmail,
		ReviewStatus:    sendCtx.ReviewStatus,
		DemoStatus:      sendCtx.DemoStatus,
		CampaignStatus:  campaign.Status,
		Suppressed:      suppressed,
	}); err != nil {
		return Campaign{}, err
	}

	updated, err := service.repo.MarkSending(ctx, campaignID, step)
	if err != nil {
		return Campaign{}, err
	}

	if err := service.enqueuer.EnqueueSendStep(ctx, campaignID, step); err != nil {
		return Campaign{}, fmt.Errorf("enqueue email send job: %w", err)
	}

	return updated, nil
}

func (service *Service) BuildTrackingURLs(ctx context.Context, campaign Campaign, sendCtx SendContext) (clickURL, openURL, unsubscribeURL, demoTarget string, err error) {
	clickToken, err := newTrackingToken()
	if err != nil {
		return "", "", "", "", err
	}
	openToken, err := newTrackingToken()
	if err != nil {
		return "", "", "", "", err
	}
	unsubToken, err := newTrackingToken()
	if err != nil {
		return "", "", "", "", err
	}

	demoTarget = buildDemoURL(service.publicWebURL, sendCtx.DemoSlug, campaign.DemoToken)
	demoSiteID := campaign.DemoSiteID
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)

	tokens := []TrackingToken{
		{Token: clickToken, CampaignID: campaign.ID, RestaurantID: campaign.RestaurantID, DemoSiteID: &demoSiteID, TokenType: TokenClick, TargetURL: demoTarget, ExpiresAt: &expires},
		{Token: openToken, CampaignID: campaign.ID, RestaurantID: campaign.RestaurantID, DemoSiteID: &demoSiteID, TokenType: TokenOpen, TargetURL: "", ExpiresAt: &expires},
		{Token: unsubToken, CampaignID: campaign.ID, RestaurantID: campaign.RestaurantID, DemoSiteID: &demoSiteID, TokenType: TokenUnsubscribe, TargetURL: "", ExpiresAt: &expires},
	}
	for _, token := range tokens {
		if err := service.repo.CreateTrackingToken(ctx, token); err != nil {
			return "", "", "", "", err
		}
	}

	base := service.publicBase
	if base == "" {
		base = "http://localhost:8080"
	}
	return base + "/t/click/" + clickToken,
		base + "/t/open/" + openToken,
		base + "/t/unsubscribe/" + unsubToken,
		demoTarget,
		nil
}

func newTrackingToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate tracking token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
