package outreach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

const (
	SequenceStatusDraft    = "draft"
	SequenceStatusApproved = "approved"
	SequenceStatusArchived = "archived"

	websiteURL = "https://tuvisolutions.com"
)

var (
	templatePlaceholderPattern = regexp.MustCompile(`\{\{[a-z_]+\}\}`)
	rawURLPattern              = regexp.MustCompile(`(?i)https?://[^\s<>"']+`)
	bareDomainPattern          = regexp.MustCompile(`(?i)(?:www\.)?(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}(?::[0-9]{1,5})?(?:/[^\s<>"']*)?`)
	htmlElementPattern         = regexp.MustCompile(`(?i)<[a-z][^>]*>`)
	validLeadEmailPattern      = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

	ErrSequenceInvalid = errors.New("outreach sequence is invalid")
	ErrSequenceStale   = errors.New("outreach sequence changed after it was loaded")
)

type Sequence struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Version    int            `json:"version"`
	Status     string         `json:"status"`
	IsActive   bool           `json:"is_active"`
	ApprovedAt *time.Time     `json:"approved_at,omitempty"`
	ApprovedBy *uuid.UUID     `json:"approved_by,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Steps      []SequenceStep `json:"steps"`
}

type SequenceStep struct {
	ID               uuid.UUID `json:"id"`
	SequenceID       uuid.UUID `json:"sequence_id"`
	Position         int       `json:"position"`
	Enabled          bool      `json:"enabled"`
	DelayHours       int       `json:"delay_hours"`
	SubjectTemplate  string    `json:"subject_template"`
	BodyTextTemplate string    `json:"body_text_template"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SequenceList struct {
	ActiveSequenceID *uuid.UUID `json:"active_sequence_id,omitempty"`
	Sequences        []Sequence `json:"sequences"`
}

type CreateSequenceInput struct {
	Name              string     `json:"name"`
	BasedOnSequenceID *uuid.UUID `json:"based_on_sequence_id,omitempty"`
}

type UpdateSequenceInput struct {
	Name              string         `json:"name"`
	ExpectedUpdatedAt time.Time      `json:"expected_updated_at"`
	Steps             []SequenceStep `json:"steps"`
}

type PreviewSequenceInput struct {
	RestaurantName string `json:"restaurant_name"`
	OwnerFirstName string `json:"owner_first_name,omitempty"`
}

type RenderedSequenceStep struct {
	Position   int    `json:"position"`
	DelayHours int    `json:"delay_hours"`
	Subject    string `json:"subject"`
	BodyText   string `json:"body_text"`
	URLCount   int    `json:"url_count"`
}

type SequencePreview struct {
	RestaurantName string                 `json:"restaurant_name"`
	Greeting       string                 `json:"greeting"`
	Steps          []RenderedSequenceStep `json:"steps"`
}

