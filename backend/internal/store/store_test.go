package store

import (
	"context"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/repositories/metadata"
)

func TestVerifyFoundationWithMockRepository(t *testing.T) {
	t.Parallel()

	store := &Store{
		Metadata: &metadata.Mock{
			Records: map[string]metadata.Record{
				schemaBaselineKey: {Key: schemaBaselineKey},
			},
		},
	}

	if err := store.VerifyFoundation(context.Background()); err != nil {
		t.Fatalf("VerifyFoundation() error = %v", err)
	}
}

func TestVerifyFoundationFailsWhenBaselineMissing(t *testing.T) {
	t.Parallel()

	store := &Store{
		Metadata: &metadata.Mock{Records: map[string]metadata.Record{}},
	}

	err := store.VerifyFoundation(context.Background())
	if err == nil {
		t.Fatal("VerifyFoundation() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "foundation migration not applied") {
		t.Fatalf("VerifyFoundation() error = %q, want foundation migration message", err.Error())
	}
}
