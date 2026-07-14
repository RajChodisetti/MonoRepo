package outreach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	platformdb "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/db"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

const emailDeliveryLease = 5 * time.Minute

var errDeliveryAttemptNotSending = errors.New("delivery attempt is not in sending state")

func (repo *Postgres) SyncEmailAccounts(
	ctx context.Context,
	accounts []emailprovider.QuotaAccountConfig,
	cooldown time.Duration,
) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	if len(accounts) == 0 {
		return fmt.Errorf("at least one outreach email account is required")
	}
	if cooldown < 24*time.Hour {
		return fmt.Errorf("outreach email cooldown must be at least 24h")
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email account sync: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, account := range accounts {
		key := strings.TrimSpace(account.Key)
		if key == "" {
			return fmt.Errorf("outreach email account key is required")
		}
		if account.SendLimit < 1 || account.SendLimit > 40 {
			return fmt.Errorf("outreach email account %q send limit must be between 1 and 40", key)
		}
		provider := strings.TrimSpace(account.Provider)
		providerIdentity := strings.ToLower(strings.TrimSpace(account.ProviderIdentity))
		fromEmail := strings.TrimSpace(account.FromEmail)
		if provider == "" || providerIdentity == "" || fromEmail == "" || account.Position < 0 {
			return fmt.Errorf("outreach email account %q has invalid provider metadata", key)
		}
		const upsert = `
			INSERT INTO outreach_email_accounts (
				account_key, provider, provider_identity, from_email, position, enabled, send_limit
			) VALUES ($1, $2, $3, $4, $5, true, $6)
			ON CONFLICT (provider_identity) DO UPDATE
			SET account_key = EXCLUDED.account_key,
			    provider = EXCLUDED.provider,
			    from_email = EXCLUDED.from_email,
			    position = EXCLUDED.position,
			    enabled = true,
			    send_limit = EXCLUDED.send_limit,
			    usage_count = LEAST(outreach_email_accounts.usage_count, EXCLUDED.send_limit),
			    available_at = CASE
			      WHEN outreach_email_accounts.send_limit > EXCLUDED.send_limit
			       AND outreach_email_accounts.usage_count >= EXCLUDED.send_limit
			       AND outreach_email_accounts.available_at <= now()
			      THEN now() + ($7 * interval '1 second')
			      ELSE outreach_email_accounts.available_at
			    END,
			    updated_at = now()`
		if _, err := tx.Exec(
			ctx,
			upsert,
			key,
			provider,
			providerIdentity,
			fromEmail,
			account.Position,
			account.SendLimit,
			int64(cooldown/time.Second),
		); err != nil {
			return fmt.Errorf("sync outreach email account %q: %w", key, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email account sync: %w", err)
	}
	return nil
}

func (repo *Postgres) ClaimEmailDelivery(
	ctx context.Context,
	accountKeys []string,
	delivery emailprovider.DeliveryContext,
	cooldown time.Duration,
) (emailprovider.DeliveryClaim, error) {
	if repo.pool == nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("database pool is not configured")
	}
	if len(accountKeys) == 0 {
		return emailprovider.DeliveryClaim{}, emailprovider.ErrAccountsExhausted
	}
	if delivery.CampaignID == uuid.Nil || delivery.RestaurantID == uuid.Nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("campaign and restaurant are required for quota-managed delivery")
	}
	if delivery.Step < 0 {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("campaign step cannot be negative")
	}
	if cooldown < 24*time.Hour {
		cooldown = 24 * time.Hour
	}
	if _, err := repo.ReconcileStaleEmailDeliveries(ctx); err != nil {
		return emailprovider.DeliveryClaim{}, err
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("begin email quota claim: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, delivery.RestaurantID); err != nil {
		return emailprovider.DeliveryClaim{}, err
	}

	const lockCampaign = `
		SELECT campaign.restaurant_id,
		       campaign.status,
		       campaign.campaign_type,
		       campaign.subject,
		       campaign.body_html,
		       campaign.body_text,
		       campaign.demo_token,
		       restaurant.email,
		       restaurant.email_sent,
		       restaurant.email_send_count,
		       profile.ocr_status,
		       profile.review_status,
		       (profile.reviewed_at IS NOT NULL
		        AND profile.reviewed_by IS NOT NULL
		        AND profile.reviewed_at >= profile.updated_at
		        AND profile.reviewed_at >= restaurant.updated_at),
		       demo.status,
		       demo.token_hash,
		       (demo.published_at IS NOT NULL AND demo.published_by IS NOT NULL),
		       demo.expires_at,
		       (campaign.approved_at IS NOT NULL AND campaign.approved_by IS NOT NULL),
		       (
		         NOT campaign.auto_generated
		         OR (
		           campaign.source_ocr_fingerprint <> ''
		           AND campaign.source_ocr_fingerprint = profile.ocr_input_fingerprint
		           AND campaign.source_profile_fingerprint <> ''
		           AND campaign.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(restaurant.id), '')
		         )
		       ),
		       (
		         NOT demo.auto_generated
		         OR (
		           demo.source_ocr_fingerprint <> ''
		           AND demo.source_ocr_fingerprint = profile.ocr_input_fingerprint
		           AND demo.source_profile_fingerprint <> ''
		           AND demo.source_profile_fingerprint = COALESCE(lead_artifact_current_profile_fingerprint(restaurant.id), '')
		         )
		       )
		FROM email_campaigns AS campaign
		JOIN restaurants AS restaurant ON restaurant.id = campaign.restaurant_id
		JOIN restaurant_profiles AS profile ON profile.restaurant_id = campaign.restaurant_id
		JOIN demo_sites AS demo ON demo.id = campaign.demo_site_id
		WHERE campaign.id = $1
		FOR UPDATE OF campaign, restaurant, profile, demo`
	var restaurantID uuid.UUID
	var campaignStatus string
	var campaignType string
	var campaignSubject string
	var campaignBodyHTML string
	var campaignBodyText string
	var campaignDemoToken string
	var restaurantEmail string
	var restaurantEmailSent bool
	var restaurantEmailSendCount int
	var ocrStatus string
	var reviewStatus string
	var profileReviewAudited bool
	var demoStatus string
	var demoTokenHash string
	var demoPublishAudited bool
	var demoExpiresAt *time.Time
	var campaignApprovalAudited bool
	var campaignProvenanceCurrent bool
	var demoProvenanceCurrent bool
	if err := tx.QueryRow(ctx, lockCampaign, delivery.CampaignID).Scan(
		&restaurantID,
		&campaignStatus,
		&campaignType,
		&campaignSubject,
		&campaignBodyHTML,
		&campaignBodyText,
		&campaignDemoToken,
		&restaurantEmail,
		&restaurantEmailSent,
		&restaurantEmailSendCount,
		&ocrStatus,
		&reviewStatus,
		&profileReviewAudited,
		&demoStatus,
		&demoTokenHash,
		&demoPublishAudited,
		&demoExpiresAt,
		&campaignApprovalAudited,
		&campaignProvenanceCurrent,
		&demoProvenanceCurrent,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: campaign review context is unavailable", campaigns.ErrNotEligible)
		}
		return emailprovider.DeliveryClaim{}, fmt.Errorf("lock outreach campaign: %w", err)
	}
	if restaurantID != delivery.RestaurantID {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("campaign does not belong to restaurant")
	}
	if campaignType != campaigns.TypeOutreach {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("quota-managed delivery requires an outreach campaign")
	}
	if campaignStatus != campaigns.StatusApproved {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: campaign must remain approved while claiming email quota", campaigns.ErrNotEligible)
	}
	if !campaignProvenanceCurrent || !demoProvenanceCurrent {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: automatic draft provenance changed after approval", campaigns.ErrNotEligible)
	}
	actualRecipient := strings.ToLower(strings.TrimSpace(restaurantEmail))
	expectedRecipient := strings.ToLower(strings.TrimSpace(delivery.Recipient))
	if expectedRecipient == "" || expectedRecipient != actualRecipient {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: outreach recipient changed before quota claim", campaigns.ErrNotEligible)
	}
	expectedArtifact := strings.TrimSpace(delivery.CampaignArtifactFingerprint)
	currentArtifact := emailprovider.CampaignArtifactFingerprint(
		campaignSubject,
		campaignBodyHTML,
		campaignBodyText,
		campaignDemoToken,
	)
	if expectedArtifact == "" || expectedArtifact != currentArtifact {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: campaign content changed before quota claim", campaigns.ErrNotEligible)
	}
	if strings.TrimSpace(campaignDemoToken) == "" || demos.CheckDemoToken(demoTokenHash, campaignDemoToken) != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: campaign demo token is no longer valid", campaigns.ErrNotEligible)
	}
	demoExpired := demoExpiresAt != nil && !demoExpiresAt.After(time.Now().UTC())
	if restaurantEmailSent || restaurantEmailSendCount != 0 {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: restaurant already has a confirmed outreach send", campaigns.ErrNotEligible)
	}

	// Serialize the final suppression check with unsubscribe writes. The row may
	// not exist yet, so a row lock alone cannot prevent an insertion phantom.
	if _, err := tx.Exec(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtextextended(lower(trim($1)), 0))`,
		restaurantEmail,
	); err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("lock outreach recipient suppression key: %w", err)
	}
	var suppressed bool
	if err := tx.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM email_suppressions WHERE email = lower(trim($1)))`,
		restaurantEmail,
	).Scan(&suppressed); err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("recheck outreach suppression: %w", err)
	}
	if err := campaigns.CheckEligibility(campaigns.EligibilityInput{
		RestaurantEmail:         restaurantEmail,
		OCRStatus:               ocrStatus,
		ReviewStatus:            reviewStatus,
		ProfileReviewAudited:    profileReviewAudited,
		DemoStatus:              demoStatus,
		DemoPublishAudited:      demoPublishAudited,
		DemoExpired:             demoExpired,
		CampaignStatus:          campaignStatus,
		CampaignApprovalAudited: campaignApprovalAudited,
		Suppressed:              suppressed,
	}); err != nil {
		return emailprovider.DeliveryClaim{}, err
	}

	const selectAccount = `
		SELECT id, account_key, cycle_number, usage_count, send_limit
		FROM outreach_email_accounts
		WHERE enabled = true
		  AND account_key = ANY($1::text[])
		  AND (usage_count < send_limit OR available_at <= now())
		ORDER BY CASE WHEN usage_count < send_limit THEN 0 ELSE 1 END,
		         position ASC,
		         created_at ASC
		FOR UPDATE
		LIMIT 1`
	var accountID uuid.UUID
	var accountKey string
	var cycleNumber int64
	var usageCount int
	var sendLimit int
	if err := tx.QueryRow(ctx, selectAccount, accountKeys).Scan(
		&accountID,
		&accountKey,
		&cycleNumber,
		&usageCount,
		&sendLimit,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return emailprovider.DeliveryClaim{}, emailprovider.ErrAccountsExhausted
		}
		return emailprovider.DeliveryClaim{}, fmt.Errorf("select available outreach email account: %w", err)
	}

	resetCycle := usageCount >= sendLimit
	if resetCycle {
		cycleNumber++
		usageCount = 0
	}
	usageCount++
	reachesLimit := usageCount >= sendLimit

	const consumeSlot = `
		UPDATE outreach_email_accounts
		SET cycle_number = $2,
		    usage_count = $3,
		    cycle_started_at = CASE WHEN $4 OR cycle_started_at IS NULL THEN now() ELSE cycle_started_at END,
		    available_at = CASE
		      WHEN $5 THEN now() + ($6 * interval '1 second')
		      ELSE now()
		    END,
		    last_used_at = now(),
		    updated_at = now()
		WHERE id = $1`
	if _, err := tx.Exec(
		ctx,
		consumeSlot,
		accountID,
		cycleNumber,
		usageCount,
		resetCycle,
		reachesLimit,
		int64(cooldown/time.Second),
	); err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("consume outreach email quota: %w", err)
	}

	const markSending = `
		UPDATE email_campaigns
		SET status = $2, current_step = $3, updated_at = now()
		WHERE id = $1 AND status = $4`
	result, err := tx.Exec(
		ctx,
		markSending,
		delivery.CampaignID,
		campaigns.StatusSending,
		delivery.Step,
		campaigns.StatusApproved,
	)
	if err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("claim approved outreach campaign: %w", err)
	}
	if result.RowsAffected() != 1 {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("%w: campaign was claimed by another sender", campaigns.ErrNotEligible)
	}

	const insertAttempt = `
		INSERT INTO email_delivery_attempts (
			campaign_id, restaurant_id, account_id, bulk_job_id,
			campaign_step, account_cycle, account_sequence, recipient_email,
			status, lease_expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'sending', now() + ($9 * interval '1 second'))
		RETURNING id, send_sequence`
	var claim emailprovider.DeliveryClaim
	if err := tx.QueryRow(
		ctx,
		insertAttempt,
		delivery.CampaignID,
		delivery.RestaurantID,
		accountID,
		delivery.BulkJobID,
		delivery.Step,
		cycleNumber,
		usageCount,
		actualRecipient,
		int64(emailDeliveryLease/time.Second),
	).Scan(&claim.AttemptID, &claim.SendSequence); err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("create outreach email delivery attempt: %w", err)
	}
	claim.AccountKey = accountKey
	claim.CampaignStep = delivery.Step
	claim.AccountCycle = cycleNumber
	claim.AccountSequence = usageCount

	if err := tx.Commit(ctx); err != nil {
		return emailprovider.DeliveryClaim{}, fmt.Errorf("commit email quota claim: %w", err)
	}
	return claim, nil
}