type RecipientProgress struct {
	RestaurantID    uuid.UUID  `json:"restaurant_id"`
	RestaurantName  string     `json:"restaurant_name"`
	Email           string     `json:"email"`
	LifecycleStatus string     `json:"lifecycle_status"`
	ConsentBasis    string     `json:"consent_basis"`
	CurrentStep     int        `json:"current_step"`
	NextStep        *int       `json:"next_step,omitempty"`
	NextSendAt      *time.Time `json:"next_send_at,omitempty"`
	LastSentAt      *time.Time `json:"last_sent_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	EmailSendCount  int        `json:"email_send_count"`
	CampaignStatus  string     `json:"campaign_status"`
	Suppressed      bool       `json:"suppressed"`
	Eligible        bool       `json:"eligible"`
	HoldReason      string     `json:"hold_reason,omitempty"`
}

type RecipientProgressList struct {
	Recipients []RecipientProgress `json:"recipients"`
	Total      int                 `json:"total"`
}

type SequenceDelivery struct {
	CampaignID        uuid.UUID
	RestaurantID      uuid.UUID
	RestaurantName    string
	OwnerFirstName    string
	RecipientEmail    string
	LifecycleStatus   string
	ShownInterest     bool
	ConsentBasis      string
	ConsentSource     string
	ConsentRecordedAt *time.Time
	ConsentEvidence   json.RawMessage
	SequenceStatus    string
	Step              SequenceStep
}

type activeSequenceStepsRepository interface {
	ListActiveSequenceSteps(ctx context.Context) ([]SequenceStep, error)
}

const rebaseUntouchedEnrollmentsQuery = `
	WITH first_enabled AS (
	  SELECT position
	  FROM outreach_email_sequence_steps
	  WHERE sequence_id = $1 AND enabled = true
	  ORDER BY position
	  LIMIT 1
	)
	UPDATE email_campaigns campaign
	SET sequence_id = $1,
	    next_step = (SELECT position FROM first_enabled),
	    next_send_at = now(),
	    subject = '',
	    body_html = '',
	    body_text = '',
	    demo_token = '',
	    updated_at = now()
	WHERE campaign.campaign_type = 'outreach'
	  AND campaign.sequence_id IS NOT NULL
	  AND campaign.status = 'approved'
	  AND campaign.current_step = 0
	  AND campaign.last_sent_at IS NULL
	  AND campaign.completed_at IS NULL
	  AND NOT EXISTS (
	    SELECT 1 FROM email_delivery_attempts attempt
	    WHERE attempt.campaign_id = campaign.id
	  )
	  AND NOT EXISTS (
	    SELECT 1 FROM email_events event
	    WHERE event.campaign_id = campaign.id
	      AND event.event_type IN ('sent', 'skipped', 'failed')
	  )`

func (service *Service) ListSequences(ctx context.Context, principal auth.Principal) (SequenceList, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return SequenceList{}, restaurants.ErrForbidden
	}
	if service.pool == nil {
		return SequenceList{}, fmt.Errorf("database pool is not configured")
	}
	rows, err := service.pool.Query(ctx, `
		SELECT id, name, version, status, is_active, approved_at, approved_by, created_at, updated_at
		FROM outreach_email_sequences
		ORDER BY created_at DESC, version DESC`)
	if err != nil {
		return SequenceList{}, fmt.Errorf("list outreach sequences: %w", err)
	}
	defer rows.Close()
	result := SequenceList{Sequences: []Sequence{}}
	for rows.Next() {
		var sequence Sequence
		if err := rows.Scan(
			&sequence.ID, &sequence.Name, &sequence.Version, &sequence.Status,
			&sequence.IsActive, &sequence.ApprovedAt, &sequence.ApprovedBy,
			&sequence.CreatedAt, &sequence.UpdatedAt,
		); err != nil {
			return SequenceList{}, fmt.Errorf("scan outreach sequence: %w", err)
		}
		if sequence.IsActive {
			id := sequence.ID
			result.ActiveSequenceID = &id
		}
		result.Sequences = append(result.Sequences, sequence)
	}
	if err := rows.Err(); err != nil {
		return SequenceList{}, fmt.Errorf("iterate outreach sequences: %w", err)
	}
	for index := range result.Sequences {
		steps, err := service.listSequenceSteps(ctx, result.Sequences[index].ID)
		if err != nil {
			return SequenceList{}, err
		}
		result.Sequences[index].Steps = steps
	}
	return result, nil
}

func (service *Service) CreateSequenceDraft(ctx context.Context, principal auth.Principal, input CreateSequenceInput) (Sequence, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Sequence{}, restaurants.ErrForbidden
	}
	if service.pool == nil {
		return Sequence{}, fmt.Errorf("database pool is not configured")
	}
	name := cleanSingleLine(input.Name)
	if name == "" {
		name = "Tuvi restaurant introduction"
	}
	if len(name) > 120 {
		return Sequence{}, fmt.Errorf("%w: name must be at most 120 characters", ErrSequenceInvalid)
	}
	tx, err := service.pool.Begin(ctx)
	if err != nil {
		return Sequence{}, fmt.Errorf("begin outreach sequence draft: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	baseID := input.BasedOnSequenceID
	if baseID == nil {
		var active uuid.UUID
		if err := tx.QueryRow(ctx, `
			SELECT id FROM outreach_email_sequences
			WHERE is_active = true AND status = 'approved'
			LIMIT 1`).Scan(&active); err == nil {
			baseID = &active
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return Sequence{}, fmt.Errorf("load active outreach sequence: %w", err)
		}
	}
	var sequence Sequence
	err = tx.QueryRow(ctx, `
		INSERT INTO outreach_email_sequences (name, version, status, is_active)
		VALUES (
		  $1,
		  COALESCE((SELECT max(version) + 1 FROM outreach_email_sequences WHERE name = $1), 1),
		  'draft',
		  false
		)
		RETURNING id, name, version, status, is_active, approved_at, approved_by, created_at, updated_at`, name).Scan(
		&sequence.ID, &sequence.Name, &sequence.Version, &sequence.Status,
		&sequence.IsActive, &sequence.ApprovedAt, &sequence.ApprovedBy,
		&sequence.CreatedAt, &sequence.UpdatedAt,
	)
	if err != nil {
		return Sequence{}, fmt.Errorf("create outreach sequence draft: %w", err)
	}
	if baseID != nil {
		result, err := tx.Exec(ctx, `
			INSERT INTO outreach_email_sequence_steps (
			  sequence_id, position, enabled, delay_hours, subject_template, body_text_template
			)
			SELECT $1, position, enabled, delay_hours, subject_template, body_text_template
			FROM outreach_email_sequence_steps
			WHERE sequence_id = $2
			ORDER BY position`, sequence.ID, *baseID)
		if err != nil {
			return Sequence{}, fmt.Errorf("copy outreach sequence steps: %w", err)
		}
		if result.RowsAffected() == 0 {
			return Sequence{}, fmt.Errorf("%w: base sequence was not found or has no steps", ErrSequenceInvalid)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Sequence{}, fmt.Errorf("commit outreach sequence draft: %w", err)
	}
	sequence.Steps, err = service.listSequenceSteps(ctx, sequence.ID)
	return sequence, err
}

func (service *Service) UpdateSequenceDraft(ctx context.Context, principal auth.Principal, sequenceID uuid.UUID, input UpdateSequenceInput) (Sequence, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Sequence{}, restaurants.ErrForbidden
	}
	if input.ExpectedUpdatedAt.IsZero() {
		return Sequence{}, fmt.Errorf("%w: expected_updated_at is required", ErrSequenceInvalid)
	}
	name := cleanSingleLine(input.Name)
	if name == "" || len(name) > 120 {
		return Sequence{}, fmt.Errorf("%w: name must be between 1 and 120 characters", ErrSequenceInvalid)
	}
	steps := append([]SequenceStep(nil), input.Steps...)
	if err := validateSequenceSteps(steps); err != nil {
		return Sequence{}, err
	}
	tx, err := service.pool.Begin(ctx)
	if err != nil {
		return Sequence{}, fmt.Errorf("begin outreach sequence update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status string
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
		SELECT status, updated_at FROM outreach_email_sequences WHERE id = $1 FOR UPDATE`, sequenceID).Scan(&status, &currentUpdatedAt); errors.Is(err, pgx.ErrNoRows) {
		return Sequence{}, repository.ErrNotFound
	} else if err != nil {
		return Sequence{}, fmt.Errorf("lock outreach sequence: %w", err)
	}
	if status != SequenceStatusDraft {
		return Sequence{}, fmt.Errorf("%w: only draft versions can be edited", ErrSequenceInvalid)
	}
	if !currentUpdatedAt.Equal(input.ExpectedUpdatedAt) {
		return Sequence{}, ErrSequenceStale
	}
	if _, err := tx.Exec(ctx, `DELETE FROM outreach_email_sequence_steps WHERE sequence_id = $1`, sequenceID); err != nil {
		return Sequence{}, fmt.Errorf("replace outreach sequence steps: %w", err)
	}
	for _, step := range steps {
		stepID := step.ID
		if stepID == uuid.Nil {
			stepID = uuid.New()
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO outreach_email_sequence_steps (
			  id, sequence_id, position, enabled, delay_hours, subject_template, body_text_template
			) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			stepID, sequenceID, step.Position, step.Enabled, step.DelayHours,
			strings.TrimSpace(step.SubjectTemplate), strings.TrimSpace(step.BodyTextTemplate),
		); err != nil {
			return Sequence{}, fmt.Errorf("save outreach sequence step %d: %w", step.Position, err)
		}
	}
	var sequence Sequence
	if err := tx.QueryRow(ctx, `
		UPDATE outreach_email_sequences
		SET name = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, name, version, status, is_active, approved_at, approved_by, created_at, updated_at`, sequenceID, name).Scan(
		&sequence.ID, &sequence.Name, &sequence.Version, &sequence.Status,
		&sequence.IsActive, &sequence.ApprovedAt, &sequence.ApprovedBy,
		&sequence.CreatedAt, &sequence.UpdatedAt,
	); err != nil {
		return Sequence{}, fmt.Errorf("update outreach sequence: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Sequence{}, fmt.Errorf("commit outreach sequence update: %w", err)
	}
	sequence.Steps, err = service.listSequenceSteps(ctx, sequence.ID)
	return sequence, err
}

