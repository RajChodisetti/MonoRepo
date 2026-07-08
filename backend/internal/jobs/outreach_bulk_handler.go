package jobs

import (
	"context"
	"encoding/json"
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

		summary, err := deps.Outreach.RunBulkSend(ctx, payload.TriggeredBy)
		if updateErr := deps.Outreach.UpdateJobSummary(ctx, job.ID, payload.TriggeredBy, summary); updateErr != nil {
			log.WarnContext(ctx, "bulk_outreach_summary_update_failed", "error", updateErr)
		}
		return err
	}
}
