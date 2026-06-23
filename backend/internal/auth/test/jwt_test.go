package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestTokenManagerIssueAndParse(t *testing.T) {
	manager := auth.NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour)
	userID := uuid.New()

	token, expiresAt, err := manager.IssueToken(userID, "dev@example.com", auth.RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("IssueToken() returned empty token")
	}
	if !expiresAt.After(time.Now()) {
		t.Fatalf("expiresAt = %v, want future time", expiresAt)
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != userID.String() {
		t.Fatalf("UserID = %q, want %q", claims.UserID, userID.String())
	}
	if claims.Email != "dev@example.com" {
		t.Fatalf("Email = %q, want dev@example.com", claims.Email)
	}
	if claims.Role != auth.RoleDeveloper {
		t.Fatalf("Role = %q, want %q", claims.Role, auth.RoleDeveloper)
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	manager := auth.NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour)

	_, err := manager.ParseToken("not-a-valid-token")
	if err == nil {
		t.Fatal("ParseToken() error = nil, want error")
	}
}

func TestTokenManagerRejectsWrongSecret(t *testing.T) {
	issuer := auth.NewTokenManager("local-dev-token-secret-change-me-32chars", time.Hour)
	parser := auth.NewTokenManager("another-secret-that-is-long-enough-32c", time.Hour)

	token, _, err := issuer.IssueToken(uuid.New(), "dev@example.com", auth.RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	_, err = parser.ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken() error = nil, want error for wrong secret")
	}
}

func TestClaimsPrincipal(t *testing.T) {
	userID := uuid.New()
	claims := auth.Claims{
		UserID: userID.String(),
		Email:  "dev@example.com",
		Role:   auth.RoleDeveloper,
	}

	principal, err := claims.Principal()
	if err != nil {
		t.Fatalf("Principal() error = %v", err)
	}
	if principal.UserID != userID {
		t.Fatalf("UserID = %v, want %v", principal.UserID, userID)
	}
}
