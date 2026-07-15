package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/leadprep"
)

const LeadPrepareJobType = "lead.prepare"

type LeadPreparePayload struct {
	RestaurantID uuid.UUID `json:"restaurant_id"`
}

type LeadPreparer interface {
	Prepare(ctx context.Context, restaurantID uuid.UUID) (leadprep.Result, error)
}

func LeadPrepareHandler(preparer LeadPreparer, log *slog.Logger) Handler {
	if log == nil {
		log = slog.Default()
	}
	return func(ctx context.Context, job Job) error {
		if preparer == nil {
			return fmt.Errorf("lead preparer is not configured")
		}

		var payload LeadPreparePayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("decode lead preparation payload: %w", err)
		}
		if payload.RestaurantID == uuid.Nil {
			return fmt.Errorf("lead preparation restaurant_id is required")
		}

		result, err := preparer.Prepare(ctx, payload.RestaurantID)
		if err != nil {
			return err
		}
		log.InfoContext(ctx, "lead_drafts_prepared",
			"restaurant_id", result.RestaurantID,
			"demo_site_id", result.DemoSiteID,
			"campaign_id", result.CampaignID,
			"created", result.Created,
		)
		return nil
	}
}
