package email_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func TestGmailProviderSendsViaHTTPSAPIContract(t *testing.T) {
	var tokenForm url.Values
	var rawMessage string
	var requestThreadID string
	var tokenRequests int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenRequests++
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token form: %v", err)
			}
			tokenForm = r.Form
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access-token","token_type":"Bearer","expires_in":3600}`))
		case "/gmail/v1/users/me/messages/send":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %q, want POST", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer access-token" {
				t.Fatalf("Authorization header was not set from refreshed token")
			}
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read send body: %v", err)
			}
			var body struct {
				Raw      string `json:"raw"`
				ThreadID string `json:"threadId"`
			}
			if err := json.Unmarshal(payload, &body); err != nil {
				t.Fatalf("decode send body: %v", err)
			}
			decoded, err := base64.RawURLEncoding.DecodeString(body.Raw)
			if err != nil {
				t.Fatalf("decode raw MIME: %v", err)
			}
			rawMessage = string(decoded)
			requestThreadID = body.ThreadID
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"gmail-message-id","threadId":"gmail-thread-id"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := emailprovider.NewGmailWithClient(
		config.EmailConfig{FromName: "Tuvi Solutions"},
		config.GmailMailConfig{
			MailboxEmail: "sales1@example.com",
			FromEmail:    "sales1@example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RefreshToken: "refresh-token",
		},
		server.Client(),
		server.URL+"/gmail/v1",
		server.URL+"/token",
	)
	if err != nil {
		t.Fatalf("NewGmailWithClient() error = %v", err)
	}

	result, err := provider.Send(context.Background(), emailprovider.SendRequest{
		To:         "owner@restaurant.example",
		Subject:    "Your restaurant demo",
		TextBody:   "Text version",
		HTMLBody:   "<p>HTML version</p>",
		ReplyTo:    "contact@example.com",
		ThreadID:   "original-thread-id",
		InReplyTo:  "<original@example.com>",
		References: "<earlier@example.com> <original@example.com>",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.ProviderMessageID != "gmail-message-id" {
		t.Fatalf("ProviderMessageID = %q, want gmail-message-id", result.ProviderMessageID)
	}
	if result.ProviderThreadID != "gmail-thread-id" {
		t.Fatalf("ProviderThreadID = %q, want gmail-thread-id", result.ProviderThreadID)
	}
	if requestThreadID != "original-thread-id" {
		t.Fatalf("request threadId = %q, want original-thread-id", requestThreadID)
	}
	if result.RFCMessageID == "" || !strings.Contains(result.RFCMessageID, "@example.com") {
		t.Fatalf("RFCMessageID = %q, want a Message-ID using the from domain", result.RFCMessageID)
	}
	firstRawMessage := rawMessage
	if _, err := provider.Send(context.Background(), emailprovider.SendRequest{
		To:       "owner2@restaurant.example",
		Subject:  "Your second restaurant demo",
		TextBody: "Text version",
	}); err != nil {
		t.Fatalf("second Send() error = %v", err)
	}
	if tokenRequests != 1 {
		t.Fatalf("token requests = %d, want 1 cached refresh", tokenRequests)
	}
	if tokenForm.Get("grant_type") != "refresh_token" || tokenForm.Get("client_id") != "client-id" || tokenForm.Get("refresh_token") != "refresh-token" {
		t.Fatalf("token refresh form is incomplete")
	}
	for _, expected := range []string{
		"From: \"Tuvi Solutions\" <sales1@example.com>",
		"To: <owner@restaurant.example>",
		"Reply-To: <contact@example.com>",
		"In-Reply-To: <original@example.com>",
		"References: <earlier@example.com> <original@example.com>",
		"Message-ID:",
		"MIME-Version: 1.0",
		"multipart/alternative",
	} {
		if !strings.Contains(firstRawMessage, expected) {
			t.Fatalf("raw MIME message missing %q", expected)
		}
	}
}

func TestGmailProviderDoesNotForwardCredentialsAcrossRedirect(t *testing.T) {
	redirectRequests := 0
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectRequests++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer receiver.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, receiver.URL, http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	provider, err := emailprovider.NewGmailWithClient(
		config.EmailConfig{},
		config.GmailMailConfig{
			MailboxEmail: "sales1@example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RefreshToken: "refresh-token",
		},
		redirector.Client(),
		redirector.URL+"/gmail/v1",
		redirector.URL+"/token",
	)
	if err != nil {
		t.Fatalf("NewGmailWithClient() error = %v", err)
	}

	_, err = provider.Send(context.Background(), emailprovider.SendRequest{
		To:       "owner@restaurant.example",
		Subject:  "Your restaurant demo",
		TextBody: "hello",
	})
	if err == nil {
		t.Fatal("Send() error = nil, want redirect rejection")
	}
	if redirectRequests != 0 {
		t.Fatalf("redirect target requests = %d, want 0", redirectRequests)
	}
}

func TestGmailProviderRejectsHeaderInjection(t *testing.T) {
	provider, err := emailprovider.NewGmailWithClient(
		config.EmailConfig{},
		config.GmailMailConfig{
			MailboxEmail: "sales1@example.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			RefreshToken: "refresh-token",
		},
		http.DefaultClient,
		"http://localhost/gmail/v1",
		"http://localhost/token",
	)
	if err != nil {
		t.Fatalf("NewGmailWithClient() error = %v", err)
	}

	_, err = provider.Send(context.Background(), emailprovider.SendRequest{
		To:       "owner@restaurant.example",
		Subject:  "safe\r\nBcc: attacker@example.com",
		TextBody: "hello",
	})
	if err == nil || !strings.Contains(err.Error(), "newline") {
		t.Fatalf("Send() error = %v, want newline rejection", err)
	}
}

func TestGmailProviderRejectsInvalidStaticEmailConfig(t *testing.T) {
	tests := []struct {
		name     string
		emailCfg config.EmailConfig
		want     string
	}{
		{name: "from name newline", emailCfg: config.EmailConfig{FromName: "Tuvi\r\nBcc: attacker@example.com"}, want: "newline"},
		{name: "invalid redirect", emailCfg: config.EmailConfig{RedirectTo: "not-an-email"}, want: "redirect recipient"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := emailprovider.NewGmailWithClient(
				test.emailCfg,
				config.GmailMailConfig{
					MailboxEmail: "sales1@example.com",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RefreshToken: "refresh-token",
				},
				http.DefaultClient,
				"http://localhost/gmail/v1",
				"http://localhost/token",
			)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewGmailWithClient() error = %v, want %q", err, test.want)
			}
		})
	}
}
