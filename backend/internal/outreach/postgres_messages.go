package outreach

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const messageSelectColumns = `
	id, restaurant_id, campaign_id, delivery_attempt_id, reply_token,
	direction, from_email, to_email, reply_to, subject, body_text,
	gmail_message_id, gmail_thread_id, rfc_message_id, mailbox_key,
	unmatched, read_at, created_at`

func (repo *Postgres) InsertMessage(ctx context.Context, message Message) (Message, error) {
	if repo.pool == nil {
		return Message{}, fmt.Errorf("database pool is not configured")
	}
	direction := strings.TrimSpace(message.Direction)
	if direction != MessageDirectionOutbound && direction != MessageDirectionInbound {
		return Message{}, fmt.Errorf("email message direction is invalid")
	}
	const query = `
		INSERT INTO email_messages (
			restaurant_id, campaign_id, delivery_attempt_id, reply_token,
			direction, from_email, to_email, reply_to, subject, body_text,
			gmail_message_id, gmail_thread_id, rfc_message_id, mailbox_key,
			unmatched, read_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT DO NOTHING
		RETURNING` + messageSelectColumns
	record, err := scanMessage(repo.pool.QueryRow(
		ctx,
		query,
		message.RestaurantID,
		message.CampaignID,
		message.DeliveryAttemptID,
		message.ReplyToken,
		direction,
		strings.TrimSpace(message.FromEmail),
		strings.TrimSpace(message.ToEmail),
		strings.TrimSpace(message.ReplyTo),
		strings.TrimSpace(message.Subject),
		strings.TrimSpace(message.BodyText),
		strings.TrimSpace(message.GmailMessageID),
		strings.TrimSpace(message.GmailThreadID),
		strings.TrimSpace(message.RFCMessageID),
		strings.TrimSpace(message.MailboxKey),
		message.Unmatched,
		message.ReadAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		if strings.TrimSpace(message.GmailMessageID) != "" {
			existing, lookupErr := repo.GetMessageByGmailID(ctx, message.GmailMessageID)
			if lookupErr == nil {
				return existing, nil
			}
			if !errors.Is(lookupErr, pgx.ErrNoRows) {
				return Message{}, lookupErr
			}
		}
		if message.DeliveryAttemptID != nil {
			existing, lookupErr := scanMessage(repo.pool.QueryRow(
				ctx,
				`SELECT`+messageSelectColumns+` FROM email_messages WHERE delivery_attempt_id = $1 AND direction = $2`,
				message.DeliveryAttemptID,
				MessageDirectionOutbound,
			))
			if lookupErr == nil {
				return existing, nil
			}
			if !errors.Is(lookupErr, pgx.ErrNoRows) {
				return Message{}, lookupErr
			}
		}
		return Message{}, fmt.Errorf("insert email message conflicted without a stored row")
	}
	if err != nil {
		return Message{}, fmt.Errorf("insert email message: %w", err)
	}
	return record, nil
}

func (repo *Postgres) GetMessageByID(ctx context.Context, id uuid.UUID) (Message, error) {
	return scanMessage(repo.pool.QueryRow(ctx, `SELECT`+messageSelectColumns+` FROM email_messages WHERE id = $1`, id))
}

func (repo *Postgres) GetMessageByGmailID(ctx context.Context, gmailMessageID string) (Message, error) {
	gmailMessageID = strings.TrimSpace(gmailMessageID)
	if gmailMessageID == "" {
		return Message{}, pgx.ErrNoRows
	}
	return scanMessage(repo.pool.QueryRow(
		ctx,
		`SELECT`+messageSelectColumns+` FROM email_messages WHERE gmail_message_id = $1`,
		gmailMessageID,
	))
}

func (repo *Postgres) GetOutboundByReplyToken(ctx context.Context, token uuid.UUID) (Message, error) {
	if token == uuid.Nil {
		return Message{}, pgx.ErrNoRows
	}
	return scanMessage(repo.pool.QueryRow(
		ctx,
		`SELECT`+messageSelectColumns+`
		 FROM email_messages
		 WHERE direction = $1 AND reply_token = $2
		 ORDER BY created_at DESC
		 LIMIT 1`,
		MessageDirectionOutbound,
		token,
	))
}

func (repo *Postgres) GetOutboundByRFCMessageID(ctx context.Context, rfcMessageID string) (Message, error) {
	rfcMessageID = strings.TrimSpace(rfcMessageID)
	if rfcMessageID == "" {
		return Message{}, pgx.ErrNoRows
	}
	return scanMessage(repo.pool.QueryRow(
		ctx,
		`SELECT`+messageSelectColumns+`
		 FROM email_messages
		 WHERE direction = $1 AND lower(rfc_message_id) = lower($2)
		 ORDER BY created_at DESC
		 LIMIT 1`,
		MessageDirectionOutbound,
		rfcMessageID,
	))
}

func (repo *Postgres) GetOutboundByThreadID(ctx context.Context, threadID string) (Message, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return Message{}, pgx.ErrNoRows
	}
	return scanMessage(repo.pool.QueryRow(
		ctx,
		`SELECT`+messageSelectColumns+`
		 FROM email_messages
		 WHERE direction = $1 AND gmail_thread_id = $2
		 ORDER BY created_at DESC
		 LIMIT 1`,
		MessageDirectionOutbound,
		threadID,
	))
}

