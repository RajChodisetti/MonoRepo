package outreach

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type inboundStoreStub struct {
	historyID    string
	inserted     []Message
	attemptedKey string
	resultKey    string
	resultErr    error
}

func (store *inboundStoreStub) GetInboundSync(context.Context, string) (string, error) {
	return store.historyID, nil
}

func (store *inboundStoreStub) SetInboundSync(_ context.Context, _ string, historyID string) error {
	store.historyID = historyID
	return nil
}

func (store *inboundStoreStub) MarkInboundPollAttempt(_ context.Context, mailboxKey string) error {
	store.attemptedKey = mailboxKey
	return nil
}

func (store *inboundStoreStub) RecordInboundPollResult(_ context.Context, mailboxKey string, pollErr error) error {
	store.resultKey = mailboxKey
	store.resultErr = pollErr
	return nil
}

func (*inboundStoreStub) GetMessageByGmailID(context.Context, string, string) (Message, error) {
	return Message{}, pgx.ErrNoRows
}

func (*inboundStoreStub) GetOutboundByReplyToken(context.Context, uuid.UUID) (Message, error) {
	return Message{}, pgx.ErrNoRows
}

func (*inboundStoreStub) GetOutboundByRFCMessageID(context.Context, string) (Message, error) {
	return Message{}, pgx.ErrNoRows
}

func (*inboundStoreStub) GetOutboundByThreadID(context.Context, string) (Message, error) {
	return Message{}, pgx.ErrNoRows
}

func (*inboundStoreStub) FindRestaurantIDByEmail(context.Context, string) (uuid.UUID, error) {
	return uuid.Nil, pgx.ErrNoRows
}

func (store *inboundStoreStub) InsertMessage(_ context.Context, message Message) (Message, error) {
	store.inserted = append(store.inserted, message)
	return message, nil
}

type inboxReaderStub struct {
	query             string
	messageIDs        []string
	historyMessageIDs []string
	nextHistoryID     string
	messages          map[string]emailprovider.InboxMessage
	messageErrors     map[string]error
}

func (*inboxReaderStub) ProfileHistoryID(context.Context) (string, error) {
	return "history-10", nil
}

func (reader *inboxReaderStub) ListHistoryMessageIDs(context.Context, string) ([]string, string, error) {
	if reader.historyMessageIDs != nil || reader.nextHistoryID != "" {
		return reader.historyMessageIDs, reader.nextHistoryID, nil
	}
	return nil, "", errors.New("history should not be used during initial sync")
}

func (reader *inboxReaderStub) ListRecentMessageIDs(_ context.Context, query string) ([]string, error) {
	reader.query = query
	if reader.messageIDs != nil {
		return reader.messageIDs, nil
	}
	return []string{"current", "sent", "old"}, nil
}

func (reader *inboxReaderStub) GetMessage(_ context.Context, id string) (emailprovider.InboxMessage, error) {
	if err := reader.messageErrors[id]; err != nil {
		return emailprovider.InboxMessage{}, err
	}
	return reader.messages[id], nil
}

func TestRFCMessageIDsDedupes(t *testing.T) {
	ids := rfcMessageIDs("<one@example.com>", "<one@example.com> <two@example.com>")
	if len(ids) != 2 || ids[0] != "<one@example.com>" || ids[1] != "<two@example.com>" {
		t.Fatalf("rfcMessageIDs() = %#v", ids)
	}
}

func TestStoppedReasonInboundReply(t *testing.T) {
	if StoppedReasonInboundReply != "inbound_reply" {
		t.Fatalf("StoppedReasonInboundReply = %q", StoppedReasonInboundReply)
	}
}

func TestRecentMailQueryUsesTenDayInboxWindow(t *testing.T) {
	service := &InboundService{cfg: config.OutreachConfig{
		InboundLocalPart: "outreach",
		InboundDomain:    "tuvisolutions.com",
	}}
	got := service.recentMailQuery()
	if got != "in:inbox newer_than:10d" {
		t.Fatalf("recentMailQuery() = %q", got)
	}
}

