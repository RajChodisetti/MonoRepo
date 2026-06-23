package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

func TestMockGetReturnsRecord(t *testing.T) {
	t.Parallel()

	mock := &Mock{
		Records: map[string]Record{
			"schema_baseline": {
				Key:       "schema_baseline",
				Value:     json.RawMessage(`{"phase":"p1-e01"}`),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
	}

	record, err := mock.Get(context.Background(), "schema_baseline")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if record.Key != "schema_baseline" {
		t.Fatalf("Key = %q, want schema_baseline", record.Key)
	}
}

func TestMockGetReturnsNotFound(t *testing.T) {
	t.Parallel()

	mock := &Mock{Records: map[string]Record{}}

	_, err := mock.Get(context.Background(), "missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestMockGetReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	mock := &Mock{GetErr: repository.ErrNotFound}

	_, err := mock.Get(context.Background(), "schema_baseline")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, repository.ErrNotFound)
	}
}
