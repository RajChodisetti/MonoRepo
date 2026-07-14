package email

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"regexp"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var emailAddressPattern = regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`)

// CampaignArtifactFingerprint binds a quota claim to the exact persisted
// campaign artifact that was rendered immediately before the claim. Length
// prefixes keep the encoding unambiguous even when content contains control
// characters.
func CampaignArtifactFingerprint(subject, htmlBody, textBody, demoToken string) string {
	digest := sha256.New()
	writeFingerprintPart(digest, subject)
	writeFingerprintPart(digest, htmlBody)
	writeFingerprintPart(digest, textBody)
	writeFingerprintPart(digest, demoToken)
	return hex.EncodeToString(digest.Sum(nil))
}

func writeFingerprintPart(digest hash.Hash, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write([]byte(value))
}

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

func redactEmailAddresses(value string) string {
	return emailAddressPattern.ReplaceAllString(value, "[redacted-email]")
}
