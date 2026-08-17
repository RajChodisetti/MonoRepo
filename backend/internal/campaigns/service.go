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
	repo                Repository
	demos               demos.Repository
	access              *restaurants.Service
	enqueuer            SendJobEnqueuer
	publicBase          string
	publicWebURL        string
	presentationSiteURL string
	marketingSiteURL    string
	demoTokenTTL        time.Duration
}

const preferredPreviewTemplateID = "3"

func NewService(
	repo Repository,
	demosRepo demos.Repository,
	access *restaurants.Service,
	enqueuer SendJobEnqueuer,
	urls config.AppURLsConfig,
	demoTokenTTL ...time.Duration,
) *Service {
	tokenTTL := 30 * 24 * time.Hour
	if len(demoTokenTTL) > 0 && demoTokenTTL[0] > 0 {
		tokenTTL = demoTokenTTL[0]
	}
	return &Service{
		repo:                repo,
		demos:               demosRepo,
		access:              access,
		enqueuer:            enqueuer,
		publicBase:          strings.TrimRight(urls.PublicBaseURL, "/"),
		publicWebURL:        strings.TrimRight(urls.PublicWebURL, "/"),
		presentationSiteURL: strings.TrimRight(urls.PresentationSiteURL, "/"),
		marketingSiteURL:    strings.TrimRight(urls.PublicMarketingURL, "/"),
		demoTokenTTL:        tokenTTL,
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
	demoToken := strings.TrimSpace(input.DemoToken)
	if demoToken == "" || demos.CheckDemoToken(demoSite.TokenHash, demoToken) != nil {
		return Campaign{}, fmt.Errorf("%w: demo token does not match the selected demo", ErrNotEligible)
	}
	if demoSite.ExpiresAt != nil && !demoSite.ExpiresAt.After(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("%w: demo link has expired and must be regenerated", ErrNotEligible)
	}

	profileCtx, err := service.repo.GetRestaurantContext(ctx, input.RestaurantID)
	if err != nil {
		return Campaign{}, err
	}

	campaignType := strings.ToLower(strings.TrimSpace(input.CampaignType))
	if campaignType == "" {
		campaignType = TypeOutreach
	}
	if campaignType != TypeOutreach {
		return Campaign{}, ErrUnsupportedType
	}

	draft := BuildDraft(DraftInput{
		RestaurantName:      profileCtx.RestaurantName,
		DemoWebURL:          service.publicWebURL,
		DemoSlug:            demoSite.Slug,
		DemoToken:           demoToken,
		PresentationSiteURL: service.presentationSiteURL,
		MarketingSiteURL:    service.marketingSiteURL,
	})

	return service.repo.Create(ctx, CreateInput{
		RestaurantID: input.RestaurantID,
		DemoSiteID:   input.DemoSiteID,
		DemoToken:    demoToken,
		CampaignType: campaignType,
	}, draft)
}

type DemoSiteLink struct {
	DemoSiteID uuid.UUID  `json:"demo_site_id"`
	Slug       string     `json:"slug"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	PreviewURL string     `json:"preview_url,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ListDemoLinks surfaces every demo site created for a restaurant, with a
// live preview URL when a campaign has already been drafted against that
// demo (which is what carries the plaintext token needed to build the link).
func (service *Service) ListDemoLinks(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) ([]DemoSiteLink, error) {
	if err := service.access.CanAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return nil, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, restaurants.ErrForbidden
	}

	sites, err := service.demos.ListByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	links := make([]DemoSiteLink, 0, len(sites))
	for _, site := range sites {
		link := DemoSiteLink{
			DemoSiteID: site.ID,
			Slug:       site.Slug,
			Status:     site.Status,
			ExpiresAt:  site.ExpiresAt,
			CreatedAt:  site.CreatedAt,
			UpdatedAt:  site.UpdatedAt,
		}
		if token, tokenErr := service.repo.GetLatestDemoTokenByDemoSiteID(ctx, site.ID); tokenErr == nil && token != "" {
			link.PreviewURL = buildTokenGatedDemoPreviewURL(service.publicWebURL, site.Slug, token, preferredPreviewTemplateID)
		}
		links = append(links, link)
	}
	return links, nil
}

func (service *Service) ListByRestaurant(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) ([]Campaign, error) {
	if err := service.access.CanAccessRestaurant(ctx, principal, restaurantID); err != nil {
		return nil, err
	}
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, restaurants.ErrForbidden
	}
	records, err := service.repo.ListByRestaurant(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	restaurantName := ""
	if profileCtx, ctxErr := service.repo.GetRestaurantContext(ctx, restaurantID); ctxErr == nil {
		restaurantName = profileCtx.RestaurantName
	}
	for index := range records {
		records[index] = service.withCurrentOutreachDraft(records[index], restaurantName)
	}
	return records, nil
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
	campaign = service.withCurrentOutreachDraft(campaign, service.restaurantNameForCampaign(ctx, campaign))
	events, err := service.repo.ListEvents(ctx, campaignID)
	if err != nil {
		return Campaign{}, nil, err
	}
	return campaign, events, nil
}

func (service *Service) BuildCurrentOutreachDraft(restaurantName string) (DraftContent, error) {
	return RenderOutreachEmailWithLinks(restaurantName, OutreachLinkConfig{
		PresentationURL: service.presentationSiteURL,
		MarketingURL:    service.marketingSiteURL,
	})
}

func (service *Service) withCurrentOutreachDraft(campaign Campaign, restaurantName string) Campaign {
	if campaign.CampaignType != TypeOutreach {
		return campaign
	}
	name := strings.TrimSpace(restaurantName)
	if name == "" {
		name = nameFromOutreachSubject(campaign.Subject)
	}
	draft, err := service.BuildCurrentOutreachDraft(name)
	if err != nil {
		return campaign
	}
	campaign.Subject = draft.Subject
	campaign.BodyHTML = draft.BodyHTML
	campaign.BodyText = draft.BodyText
	return campaign
}

func (service *Service) restaurantNameForCampaign(ctx context.Context, campaign Campaign) string {
	if campaign.RestaurantID == uuid.Nil {
		return ""
	}
	profileCtx, err := service.repo.GetRestaurantContext(ctx, campaign.RestaurantID)
	if err != nil {
		return ""
	}
	return profileCtx.RestaurantName
}

func nameFromOutreachSubject(subject string) string {
	value := strings.TrimSpace(subject)
	const prefix = "A live demo for "
	const suffix = " — AI receptionist, website & more"
	if strings.HasPrefix(value, prefix) && strings.HasSuffix(value, suffix) {
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, prefix), suffix))
	}
	return ""
}

