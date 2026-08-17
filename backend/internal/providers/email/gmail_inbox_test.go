package email

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestGmailInboxPaginatesInboxAndHistoryAndFiltersLabels(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"access-token","token_type":"Bearer","expires_in":3600}`))
		case "/gmail/v1/users/me/profile":
			_, _ = w.Write([]byte(`{"historyId":"123"}`))
		case "/gmail/v1/users/me/messages":
			if r.URL.Query().Get("labelIds") != "INBOX" || r.URL.Query().Get("maxResults") != "500" {
				t.Fatalf("message list query = %q, want INBOX with 500 results per page", r.URL.RawQuery)
			}
			if r.URL.Query().Get("q") != "in:inbox newer_than:10d" {
				t.Fatalf("message list q = %q", r.URL.Query().Get("q"))
			}
			if r.URL.Query().Get("pageToken") == "" {
				_, _ = w.Write([]byte(`{"messages":[{"id":"one"},{"id":"two"}],"nextPageToken":"next"}`))
				return
			}
			_, _ = w.Write([]byte(`{"messages":[{"id":"two"},{"id":"three"}]}`))
		case "/gmail/v1/users/me/history":
			if r.URL.Query().Get("startHistoryId") != "100" || r.URL.Query().Get("historyTypes") != "messageAdded" {
				t.Fatalf("history query = %q", r.URL.RawQuery)
			}
			if r.URL.Query().Get("pageToken") == "" {
				_, _ = w.Write([]byte(`{"history":[{"messagesAdded":[{"message":{"id":"one"}},{"message":{"id":"two"}}]}],"historyId":"110","nextPageToken":"next"}`))
				return
			}
			_, _ = w.Write([]byte(`{"history":[{"messagesAdded":[{"message":{"id":"two"}},{"message":{"id":"three"}}]}],"historyId":"120"}`))
		case "/gmail/v1/users/me/messages/one":
			_, _ = w.Write([]byte(`{"id":"one","threadId":"thread-one","labelIds":["INBOX"],"internalDate":"1786665600000","payload":{"mimeType":"text/plain","headers":[{"name":"From","value":"Owner <owner@example.com>"},{"name":"To","value":"sales@example.com"},{"name":"Subject","value":"Hello"}],"body":{"data":"aGVsbG8"}}}`))
		case "/gmail/v1/users/me/messages/sent":
			_, _ = w.Write([]byte(`{"id":"sent","threadId":"thread-sent","labelIds":["SENT"],"payload":{"mimeType":"text/plain","body":{"data":"aGVsbG8"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	created, err := newGmailProvider(
		config.EmailConfig{},
		config.GmailMailConfig{
			MailboxEmail: "sales@example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RefreshToken: "refresh-token",
		},
		server.Client(),
		server.URL+"/gmail/v1",
		server.URL+"/token",
		false,
	)
	if err != nil {
		t.Fatalf("newGmailProvider() error = %v", err)
	}
	provider := created.(*gmailProvider)

	profileID, err := provider.ProfileHistoryID(context.Background())
	if err != nil || profileID != "123" {
		t.Fatalf("ProfileHistoryID() = %q, %v", profileID, err)
	}
	messageIDs, err := provider.ListRecentMessageIDs(context.Background(), "in:inbox newer_than:10d")
	if err != nil {
		t.Fatalf("ListRecentMessageIDs() error = %v", err)
	}
	if len(messageIDs) != 3 || messageIDs[0] != "one" || messageIDs[1] != "two" || messageIDs[2] != "three" {
		t.Fatalf("ListRecentMessageIDs() = %#v", messageIDs)
	}
	historyIDs, newHistoryID, err := provider.ListHistoryMessageIDs(context.Background(), "100")
	if err != nil {
		t.Fatalf("ListHistoryMessageIDs() error = %v", err)
	}
	if newHistoryID != "120" || len(historyIDs) != 3 || historyIDs[2] != "three" {
		t.Fatalf("ListHistoryMessageIDs() = %#v, %q", historyIDs, newHistoryID)
	}
	inboxMessage, err := provider.GetMessage(context.Background(), "one")
	if err != nil || !inboxMessage.Inbox || inboxMessage.From != "owner@example.com" || inboxMessage.BodyText != "hello" || inboxMessage.ReceivedAt.IsZero() {
		t.Fatalf("GetMessage(one) = %#v, %v", inboxMessage, err)
	}
	sentMessage, err := provider.GetMessage(context.Background(), "sent")
	if err != nil || sentMessage.Inbox {
		t.Fatalf("GetMessage(sent) = %#v, %v", sentMessage, err)
	}
}

func TestGmailInboxClassifiesOnlyMessageGetNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"access-token","token_type":"Bearer","expires_in":3600}`))
		case "/gmail/v1/users/me/messages/gone":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND"}}`))
		case "/gmail/v1/users/me/history":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	created, err := newGmailProvider(
		config.EmailConfig{},
		config.GmailMailConfig{
			MailboxEmail: "sales@example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RefreshToken: "refresh-token",
		},
		server.Client(),
		server.URL+"/gmail/v1",
		server.URL+"/token",
		false,
	)
	if err != nil {
		t.Fatalf("newGmailProvider() error = %v", err)
	}
	provider := created.(*gmailProvider)

	if _, err := provider.GetMessage(context.Background(), "gone"); !errors.Is(err, ErrInboxMessageNotFound) {
		t.Fatalf("GetMessage(gone) error = %v, want ErrInboxMessageNotFound", err)
	}
	if _, _, err := provider.ListHistoryMessageIDs(context.Background(), "expired"); err == nil || errors.Is(err, ErrInboxMessageNotFound) {
		t.Fatalf("ListHistoryMessageIDs(expired) error = %v, want an unclassified history error", err)
	}
}
