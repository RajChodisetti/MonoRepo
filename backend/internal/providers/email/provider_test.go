package email_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func TestNewFromConfigDisabledWhenSendingDisabled(t *testing.T) {
	provider, err := emailprovider.NewFromConfig(config.EmailConfig{
		Provider:       emailprovider.ProviderResend,
		APIKey:         "re_test_key",
		APIBaseURL:     "https://api.resend.com",
		FromAddress:    "contact@example.com",
		DisableSending: true,
	}, config.ZohoMailConfig{})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}

	result, err := provider.Send(context.Background(), emailprovider.SendRequest{
		To:      "lead@example.com",
		Subject: "Test",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !result.Skipped {
		t.Fatal("Send() skipped = false, want true")
	}
}

func TestNewFromConfigProviderDisabled(t *testing.T) {
	provider, err := emailprovider.NewFromConfig(config.EmailConfig{
		Provider: emailprovider.ProviderDisabled,
	}, config.ZohoMailConfig{})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider = nil")
	}
}

func TestNewFromConfigRejectsSMTP(t *testing.T) {
	_, err := emailprovider.NewFromConfig(config.EmailConfig{
		Provider:    "smtp",
		FromAddress: "contact@example.com",
	}, config.ZohoMailConfig{})
	if err == nil {
		t.Fatal("NewFromConfig() error = nil, want smtp unsupported error")
	}
	if !strings.Contains(err.Error(), "no longer supported") {
		t.Fatalf("NewFromConfig() error = %q", err.Error())
	}
}

func TestResendProviderSendViaHTTP(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/emails" {
			t.Fatalf("path = %q, want /emails", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(payload, &gotBody); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_123"}`))
	}))
	defer server.Close()

	provider := emailprovider.NewResendWithClient(config.EmailConfig{
		APIKey:      "re_test_key",
		FromAddress: "noreply@tuvisolutions.com",
		FromName:    "Tuvi Solutions",
	}, server.Client(), server.URL)

	result, err := provider.Send(context.Background(), emailprovider.SendRequest{
		To:       "owner@restaurant.com",
		Subject:  "Three website designs ready",
		HTMLBody: "<p>Hello</p>",
		TextBody: "Hello",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.ProviderMessageID != "msg_123" {
		t.Fatalf("ProviderMessageID = %q, want msg_123", result.ProviderMessageID)
	}
	if gotAuth != "Bearer re_test_key" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotBody["subject"] != "Three website designs ready" {
		t.Fatalf("subject = %v", gotBody["subject"])
	}
}

func TestResendProviderRedirectsRecipient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		to, ok := body["to"].([]any)
		if !ok || len(to) != 1 || to[0] != "dev@tuvisolutions.com" {
			t.Fatalf("to = %v, want dev@tuvisolutions.com", body["to"])
		}
		_, _ = w.Write([]byte(`{"id":"msg_redirect"}`))
	}))
	defer server.Close()

	provider := emailprovider.NewResendWithClient(config.EmailConfig{
		APIKey:      "re_test_key",
		FromAddress: "noreply@tuvisolutions.com",
		RedirectTo:  "dev@tuvisolutions.com",
	}, server.Client(), server.URL)

	result, err := provider.Send(context.Background(), emailprovider.SendRequest{
		To:       "owner@restaurant.com",
		Subject:  "Test",
		HTMLBody: "<p>Hi</p>",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.RedirectedTo != "dev@tuvisolutions.com" {
		t.Fatalf("RedirectedTo = %q", result.RedirectedTo)
	}
}
