package email

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

func ReplyToAddress(localPart, domain string, token uuid.UUID) string {
	localPart = strings.ToLower(strings.TrimSpace(localPart))
	domain = strings.ToLower(strings.TrimSpace(domain))
	if localPart == "" || domain == "" || token == uuid.Nil {
		return ""
	}
	return fmt.Sprintf("%s+%s@%s", localPart, token.String(), domain)
}

func ParseReplyToken(address, localPart, domain string) (uuid.UUID, bool) {
	localPart = strings.ToLower(strings.TrimSpace(localPart))
	domain = strings.ToLower(strings.TrimSpace(domain))
	if localPart == "" || domain == "" {
		return uuid.Nil, false
	}
	for _, raw := range splitAddressList(address) {
		parsed, err := mail.ParseAddress(raw)
		if err != nil {
			continue
		}
		mailbox := strings.ToLower(strings.TrimSpace(parsed.Address))
		prefix := localPart + "+"
		suffix := "@" + domain
		if !strings.HasPrefix(mailbox, prefix) || !strings.HasSuffix(mailbox, suffix) {
			continue
		}
		token := strings.TrimSuffix(strings.TrimPrefix(mailbox, prefix), suffix)
		id, err := uuid.Parse(token)
		if err != nil {
			continue
		}
		return id, true
	}
	return uuid.Nil, false
}

func splitAddressList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