// ReconcileStaleEmailDeliveries fails closed when a worker disappears after
// claiming quota but before recording the provider outcome. The slot remains
// consumed and the campaign requires an operator decision; automatic retry
// could otherwise duplicate a message accepted by the provider.
func (repo *Postgres) ReconcileStaleEmailDeliveries(ctx context.Context) (int, error) {
	if repo.pool == nil {
		return 0, fmt.Errorf("database pool is not configured")
	}

	rows, err := repo.pool.Query(ctx, `
		SELECT id
		FROM email_delivery_attempts
		WHERE status = 'sending' AND lease_expires_at <= now()
		ORDER BY lease_expires_at, id
		LIMIT 100`)
	if err != nil {
		return 0, fmt.Errorf("list stale outreach email deliveries: %w", err)
	}
	var attemptIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan stale outreach email delivery: %w", err)
		}
		attemptIDs = append(attemptIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate stale outreach email deliveries: %w", err)
	}
	rows.Close()

	count := 0
	for _, attemptID := range attemptIDs {
		err := repo.MarkEmailDeliveryUnknown(
			ctx,
			emailprovider.DeliveryClaim{AttemptID: attemptID},
			"delivery_lease_expired",
		)
		if errors.Is(err, errDeliveryAttemptNotSending) {
			continue
		}
		if err != nil {
			return count, fmt.Errorf("reconcile stale outreach email delivery %s: %w", attemptID, err)
		}
		count++
	}
	return count, nil
}