func (service *Service) Approve(
	ctx context.Context,
	principal auth.Principal,
	campaignID uuid.UUID,
	expectedUpdatedAt time.Time,
) (Campaign, error) {
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
	if expectedUpdatedAt.IsZero() || !campaign.UpdatedAt.Equal(expectedUpdatedAt) {
		return Campaign{}, ErrStaleReview
	}
	if campaign.Status == StatusStopped {
		return Campaign{}, ErrAlreadyStopped
	}
	alreadyAudited := campaign.Status == StatusApproved && campaign.ApprovedAt != nil && campaign.ApprovedBy != nil
	if campaign.Status != StatusDraft && campaign.Status != StatusApproved {
		return Campaign{}, fmt.Errorf("%w: only a draft campaign can be approved", ErrNotEligible)
	}
	demoSite, err := service.demos.GetByID(ctx, campaign.DemoSiteID)
	if err != nil {
		return Campaign{}, err
	}
	if demoSite.RestaurantID != campaign.RestaurantID {
		return Campaign{}, repository.ErrNotFound
	}
	if demoSite.ExpiresAt != nil && !demoSite.ExpiresAt.After(time.Now().UTC()) {
		return Campaign{}, fmt.Errorf("%w: demo link has expired and must be regenerated", ErrNotEligible)
	}
	if strings.TrimSpace(campaign.DemoToken) == "" || demos.CheckDemoToken(demoSite.TokenHash, campaign.DemoToken) != nil {
		return Campaign{}, fmt.Errorf("%w: campaign demo token is no longer valid", ErrNotEligible)
	}
	if alreadyAudited {
		return campaign, nil
	}
	return service.repo.Approve(ctx, campaignID, principal.UserID, expectedUpdatedAt)
}

// RegenerateDraft rotates the opaque demo token and rebuilds the current email
// content. It is deliberately limited to a draft demo and a draft/approved
// campaign, and always clears campaign approval so the changed artifact must be
// reviewed again before outreach.
func (service *Service) RegenerateDraft(ctx context.Context, principal auth.Principal, campaignID uuid.UUID) (Campaign, error) {
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
	if campaign.CampaignType != TypeOutreach {
		return Campaign{}, ErrUnsupportedType
	}
	if campaign.Status != StatusDraft && campaign.Status != StatusApproved {
		return Campaign{}, fmt.Errorf("%w: only a draft or approved campaign can be regenerated", ErrNotEligible)
	}

	demoSite, err := service.demos.GetByID(ctx, campaign.DemoSiteID)
	if err != nil {
		return Campaign{}, err
	}
	if demoSite.RestaurantID != campaign.RestaurantID {
		return Campaign{}, repository.ErrNotFound
	}
	if demoSite.Status != demos.StatusDraft {
		return Campaign{}, fmt.Errorf("%w: unpublish the demo before regenerating its token", ErrNotEligible)
	}

	profileCtx, err := service.repo.GetRestaurantContext(ctx, campaign.RestaurantID)
	if err != nil {
		return Campaign{}, err
	}
	token, err := demos.GenerateDemoToken()
	if err != nil {
		return Campaign{}, err
	}
	tokenHash, err := demos.HashDemoToken(token)
	if err != nil {
		return Campaign{}, err
	}
	expiresAt := time.Now().UTC().Add(service.demoTokenTTL)
	draft := BuildDraft(DraftInput{
		RestaurantName:      profileCtx.RestaurantName,
		DemoWebURL:          service.publicWebURL,
		DemoSlug:            demoSite.Slug,
		DemoToken:           token,
		PresentationSiteURL: service.presentationSiteURL,
		MarketingSiteURL:    service.marketingSiteURL,
	})
	return service.repo.RegenerateDraft(
		ctx,
		campaign.ID,
		draft,
		token,
		tokenHash,
		&expiresAt,
		principal.UserID,
	)
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
	if campaign.CampaignType != TypeOutreach {
		return Campaign{}, ErrUnsupportedType
	}
	return Campaign{}, ErrOutreachRequiresBulk
}

