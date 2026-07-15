package outreach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
	if !service.emailSendingEnabled() {
		return TriggerResult{}, ErrSendingDisabled
	}
	if service.emailPool == nil {
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
	if service.emailPool != nil && service.emailPool.Durable() {
		nextAvailableAt, err := service.emailPool.NextAvailableAt(ctx)
		if err != nil {
			return StatusResult{}, err
		}
		result.NextAvailableAt = nextAvailableAt
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

func (service *Service) RunBulkSend(ctx context.Context, triggeredBy uuid.UUID, jobIDs ...string) (BulkSendSummary, error) {
	summary := BulkSendSummary{MaxSends: service.outreachCfg.BulkMax}
	var bulkJobID *uuid.UUID
	if len(jobIDs) > 0 && jobIDs[0] != "" {
		if parsed, err := uuid.Parse(jobIDs[0]); err == nil {
			bulkJobID = &parsed
			persisted, loadErr := service.loadBulkJobSummary(ctx, parsed)
			if loadErr != nil {
				return summary, loadErr
			}
			summary = persisted
			summary.MaxSends = service.outreachCfg.BulkMax
			summary.StoppedReason = ""
			summary.NextAvailableAt = nil
		}
	}
	startedAttempted := summary.Attempted

	if !service.emailSendingEnabled() {
		return summary, ErrSendingDisabled
	}
	if service.emailPool == nil {
		return summary, ErrNotConfigured
	}
	if err := validateBulkMax(service.outreachCfg.BulkMax); err != nil {
		return summary, err
	}
	leadLimit := service.outreachCfg.BulkMax
	if service.emailPool.Durable() {
		// A durable execution crosses the provider boundary at most once. The
		// same job is then released until PostgreSQL says the next slot is due.
		leadLimit = 1
	}
	leads, err := service.repo.ListEligibleLeads(ctx, leadLimit)
	if err != nil {
		return summary, err
	}

	if len(leads) == 0 {
		summary.StoppedReason = "no_eligible_leads"
		return summary, nil
	}

	for _, lead := range leads {
		if !service.emailPool.Durable() && service.emailPool.Exhausted() {
			summary.StoppedReason = "account_limit_reached"
			break
		}

		sent, err := service.sendLead(ctx, lead, bulkJobID)
		if err != nil {
			if errors.Is(err, emailprovider.ErrAccountsExhausted) {
				summary.StoppedReason = "paced"
				nextAvailableAt, availabilityErr := service.emailPool.NextAvailableAt(ctx)
				if availabilityErr != nil {
					return summary, availabilityErr
				}
				if nextAvailableAt == nil {
					// An account can cross its availability boundary between the
					// failed claim and this lookup. Requeue briefly instead of
					// failing a one-attempt bulk job at that boundary.
					retryAt := time.Now().UTC().Add(time.Second)
					nextAvailableAt = &retryAt
				}
				summary.NextAvailableAt = nextAvailableAt
				break
			}
			summary.Attempted++
			if errors.Is(err, campaigns.ErrNotEligible) {
				summary.Skipped++
				service.log.InfoContext(ctx, "bulk_outreach_lead_became_ineligible",
					"restaurant_id", lead.RestaurantID,
					"error", err,
				)
				continue
			}
			service.log.WarnContext(ctx, "bulk_outreach_lead_failed",
				"restaurant_id", lead.RestaurantID,
				"error", err,
			)
			summary.Failed++
			summary.StoppedReason = "delivery_error"
			return summary, err
		} else if sent {
			summary.Attempted++
			summary.Sent++
		} else {
			summary.Attempted++
			summary.Skipped++
		}
	}

	batchAttempted := summary.Attempted - startedAttempted
	if summary.StoppedReason == "" {
		if batchAttempted >= len(leads) {
			summary.StoppedReason = "batch_complete"
		} else if !service.emailPool.Durable() && service.emailPool.Exhausted() {
			summary.StoppedReason = "account_limit_reached"
		}
	}

	// The same human-triggered durable job continues one provider reservation at
	// a time while approved leads remain. PostgreSQL's account and global gates
	// supply the restart-safe next timestamp; no worker sleeps between sends.
	if summary.NextAvailableAt == nil && batchAttempted > 0 {
		remaining, countErr := service.repo.CountEligibleLeads(ctx)
		if countErr != nil {
			return summary, countErr
		}
		if remaining > 0 {
			nextAvailableAt, availabilityErr := service.emailPool.NextAvailableAt(ctx)
			if availabilityErr != nil {
				return summary, availabilityErr
			}
			if nextAvailableAt == nil {
				// The persisted gates can already be due after an eligibility
				// rejection or a slow provider call. Yield briefly; the next claim
				// still rechecks both database gates atomically.
				resumeAt := time.Now().UTC().Add(time.Second)
				nextAvailableAt = &resumeAt
				summary.StoppedReason = "more_eligible_leads"
			} else {
				summary.StoppedReason = "paced"
			}
			summary.NextAvailableAt = nextAvailableAt
		}
	}

	service.log.InfoContext(ctx, "bulk_outreach_completed",
		"attempted", summary.Attempted,
		"sent", summary.Sent,
		"skipped", summary.Skipped,
		"failed", summary.Failed,
		"triggered_by", triggeredBy,
		"stopped_reason", summary.StoppedReason,
	)

	return summary, nil
}

func (service *Service) loadBulkJobSummary(ctx context.Context, jobID uuid.UUID) (BulkSendSummary, error) {
	if service.pool == nil {
		return BulkSendSummary{}, fmt.Errorf("database pool is not configured")
	}
	var payload []byte
	if err := service.pool.QueryRow(ctx, `SELECT payload FROM job_runs WHERE id = $1`, jobID).Scan(&payload); err != nil {
		return BulkSendSummary{}, fmt.Errorf("load bulk outreach job summary: %w", err)
	}
	summary, err := decodeBulkSummary(payload)
	if err != nil {
		return BulkSendSummary{}, err
	}
	return summary, nil
}

func (service *Service) sendLead(ctx context.Context, lead EligibleLead, bulkJobID *uuid.UUID) (bool, error) {
	campaign, err := service.campaigns.GetByID(ctx, lead.CampaignID)
	if err != nil {
		return false, fmt.Errorf("load approved campaign: %w", err)
	}
	if campaign.RestaurantID != lead.RestaurantID || campaign.DemoSiteID != lead.DemoSiteID {
		return false, fmt.Errorf("approved campaign does not match the eligible lead")
	}
	if campaign.Status != campaigns.StatusApproved {
		return false, fmt.Errorf("%w: campaign must be approved before bulk sending", campaigns.ErrNotEligible)
	}

	sendCtx, err := service.campaigns.GetSendContext(ctx, campaign.ID)
	if err != nil {
		return false, fmt.Errorf("load send context: %w", err)
	}

	suppressed, err := service.campaigns.IsSuppressed(ctx, sendCtx.RestaurantEmail)
	if err != nil {
		return false, fmt.Errorf("check suppression: %w", err)
	}
	if err := campaigns.CheckEligibility(campaigns.EligibilityInput{
		RestaurantEmail:         sendCtx.RestaurantEmail,
		OCRStatus:               sendCtx.OCRStatus,
		ReviewStatus:            sendCtx.ReviewStatus,
		ProfileReviewAudited:    sendCtx.ProfileReviewAudited,
		DemoStatus:              sendCtx.DemoStatus,
		DemoPublishAudited:      sendCtx.DemoPublishAudited,
		DemoExpired:             sendCtx.DemoExpired,
		CampaignStatus:          campaign.Status,
		CampaignApprovalAudited: campaign.ApprovedAt != nil && campaign.ApprovedBy != nil,
		Suppressed:              suppressed,
	}); err != nil {
		return false, err
	}

	trackingURLs, err := service.campaignService.BuildTrackingURLs(ctx, campaign, sendCtx)
	if err != nil {
		return false, fmt.Errorf("build tracking urls: %w", err)
	}

	draft := campaigns.InjectTracking(campaigns.DraftContent{
		Subject:  campaign.Subject,
		BodyHTML: campaign.BodyHTML,
		BodyText: campaign.BodyText,
	}, trackingURLs, service.emailCfg.OpenTrackingEnabled)
	if err := campaigns.ValidateRenderedEmail(draft, service.emailCfg.RequireHTTPSLinks, service.emailCfg.AllowedLinkHosts...); err != nil {
		return false, fmt.Errorf("validate rendered outreach email: %w", err)
	}

	if !service.emailPool.Durable() {
		if _, err := service.campaigns.MarkSending(ctx, campaign.ID, 0); err != nil {
			return false, fmt.Errorf("mark campaign sending: %w", err)
		}
	}

	result, err := service.emailPool.Send(ctx, emailprovider.SendRequest{
		To:       sendCtx.RestaurantEmail,
		Subject:  draft.Subject,
		HTMLBody: draft.BodyHTML,
		TextBody: draft.BodyText,
		Metadata: map[string]string{
			"campaign_id":   campaign.ID.String(),
			"restaurant_id": campaign.RestaurantID.String(),
			"bulk_outreach": "true",
		},
		Delivery: &emailprovider.DeliveryContext{
			CampaignID:   campaign.ID,
			RestaurantID: campaign.RestaurantID,
			BulkJobID:    bulkJobID,
			Step:         0,
			CampaignArtifactFingerprint: emailprovider.CampaignArtifactFingerprint(
				campaign.Subject,
				campaign.BodyHTML,
				campaign.BodyText,
				campaign.DemoToken,
			),
		},
	})
	if err != nil {
		if !result.QuotaManaged {
			meta, _ := json.Marshal(map[string]string{"error": err.Error(), "bulk_outreach": "true"})
			_ = service.campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventFailed, meta)
		}
		return false, err
	}

	if result.Skipped || result.RedirectedTo != "" {
		if result.QuotaManaged {
			if !result.Finalized {
				return false, fmt.Errorf("quota-managed skipped delivery was not finalized")
			}
			service.log.InfoContext(ctx, "bulk_outreach_lead_skipped",
				"restaurant_id", lead.RestaurantID,
				"campaign_id", campaign.ID,
				"redirected", result.RedirectedTo != "",
				"account_key", result.AccountKey,
			)
			return false, nil
		}
		eventMeta, _ := json.Marshal(map[string]any{
			"step":          0,
			"skipped":       result.Skipped,
			"redirected":    result.RedirectedTo != "",
			"bulk_outreach": true,
			"account_index": service.emailPool.CurrentAccountIndex(),
		})
		if err := service.campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventSkipped, eventMeta); err != nil {
			return false, err
		}
		if _, err := service.campaigns.MarkSendSkipped(ctx, campaign.ID, 0); err != nil {
			return false, err
		}
		service.log.InfoContext(ctx, "bulk_outreach_lead_skipped",
			"restaurant_id", lead.RestaurantID,
			"campaign_id", campaign.ID,
			"redirected", result.RedirectedTo != "",
		)
		return false, nil
	}
	if result.QuotaManaged {
		if !result.Finalized {
			return false, fmt.Errorf("quota-managed accepted delivery was not finalized")
		}
		service.log.InfoContext(ctx, "bulk_outreach_lead_sent",
			"restaurant_id", lead.RestaurantID,
			"campaign_id", campaign.ID,
			"account_key", result.AccountKey,
			"account_sequence", result.AccountSequence,
			"send_sequence", result.SendSequence,
		)
		return true, nil
	}

	eventMeta, _ := json.Marshal(map[string]any{
		"step":             0,
		"provider_message": result.ProviderMessageID,
		"bulk_outreach":    true,
		"account_index":    service.emailPool.CurrentAccountIndex(),
	})
	if err := service.campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventSent, eventMeta); err != nil {
		return false, err
	}
	if _, err := service.campaigns.MarkSent(ctx, campaign.ID, 0); err != nil {
		return false, err
	}
	if err := service.campaigns.MarkRestaurantEmailed(ctx, campaign.RestaurantID); err != nil {
		return false, err
	}

	service.log.InfoContext(ctx, "bulk_outreach_lead_sent",
		"restaurant_id", lead.RestaurantID,
		"campaign_id", campaign.ID,
	)
	return true, nil
}

