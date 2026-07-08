package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

const ProviderDisabled = "disabled"

type SendRequest struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
	ReplyTo  string
	Metadata map[string]string
}

type SendResult struct {
	ProviderMessageID string
	RedirectedTo      string
	Skipped           bool
}

type Provider interface {
	Send(ctx context.Context, req SendRequest) (SendResult, error)
}

func NewFromConfig(emailCfg config.EmailConfig, zohoCfg config.ZohoMailConfig) (Provider, error) {
	provider := strings.ToLower(strings.TrimSpace(emailCfg.Provider))
	if provider == "" || provider == ProviderDisabled || emailCfg.DisableSending {
		return NewDisabled(), nil
	}

	switch provider {
	case ProviderResend, "http", "https":
		return NewResend(emailCfg)
	case "zoho":
		return NewZoho(emailCfg, zohoCfg)
	case "smtp":
		return nil, fmt.Errorf("email provider smtp is no longer supported; use EMAIL_PROVIDER=resend or zoho")
	default:
		return nil, fmt.Errorf("unsupported email provider %q (supported: resend, zoho, http, https, disabled)", emailCfg.Provider)
	}
}
