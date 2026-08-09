package outreach

import (
	"context"
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
		control, err := DisableEmailJobControl(ctx, service.pool)
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
	pending, err := service.repo.CountEligibleLeads(ctx)
	if err != nil {
		return EmailJobActionResult{}, err
	}
	control, activeJobID, err := EnableEmailJobControl(ctx, service.pool, principal.UserID)
	if err != nil {
		return EmailJobActionResult{}, err
	}
	if activeJobID != "" {
		return EmailJobActionResult{
			EmailJob:             control,
			JobID:                activeJobID,
			Status:               "resuming",
			MaxSends:             service.outreachCfg.BulkMax,
			PendingEligibleCount: pending,
		}, nil
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

// SettleDisabledBulkJob atomically hands a running disabled job either back to
// the queue when a concurrent re-enable won, or to completed when the control
// remains disabled. Both this method and EnableEmailJobControl lock the control
// row before the job row, closing the enabled-without-an-active-job race.
func (service *Service) SettleDisabledBulkJob(
	ctx context.Context,
	jobID string,
	lockedBy string,
	triggeredBy uuid.UUID,
	summary BulkSendSummary,
) error {
	if service.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	payload, err := encodeBulkSummary(summary, triggeredBy.String())
	if err != nil {
		return err
	}
	tx, err := service.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin settling disabled outreach job: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var enabled bool
	if err := tx.QueryRow(ctx, `
		SELECT enabled FROM outreach_runtime_control
		WHERE control_key = 'email_job'
		FOR UPDATE`).Scan(&enabled); err != nil {
		return fmt.Errorf("lock outreach email control while settling job: %w", err)
	}
	status := "completed"
	availableAt := time.Now().UTC()
	if enabled {
		status = "queued"
	}
	result, err := tx.Exec(ctx, `
		UPDATE job_runs
		SET status = $4,
		    payload = $2,
		    available_at = $5,
		    attempts = CASE WHEN $4 = 'queued' THEN 0 ELSE attempts END,
		    locked_at = NULL,
		    locked_by = NULL,
		    lease_expires_at = NULL,
		    updated_at = now()
		WHERE id = $1::uuid
		  AND job_type = $6
		  AND status = 'running'
		  AND locked_by = $3`,
		jobID, payload, lockedBy, status, availableAt, BulkSendJobType,
	)
	if err != nil {
		return fmt.Errorf("settle disabled outreach job: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("bulk outreach job lease was lost while settling disabled control")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit disabled outreach job settlement: %w", err)
	}
	return nil
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
	if counter, ok := service.repo.(interface {
		CountRecipientStatuses(context.Context) (RecipientStatusCounts, error)
	}); ok {
		counts, countErr := counter.CountRecipientStatuses(ctx)
		if countErr != nil {
			return StatusResult{}, countErr
		}
		result.DueFollowupCount = counts.DueFollowups
		result.NewRecipientCount = counts.NewRecipients
		result.PausedRecipientCount = counts.Paused
		result.CompletedRecipientCount = counts.Completed
	} else {
		result.NewRecipientCount = pending
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
		summary.StoppedReason = "waiting_for_leads"
		pollAt := time.Now().UTC().Add(5 * time.Minute)
		if sequenceRepo, ok := service.repo.(sequenceDeliveryRepository); ok {
			dueAt, dueErr := sequenceRepo.NextSequenceDueAt(ctx)
			if dueErr != nil {
				return summary, dueErr
			}
			if dueAt != nil && dueAt.Before(pollAt) {
				pollAt = dueAt.UTC()
				summary.StoppedReason = "waiting_for_followup"
			}
		}
		summary.NextAvailableAt = &pollAt
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
		} else {
			pollAt := time.Now().UTC().Add(5 * time.Minute)
			if sequenceRepo, ok := service.repo.(sequenceDeliveryRepository); ok {
				dueAt, dueErr := sequenceRepo.NextSequenceDueAt(ctx)
				if dueErr != nil {
					return summary, dueErr
				}
				if dueAt != nil && dueAt.Before(pollAt) {
					pollAt = dueAt.UTC()
					summary.StoppedReason = "waiting_for_followup"
				} else {
					summary.StoppedReason = "waiting_for_leads"
				}
			}
			summary.NextAvailableAt = &pollAt
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
	sequenceRepo, ok := service.repo.(sequenceDeliveryRepository)
	if !ok {
		return false, fmt.Errorf("sequence delivery repository is not configured")
	}
	if lead.Step < 1 {
		return false, fmt.Errorf("%w: sequence step must be 1-based", campaigns.ErrNotEligible)
	}
	delivery, err := sequenceRepo.GetSequenceDelivery(ctx, lead.CampaignID, lead.Step)
	if err != nil {
		return false, err
	}
	if delivery.RestaurantID != lead.RestaurantID {
		return false, fmt.Errorf("sequence campaign does not match the eligible lead")
	}
	suppressed, err := service.campaigns.IsSuppressed(ctx, delivery.RecipientEmail)
	if err != nil {
		return false, fmt.Errorf("check suppression: %w", err)
	}
	if err := checkSequenceDeliveryEligibility(delivery, suppressed); err != nil {
		return false, err
	}
	unsubscribeURL, err := service.campaignService.BuildUnsubscribeURL(
		ctx, delivery.CampaignID, delivery.RestaurantID, delivery.RecipientEmail,
	)
	if err != nil {
		return false, fmt.Errorf("build unsubscribe url: %w", err)
	}
	rendered, err := renderSequenceStep(
		delivery.Step, delivery.RestaurantName, delivery.OwnerFirstName, unsubscribeURL,
	)
	if err != nil {
		return false, fmt.Errorf("render outreach sequence step: %w", err)
	}
	if err := sequenceRepo.PrepareSequenceDelivery(ctx, delivery.CampaignID, delivery.Step.Position, rendered.Subject, rendered.BodyText); err != nil {
		return false, err
	}

	result, err := service.emailPool.Send(ctx, emailprovider.SendRequest{
		To:       delivery.RecipientEmail,
		Subject:  rendered.Subject,
		HTMLBody: "",
		TextBody: rendered.BodyText,
		Metadata: map[string]string{
			"campaign_id":   delivery.CampaignID.String(),
			"restaurant_id": delivery.RestaurantID.String(),
			"bulk_outreach": "true",
			"sequence_step": fmt.Sprintf("%d", delivery.Step.Position),
		},
		Delivery: &emailprovider.DeliveryContext{
			CampaignID:   delivery.CampaignID,
			RestaurantID: delivery.RestaurantID,
			BulkJobID:    bulkJobID,
			Step:         delivery.Step.Position,
			CampaignArtifactFingerprint: emailprovider.CampaignArtifactFingerprint(
				rendered.Subject,
				"",
				rendered.BodyText,
				"",
			),
		},
	})
	if err != nil {
		if !result.QuotaManaged {
			finalizeErr := sequenceRepo.FinalizeSequenceDelivery(ctx, SequenceDeliveryFinalization{
				CampaignID:     delivery.CampaignID,
				RestaurantID:   delivery.RestaurantID,
				Step:           delivery.Step.Position,
				RecipientEmail: delivery.RecipientEmail,
				Outcome:        sequenceOutcomeUnknown,
				ErrorCode:      "provider_send_unknown",
			})
			if finalizeErr != nil {
				return false, fmt.Errorf("provider send failed: %v; finalize unknown sequence delivery: %w", err, finalizeErr)
			}
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
				"campaign_id", delivery.CampaignID,
				"redirected", result.RedirectedTo != "",
				"account_key", result.AccountKey,
			)
			return false, nil
		}
		if err := sequenceRepo.FinalizeSequenceDelivery(ctx, SequenceDeliveryFinalization{
			CampaignID:     delivery.CampaignID,
			RestaurantID:   delivery.RestaurantID,
			Step:           delivery.Step.Position,
			RecipientEmail: delivery.RecipientEmail,
			Outcome:        sequenceOutcomeSkipped,
			Skipped:        result.Skipped,
			Redirected:     result.RedirectedTo != "",
		}); err != nil {
			return false, err
		}
		service.log.InfoContext(ctx, "bulk_outreach_lead_skipped",
			"restaurant_id", lead.RestaurantID,
			"campaign_id", delivery.CampaignID,
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
			"campaign_id", delivery.CampaignID,
			"account_key", result.AccountKey,
			"account_sequence", result.AccountSequence,
			"send_sequence", result.SendSequence,
		)
		return true, nil
	}

	if err := sequenceRepo.FinalizeSequenceDelivery(ctx, SequenceDeliveryFinalization{
		CampaignID:        delivery.CampaignID,
		RestaurantID:      delivery.RestaurantID,
		Step:              delivery.Step.Position,
		RecipientEmail:    delivery.RecipientEmail,
		Outcome:           sequenceOutcomeSent,
		ProviderMessageID: result.ProviderMessageID,
	}); err != nil {
		return false, err
	}

	service.log.InfoContext(ctx, "bulk_outreach_lead_sent",
		"restaurant_id", lead.RestaurantID,
		"campaign_id", delivery.CampaignID,
	)
	return true, nil
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
