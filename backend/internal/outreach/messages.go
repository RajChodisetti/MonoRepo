package outreach

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

const (
	MessageDirectionOutbound  = "outbound"
	MessageDirectionInbound   = "inbound"
	MailboxKeyInbound         = "inbound"
	StoppedReasonInboundReply = "inbound_reply"
)

type Message struct {
	ID                uuid.UUID  `json:"id"`
	RestaurantID      *uuid.UUID `json:"restaurant_id,omitempty"`
	CampaignID        *uuid.UUID `json:"campaign_id,omitempty"`
	DeliveryAttemptID *uuid.UUID `json:"delivery_attempt_id,omitempty"`
	ReplyToken        *uuid.UUID `json:"reply_token,omitempty"`
	Direction         string     `json:"direction"`
	FromEmail         string     `json:"from_email"`
	ToEmail           string     `json:"to_email"`
	ReplyTo           string     `json:"reply_to,omitempty"`
	Subject           string     `json:"subject"`
	BodyText          string     `json:"body_text"`
	GmailMessageID    string     `json:"gmail_message_id,omitempty"`
	GmailThreadID     string     `json:"gmail_thread_id,omitempty"`
	RFCMessageID      string     `json:"rfc_message_id,omitempty"`
	MailboxKey        string     `json:"mailbox_key,omitempty"`
	Unmatched         bool       `json:"unmatched"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	ReceivedAt        time.Time  `json:"received_at"`
	CreatedAt         time.Time  `json:"created_at"`
	RestaurantName    string     `json:"restaurant_name,omitempty"`
}

type InboxThread struct {
	RestaurantID   *uuid.UUID `json:"restaurant_id,omitempty"`
	RestaurantName string     `json:"restaurant_name,omitempty"`
	Email          string     `json:"email,omitempty"`
	MailboxKey     string     `json:"mailbox_key"`
	MailboxEmail   string     `json:"mailbox_email"`
	Unmatched      bool       `json:"unmatched"`
	UnreadCount    int        `json:"unread_count"`
	Subject        string     `json:"subject"`
	TextSnippet    string     `json:"text_snippet"`
	FromEmail      string     `json:"from_email"`
	ToEmail        string     `json:"to_email"`
	ReceivedAt     time.Time  `json:"received_at"`
	LastDirection  string     `json:"last_direction"`
	LastSnippet    string     `json:"last_snippet"`
	LastAt         time.Time  `json:"last_at"`
	LastMessageID  uuid.UUID  `json:"last_message_id"`
	ReplyMessageID uuid.UUID  `json:"reply_message_id"`
}

type InboxMailboxStatus struct {
	MailboxKey    string     `json:"mailbox_key"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

type InboxList struct {
	Threads   []InboxThread        `json:"threads"`
	Mailboxes []InboxMailboxStatus `json:"mailboxes"`
	Total     int                  `json:"total"`
}

type MessageList struct {
	Messages []Message `json:"messages"`
}

type ReplyMessageInput struct {
	Subject  string `json:"subject,omitempty"`
	BodyText string `json:"body_text"`
}

func prepareInboxReply(target Message, input ReplyMessageInput) (emailprovider.SendRequest, error) {
	if target.Direction != MessageDirectionInbound {
		return emailprovider.SendRequest{}, fmt.Errorf("%w: only inbound messages can be replied to", ErrInvalidInboxReply)
	}
	body := strings.TrimSpace(input.BodyText)
	if body == "" || len(body) > 10000 || !utf8.ValidString(body) || strings.ContainsRune(body, '\x00') {
		return emailprovider.SendRequest{}, fmt.Errorf("%w: body_text must be valid plain text between 1 and 10000 bytes", ErrInvalidInboxReply)
	}
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		subject = strings.TrimSpace(target.Subject)
		if subject == "" {
			subject = "(no subject)"
		}
		if !strings.HasPrefix(strings.ToLower(subject), "re:") {
			subject = "Re: " + subject
		}
	}
	if len(subject) > 200 || strings.ContainsAny(subject, "\r\n") {
		return emailprovider.SendRequest{}, fmt.Errorf("%w: subject must be a single line between 1 and 200 bytes", ErrInvalidInboxReply)
	}
	recipient, err := cleanTestRecipient(target.FromEmail)
	if err != nil {
		return emailprovider.SendRequest{}, fmt.Errorf("%w: inbound sender address is invalid", ErrInvalidInboxReply)
	}
	request := emailprovider.SendRequest{
		To:         recipient,
		Subject:    subject,
		TextBody:   body,
		ThreadID:   strings.TrimSpace(target.GmailThreadID),
		InReplyTo:  strings.TrimSpace(target.RFCMessageID),
		References: strings.TrimSpace(target.RFCMessageID),
		Metadata: map[string]string{
			"purpose":             "outreach_inbox_reply",
			"in_reply_to_message": target.ID.String(),
		},
	}
	if receivedAt, err := cleanTestRecipient(target.ToEmail); err == nil {
		request.FromEmail = receivedAt
	}
	return request, nil
}

func snapshotBody(textBody, htmlBody string) string {
	if body := strings.TrimSpace(textBody); body != "" {
		return body
	}
	return strings.TrimSpace(htmlBody)
}

func snippet(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if max < 1 || len(value) <= max {
		return value
	}
	return strings.TrimSpace(value[:max]) + "…"
}
