package email_test

import (
	"context"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

func TestNewFromConfigDisabledWhenSendingDisabled(t *testing.T) {
	provider, err := emailprovider.NewFromConfig(config.EmailConfig{
		Provider:       "smtp",
		FromAddress:    "contact@example.com",
		DisableSending: true,
	}, config.SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "contact@example.com",
		Password: "secret",
	})
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
	}, config.SMTPConfig{})
	if err != nil {
		t.Fatalf("NewFromConfig() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider = nil")
	}
}