func (service *Service) ApproveSequence(ctx context.Context, principal auth.Principal, sequenceID uuid.UUID, expectedUpdatedAt time.Time) (Sequence, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return Sequence{}, restaurants.ErrForbidden
	}
	if expectedUpdatedAt.IsZero() {
		return Sequence{}, fmt.Errorf("%w: expected_updated_at is required", ErrSequenceInvalid)
	}
	steps, err := service.listSequenceSteps(ctx, sequenceID)
	if err != nil {
		return Sequence{}, err
	}
	if err := validateSequenceSteps(steps); err != nil {
		return Sequence{}, err
	}
	tx, err := service.pool.Begin(ctx)
	if err != nil {
		return Sequence{}, fmt.Errorf("begin outreach sequence approval: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status string
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
		SELECT status, updated_at FROM outreach_email_sequences WHERE id = $1 FOR UPDATE`, sequenceID).Scan(&status, &currentUpdatedAt); errors.Is(err, pgx.ErrNoRows) {
		return Sequence{}, repository.ErrNotFound
	} else if err != nil {
		return Sequence{}, fmt.Errorf("lock outreach sequence for approval: %w", err)
	}
	if status != SequenceStatusDraft {
		return Sequence{}, fmt.Errorf("%w: only a draft version can be approved", ErrSequenceInvalid)
	}
	if !currentUpdatedAt.Equal(expectedUpdatedAt) {
		return Sequence{}, ErrSequenceStale
	}
	if _, err := tx.Exec(ctx, `
		UPDATE outreach_email_sequences
		SET status = 'archived', is_active = false, updated_at = now()
		WHERE is_active = true AND status = 'approved'`); err != nil {
		return Sequence{}, fmt.Errorf("archive prior outreach sequence: %w", err)
	}
	var sequence Sequence
	if err := tx.QueryRow(ctx, `
		UPDATE outreach_email_sequences
		SET status = 'approved', is_active = true, approved_at = now(), approved_by = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, name, version, status, is_active, approved_at, approved_by, created_at, updated_at`, sequenceID, principal.UserID).Scan(
		&sequence.ID, &sequence.Name, &sequence.Version, &sequence.Status,
		&sequence.IsActive, &sequence.ApprovedAt, &sequence.ApprovedBy,
		&sequence.CreatedAt, &sequence.UpdatedAt,
	); err != nil {
		return Sequence{}, fmt.Errorf("approve outreach sequence: %w", err)
	}
	// Move only untouched enrollments onto the new approved version. Once any
	// provider attempt exists (including an ambiguous one), the campaign remains
	// pinned to its original immutable sequence for auditable follow-ups.
	if _, err := tx.Exec(ctx, rebaseUntouchedEnrollmentsQuery, sequenceID); err != nil {
		return Sequence{}, fmt.Errorf("rebase untouched outreach enrollments: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT ensure_outreach_sequence_enrollment(id) FROM restaurants`); err != nil {
		return Sequence{}, fmt.Errorf("enroll eligible restaurants: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Sequence{}, fmt.Errorf("commit outreach sequence approval: %w", err)
	}
	sequence.Steps = steps
	return sequence, nil
}

func (service *Service) PreviewSequence(ctx context.Context, principal auth.Principal, sequenceID uuid.UUID, input PreviewSequenceInput) (SequencePreview, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return SequencePreview{}, restaurants.ErrForbidden
	}
	steps, err := service.listSequenceSteps(ctx, sequenceID)
	if err != nil {
		return SequencePreview{}, err
	}
	if err := validateSequenceSteps(steps); err != nil {
		return SequencePreview{}, err
	}
	name := cleanSingleLine(input.RestaurantName)
	if name == "" {
		name = "Example Restaurant"
	}
	greeting := outreachGreeting(input.OwnerFirstName, name)
	preview := SequencePreview{RestaurantName: name, Greeting: greeting, Steps: []RenderedSequenceStep{}}
	for _, step := range steps {
		if !step.Enabled {
			continue
		}
		rendered, err := renderSequenceStep(step, name, input.OwnerFirstName, "https://api.tuvisolutions.com/t/unsubscribe/preview-token")
		if err != nil {
			return SequencePreview{}, err
		}
		preview.Steps = append(preview.Steps, rendered)
	}
	return preview, nil
}

