package outreach

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

const (
	scheduledSendTimezone       = "Australia/Sydney"
	defaultScheduledStartMinute = 7 * 60
	defaultScheduledEndMinute   = 12 * 60
)

type EmailSendSchedule struct {
	Timezone  string     `json:"timezone"`
	StartTime string     `json:"start_time"`
	EndTime   string     `json:"end_time"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UpdateEmailSendScheduleInput struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type storedEmailSendSchedule struct {
	EmailSendSchedule
	startMinute int
	endMinute   int
}

type emailScheduleRowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func defaultEmailSendSchedule() storedEmailSendSchedule {
	return newStoredEmailSendSchedule(defaultScheduledStartMinute, defaultScheduledEndMinute, nil, time.Time{})
}

func newStoredEmailSendSchedule(startMinute, endMinute int, updatedBy *uuid.UUID, updatedAt time.Time) storedEmailSendSchedule {
	return storedEmailSendSchedule{
		EmailSendSchedule: EmailSendSchedule{
			Timezone:  scheduledSendTimezone,
			StartTime: formatScheduleMinute(startMinute),
			EndTime:   formatScheduleMinute(endMinute),
			UpdatedBy: updatedBy,
			UpdatedAt: updatedAt,
		},
		startMinute: startMinute,
		endMinute:   endMinute,
	}
}

func formatScheduleMinute(minute int) string {
	return fmt.Sprintf("%02d:%02d", minute/60, minute%60)
}

func parseScheduleMinute(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	parsed, err := time.Parse("15:04", trimmed)
	if err != nil || parsed.Format("15:04") != trimmed {
		return 0, fmt.Errorf("%w: times must use 24-hour HH:MM format", ErrInvalidSendSchedule)
	}
	return parsed.Hour()*60 + parsed.Minute(), nil
}

func validateEmailSendSchedule(input UpdateEmailSendScheduleInput) (int, int, error) {
	startMinute, err := parseScheduleMinute(input.StartTime)
	if err != nil {
		return 0, 0, err
	}
	endMinute, err := parseScheduleMinute(input.EndTime)
	if err != nil {
		return 0, 0, err
	}
	if endMinute <= startMinute {
		return 0, 0, fmt.Errorf("%w: end_time must be later than start_time on the same Sydney day", ErrInvalidSendSchedule)
	}
	if endMinute-startMinute < 60 {
		return 0, 0, fmt.Errorf("%w: the scheduled outreach window must be at least 60 minutes", ErrInvalidSendSchedule)
	}
	return startMinute, endMinute, nil
}

func validateEmailScheduleCapacity(
	window time.Duration,
	totalDailyLimit int,
	largestMailboxPacingRequirement time.Duration,
) error {
	if window <= 0 || totalDailyLimit < 0 || largestMailboxPacingRequirement < 0 {
		return fmt.Errorf("%w: scheduled outreach capacity is invalid", ErrInvalidSendSchedule)
	}
	if totalDailyLimit == 0 {
		return nil
	}
	// Mailboxes are interleaved through the saved window. Each mailbox's fixed
	// slot anchors preserve its own minimum cadence, while the aggregate gate may
	// advance more frequently as different mailboxes become due. Keep at least a
	// one-second provider-boundary gap and ensure the busiest mailbox can still
	// complete its own allowance at its configured minimum delay.
	if window/time.Duration(totalDailyLimit) < time.Second || window < largestMailboxPacingRequirement {
		return fmt.Errorf(
			"%w: the configured mailbox quota cannot finish inside this window at the minimum pacing interval",
			ErrInvalidSendSchedule,
		)
	}
	return nil
}

func loadEmailSendSchedule(ctx context.Context, querier emailScheduleRowQuerier, lock bool) (storedEmailSendSchedule, error) {
	query := `
		SELECT timezone, start_minute, end_minute, updated_by, updated_at
		FROM outreach_send_schedule
		WHERE singleton = 1`
	if lock {
		query += ` FOR UPDATE`
	}
	var timezone string
	var startMinute int
	var endMinute int
	var updatedBy *uuid.UUID
	var updatedAt time.Time
	if err := querier.QueryRow(ctx, query).Scan(&timezone, &startMinute, &endMinute, &updatedBy, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return storedEmailSendSchedule{}, fmt.Errorf("outreach send schedule is not initialized")
		}
		return storedEmailSendSchedule{}, fmt.Errorf("load outreach send schedule: %w", err)
	}
	if timezone != scheduledSendTimezone {
		return storedEmailSendSchedule{}, fmt.Errorf("unsupported outreach send schedule timezone %q", timezone)
	}
	return newStoredEmailSendSchedule(startMinute, endMinute, updatedBy, updatedAt.UTC()), nil
}

func GetEmailSendSchedule(ctx context.Context, pool *pgxpool.Pool) (EmailSendSchedule, error) {
	if pool == nil {
		return defaultEmailSendSchedule().EmailSendSchedule, nil
	}
	schedule, err := loadEmailSendSchedule(ctx, pool, false)
	return schedule.EmailSendSchedule, err
}

func (service *Service) SetEmailSendSchedule(
	ctx context.Context,
	principal auth.Principal,
	input UpdateEmailSendScheduleInput,
) (EmailSendSchedule, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return EmailSendSchedule{}, fmt.Errorf("forbidden")
	}
	startMinute, endMinute, err := validateEmailSendSchedule(input)
	if err != nil {
		return EmailSendSchedule{}, err
	}
	if service.pool == nil {
		return EmailSendSchedule{}, fmt.Errorf("database pool is not configured")
	}

	tx, err := service.pool.Begin(ctx)
	if err != nil {
		return EmailSendSchedule{}, fmt.Errorf("begin outreach send schedule update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var emailJobEnabled bool
	if err := tx.QueryRow(ctx, `
		SELECT enabled
		FROM outreach_runtime_control
		WHERE control_key = 'email_job'
		FOR UPDATE`).Scan(&emailJobEnabled); err != nil {
		return EmailSendSchedule{}, fmt.Errorf("lock outreach email job while updating schedule: %w", err)
	}
	var activeJob bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM job_runs
		  WHERE job_type = $1 AND status IN ('queued', 'running')
		)`, BulkSendJobType).Scan(&activeJob); err != nil {
		return EmailSendSchedule{}, fmt.Errorf("check active outreach job while updating schedule: %w", err)
	}
	if emailJobEnabled || activeJob {
		return EmailSendSchedule{}, ErrSendScheduleLocked
	}

	if _, err := loadEmailSendSchedule(ctx, tx, true); err != nil {
		return EmailSendSchedule{}, err
	}
	var totalDailyLimit int
	var largestMailboxPacingSeconds int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(sum(send_limit), 0)::int,
		       COALESCE(max(send_limit * send_jitter_min_seconds), 0)::int
		FROM outreach_email_accounts
		WHERE enabled = true`).Scan(&totalDailyLimit, &largestMailboxPacingSeconds); err != nil {
		return EmailSendSchedule{}, fmt.Errorf("load outreach mailbox capacity: %w", err)
	}
	window := time.Duration(endMinute-startMinute) * time.Minute
	if err := validateEmailScheduleCapacity(
		window,
		totalDailyLimit,
		time.Duration(largestMailboxPacingSeconds)*time.Second,
	); err != nil {
		return EmailSendSchedule{}, err
	}

	var updatedBy *uuid.UUID
	var updatedAt time.Time
	if err := tx.QueryRow(ctx, `
		UPDATE outreach_send_schedule
		SET start_minute = $1,
		    end_minute = $2,
		    updated_by = $3,
		    updated_at = now()
		WHERE singleton = 1
		RETURNING updated_by, updated_at`, startMinute, endMinute, principal.UserID).Scan(&updatedBy, &updatedAt); err != nil {
		return EmailSendSchedule{}, fmt.Errorf("update outreach send schedule: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return EmailSendSchedule{}, fmt.Errorf("commit outreach send schedule update: %w", err)
	}
	return newStoredEmailSendSchedule(startMinute, endMinute, updatedBy, updatedAt.UTC()).EmailSendSchedule, nil
}
