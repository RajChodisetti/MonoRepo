package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

const ProviderDisabled = "disabled"

type SendRequest struct {
	To          string
	Subject     string
	HTMLBody    string
	TextBody    string
	ReplyTo     string
	Metadata    map[string]string
}

type SendResult struct {
	ProviderMessageID string
	RedirectedTo      string
	Skipped           bool
}

type Provider interface {
	Send(ctx context.Context, req SendRequest) (SendResult, error)
}

func NewFromConfig(emailCfg config.EmailConfig, smtpCfg config.SMTPConfig) (Provider, error) {
	provider := strings.ToLower(strings.TrimSpace(emailCfg.Provider))
	if provider == "" || provider == ProviderDisabled || emailCfg.DisableSending {
		return NewDisabled(), nil
	}

	switch provider {
	case "smtp":
		return NewSMTP(emailCfg, smtpCfg)
	default:
		return nil, fmt.Errorf("unsupported email provider %q", emailCfg.Provider)
	}
}
