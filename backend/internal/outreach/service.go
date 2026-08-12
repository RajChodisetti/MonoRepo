package outreach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type Service struct {
	repo            Repository
	pool            *pgxpool.Pool
	campaigns       campaigns.Repository
	campaignService *campaigns.Service
	restaurants     *restaurants.Service
	tokenResolver   DemoTokenResolver
	emailPool       *emailprovider.AccountPool
	emailProvider   emailprovider.Provider
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
	restaurantsService *restaurants.Service,
	tokenResolver DemoTokenResolver,
	emailPool *emailprovider.AccountPool,
	emailProvider emailprovider.Provider,
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
		restaurants:     restaurantsService,
		tokenResolver:   tokenResolver,
		emailPool:       emailPool,
		emailProvider:   emailProvider,
		emailCfg:        emailCfg,
		outreachCfg:     outreachCfg,
		enqueuer:        enqueuer,
		log:             log,
	}
}

func (service *Service) SetEmailJob(ctx context.Context, principal auth.Principal, enabled bool) (EmailJobActionResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return EmailJobActionResult{}, fmt.Errorf("forbidden")
	}
	if !enabled {
		control, err := SetEmailJobControl(ctx, service.pool, false, nil)
		return EmailJobActionResult{
			EmailJob: control,
			Status:   "disabled",
			MaxSends: service.outreachCfg.BulkMax,
		}, err
	}
	if service.emailPool == nil {
		return EmailJobActionResult{}, ErrNotConfigured
	}
	if err := validateBulkMax(service.outreachCfg.BulkMax); err != nil {
		return EmailJobActionResult{}, err
	}
	active, _, err := HasActiveBulkJob(ctx, service.pool, BulkSendJobType)
	if err != nil {
		return EmailJobActionResult{}, err
	}
	if active {
		return EmailJobActionResult{}, ErrBulkJobActive
	}
	pending, err := service.repo.CountEligibleLeads(ctx)
	if err != nil {
		return EmailJobActionResult{}, err
	}
	control, err := SetEmailJobControl(ctx, service.pool, true, &principal.UserID)
	if err != nil {
		return EmailJobActionResult{}, err
	}
	jobID, err := service.enqueuer.EnqueueBulkSend(ctx, principal.UserID)
	if err != nil {
		// A concurrent admin activation may have won the queue's unique active-job
		// constraint after our read check. Keep the shared control enabled for that
		// winning job; only revert when no active job was created.
		if !errors.Is(err, ErrBulkJobActive) {
			_, _ = SetEmailJobControl(ctx, service.pool, false, nil)
		}
		return EmailJobActionResult{}, err
	}
	return EmailJobActionResult{
		EmailJob:             control,
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
	control, err := GetEmailJobControl(ctx, service.pool)
	if err != nil {
		return StatusResult{}, err
	}
	result.EmailJob = control
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

	jobControl, err := GetEmailJobControl(ctx, service.pool)
	if err != nil {
		return summary, err
	}
	if service.pool != nil && !jobControl.Enabled {
		summary.StoppedReason = "disabled_by_admin"
		return summary, nil
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
		_, _ = SetEmailJobControl(ctx, service.pool, false, nil)
		return summary, nil
	}

	for _, lead := range leads {
		if service.pool != nil {
			control, controlErr := GetEmailJobControl(ctx, service.pool)
			if controlErr != nil {
				return summary, controlErr
			}
			if !control.Enabled {
				summary.StoppedReason = "disabled_by_admin"
				break
			}
		}
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
			_, _ = SetEmailJobControl(ctx, service.pool, false, nil)
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
	if summary.NextAvailableAt == nil {
		_, _ = SetEmailJobControl(ctx, service.pool, false, nil)
	}

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
	}); err != nil {
		return false, err
	}

	trackingURLs, err := service.campaignService.BuildTrackingURLs(ctx, campaign, sendCtx)
	if err != nil {
		return false, fmt.Errorf("build tracking urls: %w", err)
	}

	canonicalContent := campaigns.EnsureOutreachSignature(campaigns.DraftContent{
		Subject:  campaign.Subject,
		BodyHTML: campaign.BodyHTML,
		BodyText: campaign.BodyText,
	})
	draft := campaigns.InjectTracking(canonicalContent, trackingURLs, service.emailCfg.OpenTrackingEnabled)
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
				canonicalContent.Subject,
				canonicalContent.BodyHTML,
				canonicalContent.BodyText,
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