func (repo *Postgres) NextEmailAccountAvailableAt(
	ctx context.Context,
	accountKeys []string,
) (*time.Time, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}
	if len(accountKeys) == 0 {
		return nil, nil
	}

	const query = `
		SELECT count(*), CASE
		  WHEN count(*) FILTER (
		    WHERE usage_count < send_limit OR available_at <= now()
		  ) > 0 THEN NULL
		  ELSE min(available_at)
		END
		FROM outreach_email_accounts
		WHERE enabled = true AND account_key = ANY($1::text[])`
	var accountCount int
	var next *time.Time
	if err := repo.pool.QueryRow(ctx, query, accountKeys).Scan(&accountCount, &next); err != nil {
		return nil, fmt.Errorf("get next outreach email account availability: %w", err)
	}
	if accountCount == 0 {
		return nil, fmt.Errorf("no configured outreach email accounts are enabled")
	}
	return next, nil
}

func (repo *Postgres) CompleteEmailDelivery(
	ctx context.Context,
	claim emailprovider.DeliveryClaim,
	providerMessageID string,
) error {
	return repo.finalizeEmailDelivery(ctx, claim, deliveryFinalization{
		status:            "sent",
		eventType:         campaigns.EventSent,
		providerMessageID: strings.TrimSpace(providerMessageID),
	})
}

