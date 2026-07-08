package email

import (
	"fmt"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func NewAccountPoolFromConfig(emailCfg config.EmailConfig, outreachCfg config.OutreachConfig) (*AccountPool, error) {
	accounts := outreachCfg.ZohoAccounts
	if len(accounts) == 0 {
		return nil, fmt.Errorf("OUTREACH_ZOHO_ACCOUNTS_JSON must contain at least one Zoho account for bulk outreach")
	}

	providers := make([]Provider, 0, len(accounts))
	for index, account := range accounts {
		provider, err := NewZoho(emailCfg, account)
		if err != nil {
			return nil, fmt.Errorf("outreach zoho account %d: %w", index+1, err)
		}
		providers = append(providers, provider)
	}

	maxTotal := outreachCfg.BulkMax
	if maxTotal <= 0 {
		maxTotal = len(providers) * outreachCfg.EmailsPerAccount
	}

	return NewAccountPool(providers, outreachCfg.EmailsPerAccount, maxTotal)
}
