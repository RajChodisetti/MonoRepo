package db

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestConnectRequiredRejectsMissingURL(t *testing.T) {
	t.Parallel()

	_, err := ConnectRequired(context.Background(), config.DatabaseConfig{}, time.Second)
	if err == nil {
		t.Fatal("ConnectRequired() error = nil, want error")
	}
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("ConnectRequired() error = %v, want ErrNotConfigured", err)
	}
}

func TestConnectRequiredRejectsUnreachableDatabase(t *testing.T) {
	t.Parallel()

	cfg := config.DatabaseConfig{
		URL:            "postgres://postgres:postgres@127.0.0.1:1/restaurant_platform?sslmode=disable",
		MaxConns:       1,
		ConnectTimeout: time.Second,
	}

	_, err := ConnectRequired(context.Background(), cfg, 2*time.Second)
	if err == nil {
		t.Fatal("ConnectRequired() error = nil, want error")
	}
	if !errors.Is(err, ErrNotReady) {
		t.Fatalf("ConnectRequired() error = %v, want ErrNotReady", err)
	}
	if !strings.Contains(err.Error(), "database is not ready") {
		t.Fatalf("ConnectRequired() error = %q, want readiness detail", err.Error())
	}
}
