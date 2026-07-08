package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

type outreachZohoAccountJSON struct {
	AccountID    string `json:"account_id"`
	FromEmail    string `json:"from_email"`
	Region       string `json:"region"`
	APIBaseURL   string `json:"api_base_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func loadOutreachConfig(parser *envParser) OutreachConfig {
	cfg := OutreachConfig{
		BulkMax:          parser.int("OUTREACH_BULK_MAX", 150),
		EmailsPerAccount: parser.int("OUTREACH_EMAILS_PER_ACCOUNT", 50),
		ZohoAccountsJSON: parser.string("OUTREACH_ZOHO_ACCOUNTS_JSON", ""),
	}

	raw := strings.TrimSpace(cfg.ZohoAccountsJSON)
	if raw == "" {
		return cfg
	}

	var entries []outreachZohoAccountJSON
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON must be valid JSON array: %w", err))
		return cfg
	}

	for index, entry := range entries {
		account := ZohoMailConfig{
			AccountID:    strings.TrimSpace(entry.AccountID),
			FromEmail:    strings.TrimSpace(entry.FromEmail),
			Region:       strings.TrimSpace(entry.Region),
			APIBaseURL:   strings.TrimSpace(entry.APIBaseURL),
			ClientID:     strings.TrimSpace(entry.ClientID),
			ClientSecret: strings.TrimSpace(entry.ClientSecret),
			RefreshToken: strings.TrimSpace(entry.RefreshToken),
		}
		if account.Region == "" {
			account.Region = "com"
		}
		if account.APIBaseURL == "" {
			account.APIBaseURL = "https://mail.zoho.com/api/accounts"
		}
		if account.AccountID == "" || account.ClientID == "" || account.ClientSecret == "" || account.RefreshToken == "" {
			parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON entry %d is missing required Zoho fields", index+1))
			continue
		}
		cfg.ZohoAccounts = append(cfg.ZohoAccounts, account)
	}

	return cfg
}