func (service *Service) ListRecipientProgress(ctx context.Context, principal auth.Principal, limit, offset int) (RecipientProgressList, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return RecipientProgressList{}, restaurants.ErrForbidden
	}
	if service.pool == nil {
		return RecipientProgressList{}, fmt.Errorf("database pool is not configured")
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := service.pool.QueryRow(ctx, `SELECT count(*) FROM restaurants`).Scan(&total); err != nil {
		return RecipientProgressList{}, fmt.Errorf("count outreach recipients: %w", err)
	}
	rows, err := service.pool.Query(ctx, `
		SELECT r.id,
		       r.name,
		       r.email,
		       r.status,
		       r.outreach_consent_basis,
		       COALESCE(c.current_step, 0),
		       c.next_step,
		       c.next_send_at,
		       r.last_email_sent_at,
		       c.completed_at,
		       r.email_send_count,
		       COALESCE(c.status, ''),
		       EXISTS (SELECT 1 FROM email_suppressions s WHERE s.email = lower(trim(r.email))),
		       CASE
		         WHEN trim(r.name) = '' THEN 'missing_name'
		         WHEN lower(trim(r.email)) !~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$' THEN 'invalid_email'
		         WHEN r.status NOT IN ('lead', 'emailed') OR r.shown_interest THEN 'lifecycle_paused'
		         WHEN r.outreach_consent_basis <> 'inferred_business'
		           OR r.outreach_consent_recorded_at IS NULL
		           OR trim(r.outreach_consent_source) = ''
		           OR jsonb_typeof(r.outreach_consent_evidence) <> 'object'
		           OR r.outreach_consent_evidence = '{}'::jsonb
		         THEN 'consent_evidence_missing'
		         WHEN EXISTS (SELECT 1 FROM email_suppressions s WHERE s.email = lower(trim(r.email))) THEN 'suppressed'
		         WHEN c.id IS NULL THEN 'not_enrolled'
		         WHEN c.status = 'send_unknown' THEN 'delivery_unknown'
		         WHEN c.completed_at IS NOT NULL THEN 'complete'
		         WHEN c.next_send_at > now() THEN 'scheduled'
		         ELSE ''
		       END
		FROM restaurants r
		LEFT JOIN email_campaigns c
		  ON c.restaurant_id = r.id
		 AND c.campaign_type = 'outreach'
		 AND c.sequence_id IS NOT NULL
		ORDER BY c.current_step DESC NULLS LAST, c.next_send_at NULLS LAST, r.created_at
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return RecipientProgressList{}, fmt.Errorf("list outreach recipients: %w", err)
	}
	defer rows.Close()
	result := RecipientProgressList{Recipients: []RecipientProgress{}, Total: total}
	for rows.Next() {
		var record RecipientProgress
		if err := rows.Scan(
			&record.RestaurantID, &record.RestaurantName, &record.Email,
			&record.LifecycleStatus, &record.ConsentBasis, &record.CurrentStep,
			&record.NextStep, &record.NextSendAt, &record.LastSentAt,
			&record.CompletedAt, &record.EmailSendCount, &record.CampaignStatus,
			&record.Suppressed, &record.HoldReason,
		); err != nil {
			return RecipientProgressList{}, fmt.Errorf("scan outreach recipient: %w", err)
		}
		record.Eligible = record.HoldReason == "" || record.HoldReason == "scheduled"
		result.Recipients = append(result.Recipients, record)
	}
	if err := rows.Err(); err != nil {
		return RecipientProgressList{}, fmt.Errorf("iterate outreach recipients: %w", err)
	}
	return result, nil
}

func (service *Service) listSequenceSteps(ctx context.Context, sequenceID uuid.UUID) ([]SequenceStep, error) {
	if service.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	rows, err := service.pool.Query(ctx, `
		SELECT id, sequence_id, position, enabled, delay_hours,
		       subject_template, body_text_template, created_at, updated_at
		FROM outreach_email_sequence_steps
		WHERE sequence_id = $1
		ORDER BY position`, sequenceID)
	if err != nil {
		return nil, fmt.Errorf("list outreach sequence steps: %w", err)
	}
	defer rows.Close()
	steps := []SequenceStep{}
	for rows.Next() {
		var step SequenceStep
		if err := rows.Scan(
			&step.ID, &step.SequenceID, &step.Position, &step.Enabled,
			&step.DelayHours, &step.SubjectTemplate, &step.BodyTextTemplate,
			&step.CreatedAt, &step.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan outreach sequence step: %w", err)
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outreach sequence steps: %w", err)
	}
	if len(steps) == 0 {
		var exists bool
		if err := service.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM outreach_email_sequences WHERE id = $1)`, sequenceID).Scan(&exists); err != nil {
			return nil, fmt.Errorf("check outreach sequence: %w", err)
		}
		if !exists {
			return nil, repository.ErrNotFound
		}
	}
	return steps, nil
}

