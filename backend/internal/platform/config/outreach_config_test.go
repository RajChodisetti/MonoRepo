package config_test

import (
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
