package email_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type countingProvider struct {
	sends int
}

type directResultProvider struct {
	sends  int
	result email.SendResult
	err    error
}

func (provider *directResultProvider) Send(context.Context, email.SendRequest) (email.SendResult, error) {
	provider.sends++
	return provider.result, provider.err
}

func TestNewAccountPoolFromConfigUsesUIControlInsteadOfLegacyEmailFlag(t *testing.T) {
	t.Parallel()

	_, err := email.NewAccountPoolFromConfig(
		config.EmailConfig{Provider: "gmail", DisableSending: true},
		config.OutreachConfig{
			BulkMax:          40,
			EmailsPerAccount: 40,
			GoogleWorkspaceAccounts: []config.GmailMailConfig{{
				MailboxEmail: "sales@example.com",
				ClientID:     "client",
				ClientSecret: "secret",
				RefreshToken: "refresh",
			}},
		},
	)
	if err != nil {
		t.Fatalf("NewAccountPoolFromConfig() error = %v", err)
	}
}

func TestNewAccountPoolFromConfigAcceptsGoogleWorkspaceAccount(t *testing.T) {
	t.Parallel()

	pool, err := email.NewAccountPoolFromConfig(
		config.EmailConfig{},
		config.OutreachConfig{
			BulkMax:          40,
			EmailsPerAccount: 40,
			GoogleWorkspaceAccounts: []config.GmailMailConfig{{
				AccountKey:   "workspace-sales-1",
				MailboxEmail: "sales1@example.com",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RefreshToken: "refresh-token",
			}},
		},
	)
	if err != nil {
		t.Fatalf("NewAccountPoolFromConfig() error = %v", err)
	}
	if pool == nil {
		t.Fatal("NewAccountPoolFromConfig() pool = nil")
	}
}

func TestNewAccountPoolFromConfigRejectsEquivalentGoogleWorkspaceMailboxes(t *testing.T) {
	t.Parallel()

	_, err := email.NewAccountPoolFromConfig(
		config.EmailConfig{},
		config.OutreachConfig{
			BulkMax:          80,
			EmailsPerAccount: 40,
			GoogleWorkspaceAccounts: []config.GmailMailConfig{
				{
					AccountKey:   "workspace-sales-1",
					MailboxEmail: "sales1@example.com",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RefreshToken: "refresh-token-1",
				},
				{
					AccountKey:   "workspace-sales-alias",
					MailboxEmail: "<sales1@example.com>",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RefreshToken: "refresh-token-2",
				},
			},
		},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicates mailbox") {
		t.Fatalf("NewAccountPoolFromConfig() error = %v, want duplicate mailbox rejection", err)
	}
}

func TestNewAccountPoolFromConfigRejectsInvalidRedirectBeforeUse(t *testing.T) {
	t.Parallel()

	_, err := email.NewAccountPoolFromConfig(
		config.EmailConfig{RedirectTo: "not-an-email"},
		config.OutreachConfig{
			BulkMax:          40,
			EmailsPerAccount: 40,
			GoogleWorkspaceAccounts: []config.GmailMailConfig{{
				MailboxEmail: "sales1@example.com",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RefreshToken: "refresh-token",
			}},
		},
	)
	if err == nil || !strings.Contains(err.Error(), "redirect recipient") {
		t.Fatalf("NewAccountPoolFromConfig() error = %v, want redirect validation error", err)
	}
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

func TestAccountPoolResetStartsANewManualRunAllowance(t *testing.T) {
	t.Parallel()

	provider := &countingProvider{}
	pool, err := email.NewAccountPool([]email.Provider{provider}, 1, 1)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	if _, err := pool.Send(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"}); err != nil {
		t.Fatalf("first Send() error = %v", err)
	}
	if _, err := pool.Send(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"}); !errors.Is(err, email.ErrAccountsExhausted) {
		t.Fatalf("second Send() error = %v, want ErrAccountsExhausted", err)
	}

	pool.Reset()
	if _, err := pool.Send(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"}); err != nil {
		t.Fatalf("Send() after Reset error = %v", err)
	}
}

func TestAccountPoolSendDirectSkipsOnlySafeUnavailableAccounts(t *testing.T) {
	t.Parallel()

	unavailable := &directResultProvider{err: fmt.Errorf("%w: revoked refresh token", email.ErrAccountUnavailable)}
	healthy := &directResultProvider{result: email.SendResult{ProviderMessageID: "accepted"}}
	pool, err := email.NewAccountPool([]email.Provider{unavailable, healthy}, 1, 2)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}
	direct, err := pool.AcquireDirect(context.Background())
	if err != nil {
		t.Fatalf("AcquireDirect() error = %v", err)
	}

	for index := range 2 {
		result, sendErr := direct.Send(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"})
		if sendErr != nil {
			t.Fatalf("direct Send(%d) error = %v", index+1, sendErr)
		}
		if result.ProviderMessageID != "accepted" || result.AccountKey != "in-memory-2" {
			t.Fatalf("direct Send(%d) result = %#v", index+1, result)
		}
	}
	if unavailable.sends != 1 || healthy.sends != 2 {
		t.Fatalf("provider sends = unavailable %d, healthy %d; want 1, 2", unavailable.sends, healthy.sends)
	}

	secondOperation, err := pool.AcquireDirect(context.Background())
	if err != nil {
		t.Fatalf("second AcquireDirect() error = %v", err)
	}
	if _, err := secondOperation.Send(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"}); err != nil {
		t.Fatalf("second operation Send() error = %v", err)
	}
	if unavailable.sends != 2 || healthy.sends != 3 {
		t.Fatalf("provider sends after fresh snapshot = unavailable %d, healthy %d; want 2, 3", unavailable.sends, healthy.sends)
	}
}

func TestAccountPoolSendDirectDoesNotRetryAmbiguousFailure(t *testing.T) {
	t.Parallel()

	ambiguous := &directResultProvider{err: errors.New("provider response could not be read")}
	healthy := &directResultProvider{result: email.SendResult{ProviderMessageID: "must-not-send"}}
	pool, err := email.NewAccountPool([]email.Provider{ambiguous, healthy}, 1, 2)
	if err != nil {
		t.Fatalf("NewAccountPool() error = %v", err)
	}

	result, err := pool.SendDirect(context.Background(), email.SendRequest{To: "lead@example.com", TextBody: "hi"})
	if err == nil || errors.Is(err, email.ErrAccountUnavailable) {
		t.Fatalf("SendDirect() error = %v, want unchanged ambiguous error", err)
	}
	if result.AccountKey != "in-memory-1" || ambiguous.sends != 1 || healthy.sends != 0 {
		t.Fatalf("result/sends = %#v, %d/%d", result, ambiguous.sends, healthy.sends)
	}
}
