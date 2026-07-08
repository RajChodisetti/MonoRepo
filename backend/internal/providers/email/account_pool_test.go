package email_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type countingProvider struct {
	sends int
}

func (provider *countingProvider) Send(ctx context.Context, req email.SendRequest) (email.SendResult, error) {
	provider.sends++
	return email.SendResult{ProviderMessageID: "test"}, nil
}

func TestAccountPoolRotatesEveryFiftySends(t *testing.T) {
	t.Parallel()

	providers := []email.Provider{
		&countingProvider{},
		&countingProvider{},
		&countingProvider{},
	}
	pool, err := email.NewAccountPool(providers, 50, 150)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}

	for i := 0; i < 150; i++ {
		if _, err := pool.Send(context.Background(), email.SendRequest{To: "lead@example.com", HTMLBody: "<p>hi</p>"}); err != nil {
			t.Fatalf("Send(%d) error = %v", i, err)
		}
	}

	if _, err := pool.Send(context.Background(), email.SendRequest{To: "lead@example.com", HTMLBody: "<p>hi</p>"}); !errors.Is(err, email.ErrAccountsExhausted) {
		t.Fatalf("Send(151) error = %v, want ErrAccountsExhausted", err)
	}

	for index, provider := range providers {
		counter := provider.(*countingProvider)
		if counter.sends != 50 {
			t.Fatalf("provider %d sends = %d, want 50", index+1, counter.sends)
		}
	}
}