func (repo *Postgres) FindRestaurantIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return uuid.Nil, pgx.ErrNoRows
	}
	var id uuid.UUID
	err := repo.pool.QueryRow(
		ctx,
		`SELECT id
		 FROM restaurants
		 WHERE lower(trim(email)) = $1
		    OR lower(trim(coalesce(last_email_recipient, ''))) = $1
		 ORDER BY updated_at DESC
		 LIMIT 1`,
		email,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (repo *Postgres) ListInbox(ctx context.Context, unreadOnly bool, limit, offset int) (InboxList, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	unreadClause := ""
	if unreadOnly {
		unreadClause = `WHERE unread_count > 0`
	}
	countQuery := `
		SELECT count(*)
		FROM (
		  SELECT COALESCE(restaurant_id::text, id::text) AS thread_key,
		         count(*) FILTER (WHERE direction = 'inbound' AND read_at IS NULL) AS unread_count
		  FROM email_messages
		  GROUP BY COALESCE(restaurant_id::text, id::text)
		) threads ` + unreadClause
	var total int
	if err := repo.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return InboxList{}, fmt.Errorf("count inbox threads: %w", err)
	}

	query := `
		SELECT restaurant_id, restaurant_name, email, unmatched, unread_count,
		       last_direction, last_snippet, last_at, last_message_id
		FROM (
		  SELECT restaurant_id,
		         max(restaurant_name) AS restaurant_name,
		         max(email) AS email,
		         bool_or(unmatched AND restaurant_id IS NULL) AS unmatched,
		         count(*) FILTER (WHERE direction = 'inbound' AND read_at IS NULL)::int AS unread_count,
		         (array_agg(direction ORDER BY created_at DESC))[1] AS last_direction,
		         (array_agg(left(btrim(regexp_replace(body_text, '\s+', ' ', 'g')), 180) ORDER BY created_at DESC))[1] AS last_snippet,
		         max(created_at) AS last_at,
		         (array_agg(id ORDER BY created_at DESC))[1] AS last_message_id
		  FROM (
		    SELECT m.id, m.restaurant_id, m.direction, m.body_text, m.unmatched, m.read_at, m.created_at,
	           r.name AS restaurant_name,
	           COALESCE(
	             r.email,
	             CASE WHEN m.direction = 'inbound' THEN m.from_email ELSE m.to_email END
	           ) AS email,
		           COALESCE(m.restaurant_id::text, m.id::text) AS thread_key
		    FROM email_messages m
		    LEFT JOIN restaurants r ON r.id = m.restaurant_id
		  ) messages
		  GROUP BY thread_key, restaurant_id
		) threads ` + unreadClause + `
		ORDER BY unread_count DESC, last_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := repo.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return InboxList{}, fmt.Errorf("list inbox threads: %w", err)
	}
	defer rows.Close()

	threads := make([]InboxThread, 0)
	for rows.Next() {
		var thread InboxThread
		if err := rows.Scan(
			&thread.RestaurantID,
			&thread.RestaurantName,
			&thread.Email,
			&thread.Unmatched,
			&thread.UnreadCount,
			&thread.LastDirection,
			&thread.LastSnippet,
			&thread.LastAt,
			&thread.LastMessageID,
		); err != nil {
			return InboxList{}, fmt.Errorf("scan inbox thread: %w", err)
		}
		threads = append(threads, thread)
	}
	return InboxList{Threads: threads, Total: total}, rows.Err()
}

func (repo *Postgres) ListRestaurantMessages(ctx context.Context, restaurantID uuid.UUID) ([]Message, error) {
	rows, err := repo.pool.Query(
		ctx,
		`SELECT`+messageSelectColumns+`
		 FROM email_messages
		 WHERE restaurant_id = $1
		 ORDER BY created_at DESC`,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list restaurant email messages: %w", err)
	}
	defer rows.Close()
	messages := make([]Message, 0)
	for rows.Next() {
		record, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, record)
	}
	return messages, rows.Err()
}

func (repo *Postgres) MarkMessageRead(ctx context.Context, id uuid.UUID) (Message, error) {
	record, err := scanMessage(repo.pool.QueryRow(
		ctx,
		`UPDATE email_messages
		 SET read_at = COALESCE(read_at, now())
		 WHERE id = $1
		 RETURNING`+messageSelectColumns,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, err
	}
	if err != nil {
		return Message{}, fmt.Errorf("mark email message read: %w", err)
	}
	return record, nil
}

func (repo *Postgres) MarkRestaurantInboundRead(ctx context.Context, restaurantID uuid.UUID) error {
	_, err := repo.pool.Exec(
		ctx,
		`UPDATE email_messages
		 SET read_at = now()
		 WHERE restaurant_id = $1 AND direction = $2 AND read_at IS NULL`,
		restaurantID,
		MessageDirectionInbound,
	)
	if err != nil {
		return fmt.Errorf("mark restaurant inbound messages read: %w", err)
	}
	return nil
}

func (repo *Postgres) GetInboundSync(ctx context.Context, mailboxKey string) (string, error) {
	var historyID string
	err := repo.pool.QueryRow(
		ctx,
		`SELECT history_id FROM outreach_inbound_sync WHERE mailbox_key = $1`,
		strings.TrimSpace(mailboxKey),
	).Scan(&historyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get inbound sync cursor: %w", err)
	}
	return historyID, nil
}

func (repo *Postgres) SetInboundSync(ctx context.Context, mailboxKey, historyID string) error {
	_, err := repo.pool.Exec(
		ctx,
		`INSERT INTO outreach_inbound_sync (mailbox_key, history_id, updated_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (mailbox_key) DO UPDATE
		 SET history_id = EXCLUDED.history_id, updated_at = now()`,
		strings.TrimSpace(mailboxKey),
		strings.TrimSpace(historyID),
	)
	if err != nil {
		return fmt.Errorf("set inbound sync cursor: %w", err)
	}
	return nil
}

type messageScanner interface {
	Scan(dest ...any) error
}

func scanMessage(row messageScanner) (Message, error) {
	var record Message
	err := row.Scan(
		&record.ID,
		&record.RestaurantID,
		&record.CampaignID,
		&record.DeliveryAttemptID,
		&record.ReplyToken,
		&record.Direction,
		&record.FromEmail,
		&record.ToEmail,
		&record.ReplyTo,
		&record.Subject,
		&record.BodyText,
		&record.GmailMessageID,
		&record.GmailThreadID,
		&record.RFCMessageID,
		&record.MailboxKey,
		&record.Unmatched,
		&record.ReadAt,
		&record.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}
	return record, nil
}