func (service *Service) emailSendingEnabled() bool {
	return !service.emailCfg.DisableSending
}

func (service *Service) UpdateJobSummary(
	ctx context.Context,
	jobID string,
	lockedBy string,
	triggeredBy uuid.UUID,
	summary BulkSendSummary,
) error {
	payload, err := encodeBulkSummary(summary, triggeredBy.String())
	if err != nil {
		return err
	}
	const query = `
		UPDATE job_runs
		SET payload = $2, updated_at = now()
		WHERE id = $1::uuid
		  AND status = 'running'
		  AND locked_by = $3`
	result, err := service.pool.Exec(ctx, query, jobID, payload, lockedBy)
	if err != nil {
		return fmt.Errorf("update bulk job summary: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("bulk outreach job lease was lost before its summary could be updated")
	}
	return nil
}

func (service *Service) DeferBulkJob(
	ctx context.Context,
	jobID string,
	lockedBy string,
	triggeredBy uuid.UUID,
	summary BulkSendSummary,
) error {
	if summary.NextAvailableAt == nil {
		return fmt.Errorf("next available time is required to defer bulk outreach")
	}
	payload, err := encodeBulkSummary(summary, triggeredBy.String())
	if err != nil {
		return err
	}
	const query = `
		UPDATE job_runs
		SET status = 'queued',
		    payload = $2,
		    available_at = $3,
		    attempts = 0,
		    locked_at = NULL,
		    locked_by = NULL,
		    lease_expires_at = NULL,
		    updated_at = now()
		WHERE id = $1::uuid
		  AND job_type = $4
		  AND status = 'running'
		  AND locked_by = $5`
	result, err := service.pool.Exec(
		ctx,
		query,
		jobID,
		payload,
		summary.NextAvailableAt.UTC(),
		BulkSendJobType,
		lockedBy,
	)
	if err != nil {
		return fmt.Errorf("defer bulk outreach job: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("bulk outreach job is not running and cannot be deferred")
	}
	return nil
}