func (repo *Postgres) SkipEmailDelivery(
	ctx context.Context,
	claim emailprovider.DeliveryClaim,
	skipped bool,
	redirected bool,
) error {
	return repo.finalizeEmailDelivery(ctx, claim, deliveryFinalization{
		status:     "skipped",
		eventType:  campaigns.EventSkipped,
		skipped:    skipped,
		redirected: redirected,
	})
}

func (repo *Postgres) MarkEmailDeliveryUnknown(
	ctx context.Context,
	claim emailprovider.DeliveryClaim,
	errorCode string,
) error {
	code := strings.TrimSpace(errorCode)
	if code == "" {
		code = "provider_send_unknown"
	}
	return repo.finalizeEmailDelivery(ctx, claim, deliveryFinalization{
		status:    "unknown",
		eventType: campaigns.EventFailed,
		errorCode: code,
	})
}

type deliveryFinalization struct {
	status            string
	eventType         string
	providerMessageID string
	errorCode         string
	skipped           bool
	redirected        bool
}

func (repo *Postgres) finalizeEmailDelivery(
	ctx context.Context,
	claim emailprovider.DeliveryClaim,
	finalization deliveryFinalization,
) error {
	if repo.pool == nil {
		return fmt.Errorf("database pool is not configured")
	}
	if claim.AttemptID == uuid.Nil {
		return fmt.Errorf("delivery attempt is required")
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email delivery finalization: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var lockedRestaurantID uuid.UUID
	if err := tx.QueryRow(
		ctx,
		`SELECT restaurant_id FROM email_delivery_attempts WHERE id = $1`,
		claim.AttemptID,
	).Scan(&lockedRestaurantID); errors.Is(err, pgx.ErrNoRows) {
		return errDeliveryAttemptNotSending
	} else if err != nil {
		return fmt.Errorf("load delivery attempt restaurant: %w", err)
	}
	if err := platformdb.LockRestaurantWorkflow(ctx, tx, lockedRestaurantID); err != nil {
		return err
	}

	const updateAttempt = `
		UPDATE email_delivery_attempts
		SET status = $2,
		    provider_message_id = $3,
		    error_code = $4,
		    sent_at = CASE WHEN $2 = 'sent' THEN now() ELSE sent_at END,
		    lease_expires_at = NULL,
		    updated_at = now()
		WHERE id = $1
		  AND status = 'sending'
		  AND ($2 = 'unknown' OR lease_expires_at > now())
		RETURNING campaign_id, restaurant_id, send_sequence, account_cycle, account_sequence,
		          recipient_email`
	var campaignID uuid.UUID
	var restaurantID uuid.UUID
	var sendSequence int64
	var accountCycle int64
	var accountSequence int
	var recipientEmail string
	err = tx.QueryRow(
		ctx,
		updateAttempt,
		claim.AttemptID,
		finalization.status,
		finalization.providerMessageID,
		finalization.errorCode,
	).Scan(&campaignID, &restaurantID, &sendSequence, &accountCycle, &accountSequence, &recipientEmail)
	if errors.Is(err, pgx.ErrNoRows) {
		var existingStatus string
		if statusErr := tx.QueryRow(ctx, `SELECT status FROM email_delivery_attempts WHERE id = $1`, claim.AttemptID).Scan(&existingStatus); statusErr == nil && existingStatus == finalization.status {
			return nil
		}
		return errDeliveryAttemptNotSending
	}
	if err != nil {
		return fmt.Errorf("update email delivery attempt: %w", err)
	}

	metadata, err := json.Marshal(map[string]any{
		"step":                claim.CampaignStep,
		"provider_message":    finalization.providerMessageID,
		"bulk_outreach":       true,
		"account_key":         claim.AccountKey,
		"account_cycle":       accountCycle,
		"account_sequence":    accountSequence,
		"send_sequence":       sendSequence,
		"skipped":             finalization.skipped,
		"redirected":          finalization.redirected,
		"error_code":          finalization.errorCode,
		"delivery_attempt_id": claim.AttemptID,
	})
	if err != nil {
		return fmt.Errorf("encode email delivery event: %w", err)
	}

	const insertEvent = `
		INSERT INTO email_events (
			campaign_id, restaurant_id, event_type, metadata, delivery_attempt_id
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (delivery_attempt_id, event_type)
		WHERE delivery_attempt_id IS NOT NULL
		DO NOTHING`
	if _, err := tx.Exec(ctx, insertEvent, campaignID, restaurantID, finalization.eventType, metadata, claim.AttemptID); err != nil {
		return fmt.Errorf("insert email delivery event: %w", err)
	}

	campaignStatus := campaigns.StatusApproved
	if finalization.status == "sent" {
		campaignStatus = campaigns.StatusSent
	} else if finalization.status == "unknown" {
		campaignStatus = campaigns.StatusSendUnknown
	}
	const updateCampaign = `
		UPDATE email_campaigns
		SET status = CASE
		      WHEN $2 = $7 THEN $7
		      WHEN status = $5 THEN status
		      ELSE $2
		    END,
		    last_sent_at = CASE WHEN $2 = $3 THEN now() ELSE last_sent_at END,
		    updated_at = now()
		WHERE id = $1
		  AND (
		    ($2 = $7 AND status <> $3)
		    OR ($2 <> $7 AND status IN ($4, $6, $5))
		  )`
	result, err := tx.Exec(
		ctx,
		updateCampaign,
		campaignID,
		campaignStatus,
		campaigns.StatusSent,
		campaigns.StatusSending,
		campaigns.StatusStopped,
		campaigns.StatusApproved,
		campaigns.StatusSendUnknown,
	)
	if err != nil {
		return fmt.Errorf("finalize outreach campaign: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("outreach campaign is not in sending state")
	}

	if finalization.status == "sent" {
		const updateRestaurant = `
			UPDATE restaurants
			SET is_contacted = true,
			    email_sent = true,
			    email_send_count = email_send_count + 1,
			    last_email_sent_at = now(),
			    last_email_send_sequence = $2,
			    last_email_recipient = $3,
			    status = CASE WHEN status IN ('lead', 'demo_ready') THEN 'emailed' ELSE status END,
			    updated_at = now()
			WHERE id = $1`
		result, err := tx.Exec(ctx, updateRestaurant, restaurantID, sendSequence, recipientEmail)
		if err != nil {
			return fmt.Errorf("record confirmed restaurant email: %w", err)
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("restaurant was not found while recording confirmed email")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit email delivery finalization: %w", err)
	}
	return nil
}

var _ emailprovider.QuotaStore = (*Postgres)(nil)