// latestCampaignContent returns the most recently created campaign for a
// restaurant (any status), which carries the subject/HTML/text rendered at
// draft/regenerate time. Ad hoc send/preview deliberately does not require
// approval/publish/OCR gates, but it still needs real rendered content to
// send, so it reuses whatever draft already exists rather than rendering ad
// hoc content outside the audited draft lifecycle.
func (service *Service) latestCampaignContent(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (campaigns.Campaign, error) {
	records, err := service.campaignService.ListByRestaurant(ctx, principal, restaurantID)
	if err != nil {
		return campaigns.Campaign{}, err
	}
	if len(records) == 0 {
		return campaigns.Campaign{}, ErrNoCampaignDraft
	}
	return records[0], nil
}

// PreviewAdHoc renders what an ad hoc send would deliver right now, without
// sending or mutating any state.
func (service *Service) PreviewAdHoc(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (AdHocPreview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return AdHocPreview{}, restaurants.ErrForbidden
	}

	restaurant, err := service.restaurants.GetRestaurant(ctx, principal, restaurantID)
	if err != nil {
		return AdHocPreview{}, err
	}

	campaign, err := service.latestCampaignContent(ctx, principal, restaurantID)
	if err != nil {
		return AdHocPreview{}, err
	}

	canonicalContent := campaigns.EnsureOutreachSignature(campaigns.DraftContent{
		Subject:  campaign.Subject,
		BodyHTML: campaign.BodyHTML,
		BodyText: campaign.BodyText,
	})

	return AdHocPreview{
		RestaurantID:   restaurantID,
		RestaurantName: restaurant.Name,
		RecipientEmail: restaurant.Email,
		Subject:        canonicalContent.Subject,
		BodyHTML:       canonicalContent.BodyHTML,
		BodyText:       canonicalContent.BodyText,
	}, nil
}

// SendAdHoc sends the latest campaign draft's rendered content to a single
// restaurant immediately, outside the quota-managed bulk pipeline. It still
// enforces the global sending kill switch and the opt-out suppression list —
// those are compliance requirements, not internal review-workflow gates.
func (service *Service) SendAdHoc(ctx context.Context, principal auth.Principal, restaurantID uuid.UUID) (AdHocSendResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return AdHocSendResult{RestaurantID: restaurantID}, restaurants.ErrForbidden
	}
	if !service.emailSendingEnabled() {
		return AdHocSendResult{RestaurantID: restaurantID}, ErrSendingDisabled
	}
	if service.emailProvider == nil {
		return AdHocSendResult{RestaurantID: restaurantID}, ErrNotConfigured
	}

	restaurant, err := service.restaurants.GetRestaurant(ctx, principal, restaurantID)
	if err != nil {
		return AdHocSendResult{RestaurantID: restaurantID}, err
	}
	email := strings.TrimSpace(restaurant.Email)
	if email == "" {
		return AdHocSendResult{RestaurantID: restaurantID}, ErrNoContactEmail
	}

	campaign, err := service.latestCampaignContent(ctx, principal, restaurantID)
	if err != nil {
		return AdHocSendResult{RestaurantID: restaurantID}, err
	}

	canonicalContent := campaigns.EnsureOutreachSignature(campaigns.DraftContent{
		Subject:  campaign.Subject,
		BodyHTML: campaign.BodyHTML,
		BodyText: campaign.BodyText,
	})

	_, err = service.emailProvider.Send(ctx, emailprovider.SendRequest{
		To:       email,
		Subject:  canonicalContent.Subject,
		HTMLBody: canonicalContent.BodyHTML,
		TextBody: canonicalContent.BodyText,
		Metadata: map[string]string{
			"restaurant_id": restaurantID.String(),
			"campaign_id":   campaign.ID.String(),
			"send_type":     "adhoc",
		},
	})
	if err != nil {
		return AdHocSendResult{RestaurantID: restaurantID}, fmt.Errorf("send ad hoc email: %w", err)
	}

	if err := service.repo.RecordAdHocEmailSent(ctx, restaurantID, email); err != nil {
		service.log.ErrorContext(ctx, "adhoc_email_sent_but_record_failed",
			"restaurant_id", restaurantID.String(), "error", err)
	}

	return AdHocSendResult{RestaurantID: restaurantID, Sent: true}, nil
}

// SendAdHocBatch sends to multiple restaurants, collecting a per-restaurant
// result rather than failing the whole batch on one lead's error. Capped to
// keep this synchronous HTTP-request path bounded — unlike the durable bulk
// pipeline, ad hoc sends have no async job queue or pacing.
const adHocBatchLimit = 25

func (service *Service) SendAdHocBatch(ctx context.Context, principal auth.Principal, restaurantIDs []uuid.UUID) ([]AdHocSendResult, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return nil, restaurants.ErrForbidden
	}
	if len(restaurantIDs) == 0 {
		return nil, fmt.Errorf("restaurant_ids must not be empty")
	}
	if len(restaurantIDs) > adHocBatchLimit {
		return nil, fmt.Errorf("cannot send to more than %d restaurants at once", adHocBatchLimit)
	}

	results := make([]AdHocSendResult, 0, len(restaurantIDs))
	for _, restaurantID := range restaurantIDs {
		result, err := service.SendAdHoc(ctx, principal, restaurantID)
		if err != nil {
			result.RestaurantID = restaurantID
			result.Error = err.Error()
		}
		results = append(results, result)
	}
	return results, nil
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