func TestInboundPollCapturesEveryRecentInboxMessageForItsMailbox(t *testing.T) {
	now := time.Now().UTC()
	store := &inboundStoreStub{}
	reader := &inboxReaderStub{messages: map[string]emailprovider.InboxMessage{
		"current": {
			ID: "current", ThreadID: "thread-current", Inbox: true,
			From: "owner@example.com", To: "sales@example.com", Subject: "Question",
			BodyText: "Hello", ReceivedAt: now,
		},
		"sent": {
			ID: "sent", ThreadID: "thread-sent", Inbox: false,
			From: "sales@example.com", To: "owner@example.com", ReceivedAt: now,
		},
		"old": {
			ID: "old", ThreadID: "thread-old", Inbox: true,
			From: "old@example.com", To: "sales@example.com", ReceivedAt: now.Add(-11 * 24 * time.Hour),
		},
	}}
	service := NewInboundService(
		store,
		nil,
		reader,
		"sales-one",
		config.OutreachConfig{InboundEnabled: true},
		nil,
	)

	if err := service.Poll(context.Background()); err != nil {
		t.Fatalf("Poll() error = %v", err)
	}
	if reader.query != "in:inbox newer_than:10d" {
		t.Fatalf("query = %q", reader.query)
	}
	if store.historyID != "history-10" || store.attemptedKey != "sales-one" || store.resultKey != "sales-one" || store.resultErr != nil {
		t.Fatalf("sync state = %#v", store)
	}
	if len(store.inserted) != 1 {
		t.Fatalf("inserted = %#v, want one current inbox message", store.inserted)
	}
	message := store.inserted[0]
	if message.MailboxKey != "sales-one" || !message.Unmatched || message.GmailMessageID != "current" || !message.ReceivedAt.Equal(now) {
		t.Fatalf("inserted message = %#v", message)
	}
}

func TestInboundHistoryPollSkipsMessageDeletedBeforeFetchAndAdvancesCursor(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	store := &inboundStoreStub{historyID: "history-10"}
	reader := &inboxReaderStub{
		historyMessageIDs: []string{"gone", "current"},
		nextHistoryID:     "history-20",
		messages: map[string]emailprovider.InboxMessage{
			"current": {
				ID: "current", ThreadID: "thread-current", Inbox: true,
				From: "owner@example.com", To: "sales@example.com", Subject: "Question",
				BodyText: "Hello", ReceivedAt: now,
			},
		},
		messageErrors: map[string]error{
			"gone": fmt.Errorf("provider fetch: %w", emailprovider.ErrInboxMessageNotFound),
		},
	}
	service := NewInboundService(
		store,
		nil,
		reader,
		"sales-one",
		config.OutreachConfig{InboundEnabled: true},
		nil,
	)

	if err := service.Poll(context.Background()); err != nil {
		t.Fatalf("Poll() error = %v", err)
	}
	if store.historyID != "history-20" || store.resultErr != nil {
		t.Fatalf("sync state = %#v, want successful advanced cursor", store)
	}
	if len(store.inserted) != 1 || store.inserted[0].GmailMessageID != "current" {
		t.Fatalf("inserted = %#v, want only the available message", store.inserted)
	}
}

func TestInboundPollKeepsCursorOnMessageFetchFailure(t *testing.T) {
	t.Parallel()

	store := &inboundStoreStub{}
	reader := &inboxReaderStub{
		messageIDs: []string{"failed"},
		messages:   map[string]emailprovider.InboxMessage{},
		messageErrors: map[string]error{
			"failed": errors.New("gmail unavailable"),
		},
	}
	service := NewInboundService(
		store,
		nil,
		reader,
		"sales-one",
		config.OutreachConfig{InboundEnabled: true},
		nil,
	)

	err := service.Poll(context.Background())
	if err == nil || !strings.Contains(err.Error(), "gmail unavailable") {
		t.Fatalf("Poll() error = %v, want fetch failure", err)
	}
	if store.historyID != "" || store.resultErr == nil {
		t.Fatalf("sync state = %#v, want unchanged cursor and recorded failure", store)
	}
}

func TestSnapshotBodyPrefersText(t *testing.T) {
	if got := snapshotBody(" plain ", "<p>html</p>"); got != "plain" {
		t.Fatalf("snapshotBody() = %q", got)
	}
	if got := snapshotBody("  ", "<p>html</p>"); got != "<p>html</p>" {
		t.Fatalf("snapshotBody html fallback = %q", got)
	}
}
