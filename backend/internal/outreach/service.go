package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type Service struct {
	repo            Repository
	pool            *pgxpool.Pool
	campaigns       campaigns.Repository
	campaignService *campaigns.Service
	tokenResolver   DemoTokenResolver
	emailPool       *emailprovider.AccountPool
	emailCfg        config.EmailConfig
	outreachCfg     config.OutreachConfig
	enqueuer        BulkJobEnqueuer
	log             *slog.Logger
}

func NewService(
	repo Repository,
	pool *pgxpool.Pool,
	campaignsRepo campaigns.Repository,
	campaignService *campaigns.Service,
	tokenResolver DemoTokenResolver,
	emailPool *emailprovider.AccountPool,
	emailCfg config.EmailConfig,
	outreachCfg config.OutreachConfig,
	enqueuer BulkJobEnqueuer,
	log *slog.Logger,
) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		repo:            repo,
		pool:            pool,
		campaigns:       campaignsRepo,
		campaignService: campaignService,
		tokenResolver:   tokenResolver,
		emailPool:       emailPool,
		emailCfg:        emailCfg,
		outreachCfg:     outreachCfg,
		enqueuer:        enqueuer,
		log:             log,
	}
}

func (service *Service) TriggerBulkSend(ctx context.Context, principal auth.Principal) (TriggerResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return TriggerResult{}, fmt.Errorf("forbidden")
	}
	if service.emailPool == nil || len(service.outreachCfg.ZohoAccounts) == 0 {
		return TriggerResult{}, ErrNotConfigured
	}
	if err := validateBulkMax(service.outreachCfg.BulkMax); err != nil {
		return TriggerResult{}, err
	}

	active, _, err := HasActiveBulkJob(ctx, service.pool, BulkSendJobType)
	if err != nil {
		return TriggerResult{}, err
	}
	if active {
		return TriggerResult{}, ErrBulkJobActive
	}

	pending, err := service.repo.CountEligibleLeads(ctx)
	if err != nil {
		return TriggerResult{}, err
	}

	jobID, err := service.enqueuer.EnqueueBulkSend(ctx, principal.UserID)
	if err != nil {
		return TriggerResult{}, err
	}

	return TriggerResult{
		JobID:                jobID,
		Status:               "queued",
		MaxSends:             service.outreachCfg.BulkMax,
		PendingEligibleCount: pending,
	}, nil
}

func (service *Service) GetStatus(ctx context.Context, principal auth.Principal) (StatusResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return StatusResult{}, fmt.Errorf("forbidden")
	}

	pending, err := service.repo.CountEligibleLeads(ctx)
	if err != nil {
		return StatusResult{}, err
	}

	result := StatusResult{
		PendingEligibleCount: pending,
		MaxSends:             service.outreachCfg.BulkMax,
	}

	active, activeJobID, err := HasActiveBulkJob(ctx, service.pool, BulkSendJobType)
	if err != nil {
		return StatusResult{}, err
	}
	if active {
		result.ActiveJob = &ActiveJobStatus{
			JobID:  activeJobID,
			Status: "running_or_queued",
		}
	}

	lastJob, err := GetLatestBulkJobSummary(ctx, service.pool, BulkSendJobType)
	if err != nil {
		return StatusResult{}, err
	}
	result.LastCompletedJob = lastJob
	return result, nil
}

func (service *Service) RunBulkSend(ctx context.Context, triggeredBy uuid.UUID) (BulkSendSummary, error) {
	summary := BulkSendSummary{MaxSends: service.outreachCfg.BulkMax}

	if service.emailPool == nil || len(service.outreachCfg.ZohoAccounts) == 0 {
		return summary, ErrNotConfigured
	}

	leads, err := service.repo.ListEligibleLeads(ctx, service.outreachCfg.BulkMax)
	if err != nil {
		return summary, err
	}

	if len(leads) == 0 {
		summary.StoppedReason = "no_eligible_leads"
		return summary, nil
	}

	principal := auth.Principal{Role: auth.RoleInternalAdmin, UserID: triggeredBy}

	for _, lead := range leads {
		if service.emailPool.Exhausted() {
			summary.StoppedReason = "account_limit_reached"
			break
		}

		summary.Attempted++

		if err := service.sendLead(ctx, principal, lead); err != nil {
			service.log.WarnContext(ctx, "bulk_outreach_lead_failed",
				"restaurant_id", lead.RestaurantID,
				"email", lead.Email,
				"error", err,
			)
			if err == emailprovider.ErrAccountsExhausted {
				summary.StoppedReason = "account_limit_reached"
				break
			}
			summary.Failed++
			continue
		}
		summary.Sent++
	}

	if summary.StoppedReason == "" && summary.Sent+summary.Failed >= len(leads) && !service.emailPool.Exhausted() {
		summary.StoppedReason = "batch_complete"
	}

	service.log.InfoContext(ctx, "bulk_outreach_completed",
		"attempted", summary.Attempted,
		"sent", summary.Sent,
		"failed", summary.Failed,
		"stopped_reason", summary.StoppedReason,
	)

	return summary, nil
}

