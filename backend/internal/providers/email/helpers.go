package email

import (
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func resolveRecipient(emailCfg config.EmailConfig, to string) (deliveryTo, originalTo string) {
	to = strings.TrimSpace(to)
	originalTo = to
	deliveryTo = to
	if redirect := strings.TrimSpace(emailCfg.RedirectTo); redirect != "" {
		deliveryTo = redirect
	}
	return deliveryTo, originalTo
}

func formatFromAddress(emailCfg config.EmailConfig) string {
	from := strings.TrimSpace(emailCfg.FromAddress)
	fromName := strings.TrimSpace(emailCfg.FromName)
	if fromName != "" {
		return fromName + " <" + from + ">"
	}
	return from
}