func (service *Service) activeSequenceSteps(ctx context.Context) ([]SequenceStep, error) {
	if repo, ok := service.repo.(activeSequenceStepsRepository); ok {
		return repo.ListActiveSequenceSteps(ctx)
	}
	if service.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	rows, err := service.pool.Query(ctx, `
		SELECT step.id, step.sequence_id, step.position, step.enabled, step.delay_hours,
		       step.subject_template, step.body_text_template, step.created_at, step.updated_at
		FROM outreach_email_sequences seq
		JOIN outreach_email_sequence_steps step ON step.sequence_id = seq.id
		WHERE seq.is_active = true
		  AND seq.status = 'approved'
		  AND step.enabled = true
		ORDER BY step.position`)
	if err != nil {
		return nil, fmt.Errorf("list active outreach sequence steps: %w", err)
	}
	defer rows.Close()
	steps := []SequenceStep{}
	for rows.Next() {
		var step SequenceStep
		if err := rows.Scan(
			&step.ID, &step.SequenceID, &step.Position, &step.Enabled,
			&step.DelayHours, &step.SubjectTemplate, &step.BodyTextTemplate,
			&step.CreatedAt, &step.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active outreach sequence step: %w", err)
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active outreach sequence steps: %w", err)
	}
	return steps, nil
}

func validateSequenceSteps(steps []SequenceStep) error {
	if len(steps) == 0 {
		return fmt.Errorf("%w: at least one step is required", ErrSequenceInvalid)
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Position < steps[j].Position })
	enabled := 0
	firstEnabledChecked := false
	for index, step := range steps {
		if step.Position != index+1 {
			return fmt.Errorf("%w: step positions must be contiguous and 1-based", ErrSequenceInvalid)
		}
		if step.DelayHours < 0 || step.DelayHours > 8760 {
			return fmt.Errorf("%w: step %d delay must be between 0 and 8760 hours", ErrSequenceInvalid, step.Position)
		}
		if step.Enabled {
			enabled++
			if !firstEnabledChecked {
				firstEnabledChecked = true
				if step.DelayHours != 0 {
					return fmt.Errorf("%w: the first enabled step must have a zero-hour delay", ErrSequenceInvalid)
				}
			}
		}
		if err := validateSequenceTemplate(step); err != nil {
			return err
		}
	}
	if enabled == 0 {
		return fmt.Errorf("%w: at least one step must be enabled", ErrSequenceInvalid)
	}
	return nil
}