func (service *Service) sendLead(ctx context.Context, principal auth.Principal, lead EligibleLead) error {
	demoToken, err := service.tokenResolver.Resolve(ctx, lead.DemoSiteID)
	if err != nil {
		return err
	}

	campaign, err := service.campaignService.CreateDraft(ctx, principal, campaigns.CreateInput{
		RestaurantID: lead.RestaurantID,
		DemoSiteID:   lead.DemoSiteID,
		DemoToken:    demoToken,
		CampaignType: campaigns.TypeOutreach,
	})
	if err != nil {
		return fmt.Errorf("create campaign draft: %w", err)
	}

	campaign, err = service.campaignService.Approve(ctx, principal, campaign.ID)
	if err != nil {
		return fmt.Errorf("approve campaign: %w", err)
	}

	sendCtx, err := service.campaigns.GetSendContext(ctx, campaign.ID)
	if err != nil {
		return fmt.Errorf("load send context: %w", err)
	}

	suppressed, err := service.campaigns.IsSuppressed(ctx, sendCtx.RestaurantEmail)
	if err != nil {
		return fmt.Errorf("check suppression: %w", err)
	}
	if err := campaigns.CheckBulkEligibility(campaigns.BulkEligibilityInput{
		RestaurantEmail: sendCtx.RestaurantEmail,
		DemoStatus:      sendCtx.DemoStatus,
		Suppressed:      suppressed,
	}); err != nil {
		return err
	}

	if _, err := service.campaigns.MarkSending(ctx, campaign.ID, 0); err != nil {
		return fmt.Errorf("mark campaign sending: %w", err)
	}

	trackingURLs, err := service.campaignService.BuildTrackingURLs(ctx, campaign, sendCtx)
	if err != nil {
		return fmt.Errorf("build tracking urls: %w", err)
	}

	draft := campaigns.InjectTracking(campaigns.DraftContent{
		Subject:  campaign.Subject,
		BodyHTML: campaign.BodyHTML,
		BodyText: campaign.BodyText,
	}, trackingURLs, service.emailCfg.OpenTrackingEnabled)

	result, err := service.emailPool.Send(ctx, emailprovider.SendRequest{
		To:       sendCtx.RestaurantEmail,
		Subject:  draft.Subject,
		HTMLBody: draft.BodyHTML,
		TextBody: draft.BodyText,
		Metadata: map[string]string{
			"campaign_id":   campaign.ID.String(),
			"restaurant_id": campaign.RestaurantID.String(),
			"original_to":   sendCtx.RestaurantEmail,
			"bulk_outreach": "true",
		},
	})
	if err != nil {
		meta, _ := json.Marshal(map[string]string{"error": err.Error(), "bulk_outreach": "true"})
		_ = service.campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventFailed, meta)
		return err
	}

	eventMeta, _ := json.Marshal(map[string]any{
		"step":               0,
		"provider_message":   result.ProviderMessageID,
		"redirected_to":      result.RedirectedTo,
		"skipped":            result.Skipped,
		"original_recipient": sendCtx.RestaurantEmail,
		"bulk_outreach":      true,
		"account_index":      service.emailPool.CurrentAccountIndex(),
	})
	if err := service.campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventSent, eventMeta); err != nil {
		return err
	}
	if _, err := service.campaigns.MarkSent(ctx, campaign.ID, 0); err != nil {
		return err
	}
	if err := service.campaigns.MarkRestaurantEmailed(ctx, campaign.RestaurantID); err != nil {
		return err
	}

	service.log.InfoContext(ctx, "bulk_outreach_lead_sent",
		"restaurant_id", lead.RestaurantID,
		"to", strings.TrimSpace(sendCtx.RestaurantEmail),
		"campaign_id", campaign.ID,
	)
	return nil
}

func (service *Service) UpdateJobSummary(ctx context.Context, jobID string, triggeredBy uuid.UUID, summary BulkSendSummary) error {
	payload, err := encodeBulkSummary(summary, triggeredBy.String())
	if err != nil {
		return err
	}
	const query = `
		UPDATE job_runs
		SET payload = $2, updated_at = now()
		WHERE id = $1::uuid`
	_, err = service.pool.Exec(ctx, query, jobID, payload)
	if err != nil {
		return fmt.Errorf("update bulk job summary: %w", err)
	}
	return nil
}