func (service *Service) BuildTrackingURLs(ctx context.Context, campaign Campaign, sendCtx SendContext) (TrackingURLs, error) {
	demoSite, token, err := service.validPublishedDemoTarget(ctx, campaign)
	if err != nil {
		return TrackingURLs{}, err
	}
	if demoSite.Status != demos.StatusPublished || sendCtx.DemoStatus != demos.StatusPublished {
		return TrackingURLs{}, fmt.Errorf("%w: demo site is not published", ErrNotEligible)
	}

	template1Target := buildTokenGatedDemoPreviewURL(service.publicWebURL, demoSite.Slug, token, "1")

	clickToken, err := newTrackingToken()
	if err != nil {
		return TrackingURLs{}, err
	}
	openToken, err := newTrackingToken()
	if err != nil {
		return TrackingURLs{}, err
	}

	demoSiteID := campaign.DemoSiteID
	recipientEmail := strings.ToLower(strings.TrimSpace(sendCtx.RestaurantEmail))
	if recipientEmail == "" {
		return TrackingURLs{}, fmt.Errorf("%w: outreach recipient is empty", ErrNotEligible)
	}
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)
	if demoSite.ExpiresAt != nil && demoSite.ExpiresAt.Before(expires) {
		expires = *demoSite.ExpiresAt
	}

	tokens := []TrackingToken{
		{Token: clickToken, CampaignID: campaign.ID, RestaurantID: campaign.RestaurantID, DemoSiteID: &demoSiteID, TokenType: TokenClick, TargetURL: template1Target, RecipientEmail: recipientEmail, ExpiresAt: &expires},
		{Token: openToken, CampaignID: campaign.ID, RestaurantID: campaign.RestaurantID, DemoSiteID: &demoSiteID, TokenType: TokenOpen, TargetURL: "", RecipientEmail: recipientEmail, ExpiresAt: &expires},
	}

	base := service.publicBase
	if base == "" {
		base = "http://localhost:8080"
	}
	urls := TrackingURLs{
		Click:     base + "/t/click/" + clickToken,
		Template1: base + "/t/click/" + clickToken,
		Open:      base + "/t/open/" + openToken,
	}

	for _, tokenRecord := range tokens {
		if err := service.repo.CreateTrackingToken(ctx, tokenRecord); err != nil {
			return TrackingURLs{}, err
		}
	}

	return urls, nil
}

func (service *Service) validPublishedDemoTarget(ctx context.Context, campaign Campaign) (demos.Site, string, error) {
	if service.demos == nil {
		return demos.Site{}, "", fmt.Errorf("%w: demo repository is not configured", ErrNotEligible)
	}
	demoSite, err := service.demos.GetByID(ctx, campaign.DemoSiteID)
	if err != nil {
		return demos.Site{}, "", err
	}
	if demoSite.RestaurantID != campaign.RestaurantID {
		return demos.Site{}, "", fmt.Errorf("%w: campaign demo does not belong to the restaurant", ErrNotEligible)
	}
	if demoSite.Status != demos.StatusPublished {
		return demos.Site{}, "", fmt.Errorf("%w: demo site is not published", ErrNotEligible)
	}
	if demoSite.ExpiresAt != nil && !demoSite.ExpiresAt.After(time.Now().UTC()) {
		return demos.Site{}, "", fmt.Errorf("%w: demo link has expired and must be regenerated and republished", ErrNotEligible)
	}
	token := strings.TrimSpace(campaign.DemoToken)
	if token == "" || demos.CheckDemoToken(demoSite.TokenHash, token) != nil {
		return demos.Site{}, "", fmt.Errorf("%w: campaign demo token is no longer valid", ErrNotEligible)
	}
	return demoSite, token, nil
}

func newTrackingToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate tracking token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