func validateSequenceTemplate(step SequenceStep) error {
	subject := strings.TrimSpace(step.SubjectTemplate)
	body := strings.TrimSpace(step.BodyTextTemplate)
	if subject == "" || len(subject) > 200 || strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("%w: step %d subject must be one line and at most 200 characters", ErrSequenceInvalid, step.Position)
	}
	if body == "" || len(body) > 10000 || htmlElementPattern.MatchString(body) {
		return fmt.Errorf("%w: step %d body must be plain text", ErrSequenceInvalid, step.Position)
	}
	allowed := map[string]bool{
		"{{greeting}}": true, "{{restaurant_name}}": true,
		"{{website_url}}": true, "{{unsubscribe_url}}": true,
	}
	for _, value := range append(templatePlaceholderPattern.FindAllString(subject, -1), templatePlaceholderPattern.FindAllString(body, -1)...) {
		if !allowed[value] {
			return fmt.Errorf("%w: step %d contains unsupported placeholder %s", ErrSequenceInvalid, step.Position, value)
		}
	}
	return nil
}

func containsUnmanagedLinkCandidate(value string) bool {
	return rawURLPattern.MatchString(value) || bareDomainPattern.MatchString(value)
}

func renderSequenceStep(step SequenceStep, restaurantName, ownerFirstName, unsubscribeURL string) (RenderedSequenceStep, error) {
	name := cleanSingleLine(restaurantName)
	if name == "" {
		return RenderedSequenceStep{}, fmt.Errorf("%w: restaurant name is required", campaigns.ErrNotEligible)
	}
	greeting := outreachGreeting(ownerFirstName, name)
	replacer := strings.NewReplacer(
		"{{greeting}}", greeting,
		"{{restaurant_name}}", name,
		"{{website_url}}", websiteURL,
		"{{unsubscribe_url}}", strings.TrimSpace(unsubscribeURL),
	)
	subject := replacer.Replace(strings.TrimSpace(step.SubjectTemplate))
	body := replacer.Replace(strings.TrimSpace(step.BodyTextTemplate))
	if templatePlaceholderPattern.MatchString(subject + "\n" + body) {
		return RenderedSequenceStep{}, fmt.Errorf("%w: rendered step contains an unresolved placeholder", ErrSequenceInvalid)
	}
	body = ensureSequenceOptOut(body, strings.TrimSpace(unsubscribeURL))
	urls := rawURLPattern.FindAllString(body, -1)
	return RenderedSequenceStep{
		Position: step.Position, DelayHours: step.DelayHours,
		Subject: subject, BodyText: body, URLCount: len(urls),
	}, nil
}

