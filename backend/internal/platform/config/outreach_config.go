package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type outreachZohoAccountJSON struct {
	Key          string `json:"key"`
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
		EmailsPerAccount: parser.int("OUTREACH_EMAILS_PER_ACCOUNT", 40),
		SendInterval:     parser.duration("OUTREACH_SEND_INTERVAL", 2*time.Second),
		AccountCooldown:  parser.duration("OUTREACH_EMAIL_COOLDOWN", 24*time.Hour),
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

	keys := make(map[string]struct{}, len(entries))
	identities := make(map[string]struct{}, len(entries))
	for index, entry := range entries {
		accountKey := strings.TrimSpace(entry.Key)
		if accountKey == "" {
			accountKey = strings.TrimSpace(entry.AccountID)
		}
		account := ZohoMailConfig{
			AccountKey:   accountKey,
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
		if account.AccountKey == "" || account.AccountID == "" || account.ClientID == "" || account.ClientSecret == "" || account.RefreshToken == "" {
			parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON entry %d is missing required Zoho fields", index+1))
			continue
		}
		if _, exists := keys[account.AccountKey]; exists {
			parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON entry %d has duplicate key %q", index+1, account.AccountKey))
			continue
		}
		identity := strings.ToLower(account.Region) + "|" + strings.ToLower(account.AccountID)
		if _, exists := identities[identity]; exists {
			parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON entry %d duplicates Zoho account %q in region %q", index+1, account.AccountID, account.Region))
			continue
		}
		keys[account.AccountKey] = struct{}{}
		identities[identity] = struct{}{}
		cfg.ZohoAccounts = append(cfg.ZohoAccounts, account)
	}

	return cfg
}
