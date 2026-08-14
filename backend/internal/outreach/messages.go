package outreach

import (
	"strings"
	"time"

	"github.com/google/uuid"
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
	CreatedAt         time.Time  `json:"created_at"`
	RestaurantName    string     `json:"restaurant_name,omitempty"`
}

type InboxThread struct {
	RestaurantID   *uuid.UUID `json:"restaurant_id,omitempty"`
	RestaurantName string     `json:"restaurant_name,omitempty"`
	Email          string     `json:"email,omitempty"`
	Unmatched      bool       `json:"unmatched"`
	UnreadCount    int        `json:"unread_count"`
	LastDirection  string     `json:"last_direction"`
	LastSnippet    string     `json:"last_snippet"`
	LastAt         time.Time  `json:"last_at"`
	LastMessageID  uuid.UUID  `json:"last_message_id"`
}

type InboxList struct {
	Threads []InboxThread `json:"threads"`
	Total   int           `json:"total"`
}

type MessageList struct {
	Messages []Message `json:"messages"`
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