func ensureSequenceOptOut(body, unsubscribeURL string) string {
	body = strings.TrimSpace(body)
	unsubscribeURL = strings.TrimSpace(unsubscribeURL)
	if unsubscribeURL == "" || strings.Contains(body, unsubscribeURL) {
		return body
	}
	if body == "" {
		return "Unsubscribe: " + unsubscribeURL
	}
	return body + "\n\nUnsubscribe: " + unsubscribeURL
}

func checkSequenceDeliveryEligibility(delivery SequenceDelivery, suppressed bool) error {
	if strings.TrimSpace(delivery.RestaurantName) == "" {
		return fmt.Errorf("%w: restaurant has no name", campaigns.ErrNotEligible)
	}
	if !validLeadEmailPattern.MatchString(strings.ToLower(strings.TrimSpace(delivery.RecipientEmail))) {
		return fmt.Errorf("%w: restaurant has no valid contact email", campaigns.ErrNotEligible)
	}
	if suppressed {
		return fmt.Errorf("%w: recipient is suppressed", campaigns.ErrNotEligible)
	}
	if delivery.ShownInterest || delivery.LifecycleStatus == restaurants.StatusInterested {
		return fmt.Errorf("%w: restaurant has expressed interest and automated outreach is paused", campaigns.ErrNotEligible)
	}
	switch delivery.LifecycleStatus {
	case restaurants.StatusLead, restaurants.StatusEmailed:
	default:
		return fmt.Errorf("%w: restaurant lifecycle is not eligible for outreach", campaigns.ErrNotEligible)
	}
	if delivery.ConsentBasis != "inferred_business" ||
		delivery.ConsentRecordedAt == nil ||
		strings.TrimSpace(delivery.ConsentSource) == "" ||
		!validConsentEvidence(delivery.ConsentEvidence) {
		return fmt.Errorf("%w: inferred-business consent evidence is missing", campaigns.ErrNotEligible)
	}
	if delivery.SequenceStatus != SequenceStatusApproved && delivery.SequenceStatus != SequenceStatusArchived {
		return fmt.Errorf("%w: sequence version was not approved", campaigns.ErrNotEligible)
	}
	if !delivery.Step.Enabled {
		return fmt.Errorf("%w: sequence step is disabled", campaigns.ErrNotEligible)
	}
	return nil
}

func validConsentEvidence(raw json.RawMessage) bool {
	var evidence map[string]any
	return len(raw) > 0 && json.Unmarshal(raw, &evidence) == nil && len(evidence) > 0
}

func outreachGreeting(ownerFirstName, restaurantName string) string {
	if first := cleanFirstName(ownerFirstName); first != "" {
		return "Hi " + first + ","
	}
	name := cleanSingleLine(restaurantName)
	if name == "" {
		name = "restaurant"
	}
	return "Hi " + name + " team,"
}

func cleanFirstName(value string) string {
	value = cleanSingleLine(value)
	if value == "" {
		return ""
	}
	value = strings.Fields(value)[0]
	var builder strings.Builder
	for _, char := range value {
		if unicode.IsLetter(char) || char == '-' || char == '\'' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func cleanSingleLine(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
