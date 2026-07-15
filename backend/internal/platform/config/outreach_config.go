package config

import (
	"encoding/json"
	"fmt"
	"net/mail"
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

type outreachGoogleWorkspaceAccountJSON struct {
	Key          string `json:"key"`
	MailboxEmail string `json:"mailbox_email"`
	FromEmail    string `json:"from_email"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func loadOutreachConfig(parser *envParser) OutreachConfig {
	cfg := OutreachConfig{
		BulkMax:                     parser.int("OUTREACH_BULK_MAX", 150),
		EmailsPerAccount:            parser.int("OUTREACH_EMAILS_PER_ACCOUNT", 40),
		SendInterval:                parser.duration("OUTREACH_SEND_INTERVAL", 2*time.Second),
		AccountCooldown:             parser.duration("OUTREACH_EMAIL_COOLDOWN", 24*time.Hour),
		ZohoAccountsJSON:            parser.string("OUTREACH_ZOHO_ACCOUNTS_JSON", ""),
		GoogleWorkspaceAccountsJSON: parser.string("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON", ""),
	}

	keys := make(map[string]struct{})
	loadOutreachZohoAccounts(parser, &cfg, keys)
	loadOutreachGoogleWorkspaceAccounts(parser, &cfg, keys)
	return cfg
}

func loadOutreachZohoAccounts(parser *envParser, cfg *OutreachConfig, keys map[string]struct{}) {
	raw := strings.TrimSpace(cfg.ZohoAccountsJSON)
	if raw == "" {
		return
	}

	var entries []outreachZohoAccountJSON
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		parser.addError(fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON must be valid JSON array: %w", err))
		return
	}

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
}

func loadOutreachGoogleWorkspaceAccounts(parser *envParser, cfg *OutreachConfig, keys map[string]struct{}) {
	raw := strings.TrimSpace(cfg.GoogleWorkspaceAccountsJSON)
	if raw == "" {
		return
	}

	var entries []outreachGoogleWorkspaceAccountJSON
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON must be valid JSON array: %w", err))
		return
	}

	identities := make(map[string]struct{}, len(entries))
	for index, entry := range entries {
		mailboxEmail, err := canonicalOutreachMailbox(entry.MailboxEmail)
		if err != nil {
			parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON entry %d mailbox_email: %w", index+1, err))
			continue
		}
		fromEmail := mailboxEmail
		if strings.TrimSpace(entry.FromEmail) != "" {
			fromEmail, err = canonicalOutreachMailbox(entry.FromEmail)
			if err != nil {
				parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON entry %d from_email: %w", index+1, err))
				continue
			}
		}
		if fromEmail == "" {
			fromEmail = mailboxEmail
		}
		accountKey := strings.TrimSpace(entry.Key)
		if accountKey == "" {
			accountKey = mailboxEmail
		}
		account := GmailMailConfig{
			AccountKey:   accountKey,
			MailboxEmail: mailboxEmail,
			FromEmail:    fromEmail,
			ClientID:     strings.TrimSpace(entry.ClientID),
			ClientSecret: strings.TrimSpace(entry.ClientSecret),
			RefreshToken: strings.TrimSpace(entry.RefreshToken),
		}
		if account.AccountKey == "" || account.MailboxEmail == "" || account.FromEmail == "" || account.ClientID == "" || account.ClientSecret == "" || account.RefreshToken == "" {
			parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON entry %d is missing required Google Workspace fields", index+1))
			continue
		}
		if _, exists := keys[account.AccountKey]; exists {
			parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON entry %d has duplicate key %q", index+1, account.AccountKey))
			continue
		}
		if _, exists := identities[account.MailboxEmail]; exists {
			parser.addError(fmt.Errorf("OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON entry %d duplicates Google Workspace mailbox %q", index+1, account.MailboxEmail))
			continue
		}
		keys[account.AccountKey] = struct{}{}
		identities[account.MailboxEmail] = struct{}{}
		cfg.GoogleWorkspaceAccounts = append(cfg.GoogleWorkspaceAccounts, account)
	}
}

func canonicalOutreachMailbox(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("email address is required")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || strings.TrimSpace(address.Address) == "" || address.Name != "" {
		return "", fmt.Errorf("email address is invalid")
	}
	return strings.ToLower(strings.TrimSpace(address.Address)), nil
}
