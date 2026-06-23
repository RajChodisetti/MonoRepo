package auth_test

import (
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if err := auth.CheckPassword(hash, "password123"); err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}
	if err := auth.CheckPassword(hash, "wrong-password"); err == nil {
		t.Fatal("CheckPassword() error = nil, want mismatch error")
	}
}
