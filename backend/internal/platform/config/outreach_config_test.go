package config_test

import (
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestLoadOutreachZohoAccountsJSON(t *testing.T) {
	t.Setenv("OUTREACH_ZOHO_ACCOUNTS_JSON", `[{"account_id":"acc1","from_email":"a@example.com","client_id":"cid","client_secret":"sec","refresh_token":"rt","region":"com"}]`)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Outreach.ZohoAccounts) != 1 {
		t.Fatalf("ZohoAccounts len = %d, want 1", len(cfg.Outreach.ZohoAccounts))
	}
	if cfg.Outreach.ZohoAccounts[0].AccountID != "acc1" {
		t.Fatalf("AccountID = %q, want acc1", cfg.Outreach.ZohoAccounts[0].AccountID)
	}
}

func TestLoadOutreachGoogleWorkspaceAccountsJSON(t *testing.T) {
	t.Setenv("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON", `[{
		"key":"sales-au-1",
		"mailbox_email":"sales1@example.com",
		"client_id":"client-id",
		"client_secret":"client-secret",
		"refresh_token":"refresh-token"
	}]`)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Outreach.GoogleWorkspaceAccounts) != 1 {
		t.Fatalf("GoogleWorkspaceAccounts len = %d, want 1", len(cfg.Outreach.GoogleWorkspaceAccounts))
	}
	account := cfg.Outreach.GoogleWorkspaceAccounts[0]
	if account.AccountKey != "sales-au-1" {
		t.Fatalf("AccountKey = %q, want sales-au-1", account.AccountKey)
	}
	if account.MailboxEmail != "sales1@example.com" {
		t.Fatalf("MailboxEmail = %q, want sales1@example.com", account.MailboxEmail)
	}
	if account.FromEmail != account.MailboxEmail {
		t.Fatalf("FromEmail = %q, want mailbox fallback %q", account.FromEmail, account.MailboxEmail)
	}
}

func TestLoadOutreachGoogleWorkspaceAccountsCanonicalizesMailbox(t *testing.T) {
	t.Setenv("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON", `[{
		"mailbox_email":"<Sales1@Example.com>",
		"client_id":"client-id",
		"client_secret":"client-secret",
		"refresh_token":"refresh-token"
	}]`)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	account := cfg.Outreach.GoogleWorkspaceAccounts[0]
	if account.MailboxEmail != "sales1@example.com" || account.AccountKey != "sales1@example.com" {
		t.Fatalf("canonical mailbox/key = %q/%q, want sales1@example.com", account.MailboxEmail, account.AccountKey)
	}
}

func TestLoadOutreachGoogleWorkspaceAccountsRejectsInvalidMailbox(t *testing.T) {
	t.Setenv("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON", `[{
		"mailbox_email":"Sales Team <sales1@example.com>",
		"client_id":"client-id",
		"client_secret":"client-secret",
		"refresh_token":"refresh-token"
	}]`)

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "mailbox_email") {
		t.Fatalf("Load() error = %v, want mailbox_email validation error", err)
	}
}
