package demos

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

const demoTokenBytes = 32

func GenerateDemoToken() (string, error) {
	buf := make([]byte, demoTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate demo token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashDemoToken(token string) (string, error) {
	return auth.HashPassword(token)
}

func CheckDemoToken(hash, token string) error {
	return auth.CheckPassword(hash, token)
}
