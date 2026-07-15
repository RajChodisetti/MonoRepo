package jobs

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type EmailSendDeps struct {
	Campaigns        campaigns.Repository
	CampaignsService *campaigns.Service
	Email            emailprovider.Provider
	EmailCfg         config.EmailConfig
	AppURLs          config.AppURLsConfig
}

func EmailSendHandler(deps EmailSendDeps, _ *slog.Logger) Handler {
	return func(ctx context.Context, job Job) error {
		var payload EmailSendPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}

		campaign, err := deps.Campaigns.GetByID(ctx, payload.CampaignID)
		if err != nil {
			return err
		}
		if campaign.Status == campaigns.StatusStopped {
			return nil
		}
		if campaign.CampaignType != campaigns.TypeOutreach {
			return campaigns.ErrUnsupportedType
		}
		// Never mutate a `sending` outreach campaign here. It may be owned by
		// a live quota-managed delivery attempt. Legacy email.send jobs are
		// failed by migration 000020 and any stragglers fail closed here.
		return campaigns.ErrOutreachRequiresBulk
	}
}
