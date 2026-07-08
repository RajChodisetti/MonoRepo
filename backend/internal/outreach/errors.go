package outreach

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const BulkSendJobType = "outreach.bulk_send"

type BulkJobEnqueuer interface {
	EnqueueBulkSend(ctx context.Context, triggeredBy uuid.UUID) (jobID string, err error)
}

var (
	ErrBulkJobActive = errors.New("a bulk outreach job is already queued or running")
	ErrNotConfigured = errors.New("bulk outreach email accounts are not configured")
)

type BulkSendSummary struct {
	Attempted     int    `json:"attempted"`
	Sent          int    `json:"sent"`
	Failed        int    `json:"failed"`
	Skipped       int    `json:"skipped"`
	MaxSends      int    `json:"max_sends"`
	StoppedReason string `json:"stopped_reason,omitempty"`
}

type TriggerResult struct {
	JobID                string `json:"job_id"`
	Status               string `json:"status"`
	MaxSends             int    `json:"max_sends"`
	PendingEligibleCount int    `json:"pending_eligible_count"`
}

type StatusResult struct {
	PendingEligibleCount int                 `json:"pending_eligible_count"`
	MaxSends             int                 `json:"max_sends"`
	ActiveJob            *ActiveJobStatus    `json:"active_job,omitempty"`
	LastCompletedJob     *CompletedJobStatus `json:"last_completed_job,omitempty"`
}

type ActiveJobStatus struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type CompletedJobStatus struct {
	JobID   string          `json:"job_id"`
	Status  string          `json:"status"`
	Summary BulkSendSummary `json:"summary"`
}

func validateBulkMax(max int) error {
	if max < 1 {
		return fmt.Errorf("bulk max sends must be at least 1")
	}
	return nil
}
