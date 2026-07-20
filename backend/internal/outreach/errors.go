package outreach

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const BulkSendJobType = "outreach.bulk_send"

type BulkJobEnqueuer interface {
	EnqueueBulkSend(ctx context.Context, triggeredBy uuid.UUID) (jobID string, err error)
}

var (
	ErrBulkJobActive   = errors.New("a bulk outreach job is already queued or running")
	ErrNotConfigured   = errors.New("bulk outreach email accounts are not configured")
	ErrSendingDisabled = errors.New("email sending is disabled")

	// Ad hoc send errors (single-lead / multi-select, outside the quota-managed
	// bulk pipeline).
	ErrNoContactEmail  = errors.New("restaurant has no valid contact email")
	ErrEmailSuppressed = errors.New("recipient has opted out of outreach email")
	ErrNoCampaignDraft = errors.New("no campaign draft exists yet for this restaurant")
	ErrDeliverySkipped = errors.New("email delivery was skipped")
)

// AdHocSendResult is the per-restaurant outcome of an ad hoc (non-bulk) send.
type AdHocSendResult struct {
	RestaurantID uuid.UUID `json:"restaurant_id"`
	Sent         bool      `json:"sent"`
	Error        string    `json:"error,omitempty"`
}

// AdHocPreview is the rendered content of the latest campaign draft for a
// restaurant, shown before an ad hoc send is confirmed.
type AdHocPreview struct {
	RestaurantID   uuid.UUID `json:"restaurant_id"`
	RestaurantName string    `json:"restaurant_name"`
	RecipientEmail string    `json:"recipient_email"`
	Subject        string    `json:"subject"`
	BodyHTML       string    `json:"body_html"`
	BodyText       string    `json:"body_text"`
}

type BulkSendSummary struct {
	Attempted       int        `json:"attempted"`
	Sent            int        `json:"sent"`
	Failed          int        `json:"failed"`
	Skipped         int        `json:"skipped"`
	MaxSends        int        `json:"max_sends"`
	StoppedReason   string     `json:"stopped_reason,omitempty"`
	NextAvailableAt *time.Time `json:"next_available_at,omitempty"`
}

type EmailJobControl struct {
	Enabled   bool       `json:"enabled"`
	EnabledAt *time.Time `json:"enabled_at,omitempty"`
	EnabledBy *uuid.UUID `json:"enabled_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type EmailJobActionResult struct {
	EmailJob             EmailJobControl `json:"email_job"`
	JobID                string          `json:"job_id,omitempty"`
	Status               string          `json:"status"`
	MaxSends             int             `json:"max_sends"`
	PendingEligibleCount int             `json:"pending_eligible_count"`
}

type StatusResult struct {
	PendingEligibleCount int                 `json:"pending_eligible_count"`
	MaxSends             int                 `json:"max_sends"`
	ActiveJob            *ActiveJobStatus    `json:"active_job,omitempty"`
	LastCompletedJob     *CompletedJobStatus `json:"last_completed_job,omitempty"`
	NextAvailableAt      *time.Time          `json:"next_available_at,omitempty"`
	EmailJob             EmailJobControl     `json:"email_job"`
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
	if max < 1 || max > 150 {
		return fmt.Errorf("bulk max sends must be between 1 and 150")
	}
	return nil
}
