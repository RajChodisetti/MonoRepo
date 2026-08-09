package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
)

type OutreachBulkEnqueuer struct {
	Queue Queue
}

func (enqueuer *OutreachBulkEnqueuer) EnqueueBulkSend(ctx context.Context, triggeredBy uuid.UUID) (string, error) {
	job, err := NewOutreachBulkSendJob(triggeredBy)
	if err != nil {
		return "", err
	}
	queued, err := enqueuer.Queue.Enqueue(ctx, job)
	if err != nil {
		if errors.Is(err, ErrActiveBulkJob) {
			return "", outreach.ErrBulkJobActive
		}
		return "", err
	}
	return queued.ID, nil
}

type OutreachBulkSendDeps struct {
	Outreach *outreach.Service
}

func OutreachBulkSendHandler(deps OutreachBulkSendDeps, log *slog.Logger) Handler {
	return func(ctx context.Context, job Job) error {
		var payload OutreachBulkSendPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}

		summary, err := deps.Outreach.RunBulkSend(ctx, payload.TriggeredBy, job.ID)
		if err == nil && summary.StoppedReason == "disabled_by_admin" && summary.NextAvailableAt == nil {
			if settleErr := deps.Outreach.SettleDisabledBulkJob(ctx, job.ID, job.LockedBy, payload.TriggeredBy, summary); settleErr != nil {
				return settleErr
			}
			// The service already moved the row to queued or completed atomically.
			// Prevent the generic worker finisher from racing that fenced handoff.
			return ErrJobDeferred
		}
		if err == nil && summary.NextAvailableAt != nil {
			if deferErr := deps.Outreach.DeferBulkJob(ctx, job.ID, job.LockedBy, payload.TriggeredBy, summary); deferErr != nil {
				return deferErr
			}
			return ErrJobDeferred
		}
		if updateErr := deps.Outreach.UpdateJobSummary(ctx, job.ID, job.LockedBy, payload.TriggeredBy, summary); updateErr != nil {
			log.ErrorContext(ctx, "bulk_outreach_summary_update_failed", "error", updateErr)
			return updateErr
		}
		return err
	}
}
