package jobs

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type EmailSendDeps struct {
	Campaigns campaigns.Repository
	CampaignsService *campaigns.Service
	Email     emailprovider.Provider
	EmailCfg  config.EmailConfig
	AppURLs   config.AppURLsConfig
}

func EmailSendHandler(deps EmailSendDeps, log *slog.Logger) Handler {
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

		sendCtx, err := deps.Campaigns.GetSendContext(ctx, payload.CampaignID)
		if err != nil {
			return err
		}

		suppressed, err := deps.Campaigns.IsSuppressed(ctx, sendCtx.RestaurantEmail)
		if err != nil {
			return err
		}
		if err := campaigns.CheckEligibility(campaigns.EligibilityInput{
			RestaurantEmail: sendCtx.RestaurantEmail,
			ReviewStatus:    sendCtx.ReviewStatus,
			DemoStatus:      sendCtx.DemoStatus,
			CampaignStatus:  campaign.Status,
			Suppressed:      suppressed,
		}); err != nil {
			return err
		}

		trackingURLs, err := deps.CampaignsService.BuildTrackingURLs(ctx, campaign, sendCtx)
		if err != nil {
			return err
		}

		draft := campaigns.InjectTracking(campaigns.DraftContent{
			Subject:  campaign.Subject,
			BodyHTML: campaign.BodyHTML,
			BodyText: campaign.BodyText,
		}, trackingURLs, deps.EmailCfg.OpenTrackingEnabled)

		result, err := deps.Email.Send(ctx, emailprovider.SendRequest{
			To:       sendCtx.RestaurantEmail,
			Subject:  draft.Subject,
			HTMLBody: draft.BodyHTML,
			TextBody: draft.BodyText,
			Metadata: map[string]string{
				"campaign_id": campaign.ID.String(),
				"restaurant_id": campaign.RestaurantID.String(),
				"original_to": sendCtx.RestaurantEmail,
			},
		})
		if err != nil {
			meta, _ := json.Marshal(map[string]string{"error": err.Error()})
			_ = deps.Campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventFailed, meta)
			return err
		}

		eventMeta, _ := json.Marshal(map[string]any{
			"step":              payload.Step,
			"provider_message":  result.ProviderMessageID,
			"redirected_to":     result.RedirectedTo,
			"skipped":           result.Skipped,
			"original_recipient": sendCtx.RestaurantEmail,
		})
		if err := deps.Campaigns.InsertEvent(ctx, campaign.ID, campaign.RestaurantID, campaigns.EventSent, eventMeta); err != nil {
			return err
		}

		if _, err := deps.Campaigns.MarkSent(ctx, campaign.ID, payload.Step); err != nil {
			return err
		}
		if err := deps.Campaigns.MarkRestaurantEmailed(ctx, campaign.RestaurantID); err != nil {
			return err
		}

		log.InfoContext(ctx, "email_send_completed",
			"job_id", job.ID,
			"campaign_id", campaign.ID,
			"to", sendCtx.RestaurantEmail,
			"redirected", strings.TrimSpace(result.RedirectedTo) != "",
		)
		return nil
	}
}
