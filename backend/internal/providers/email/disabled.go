package email

import (
	"context"
	"log/slog"
)

type disabledProvider struct {
	log *slog.Logger
}

func NewDisabled() Provider {
	return &disabledProvider{log: slog.Default()}
}

func (provider *disabledProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	provider.log.InfoContext(ctx, "email_send_skipped",
		"to", req.To,
		"subject", req.Subject,
		"reason", "email sending disabled",
	)
	return SendResult{Skipped: true}, nil
}
