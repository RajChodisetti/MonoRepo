package outreach

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/campaigns"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
)

type DemoTokenResolver struct {
	Campaigns campaigns.Repository
	Demos     demos.Repository
}

func (resolver *DemoTokenResolver) Resolve(ctx context.Context, demoSiteID uuid.UUID) (string, error) {
	if resolver.Campaigns != nil {
		token, err := resolver.Campaigns.GetLatestDemoTokenByDemoSiteID(ctx, demoSiteID)
		if err != nil {
			return "", err
		}
		if token != "" {
			return token, nil
		}
	}

	token, err := demos.GenerateDemoToken()
	if err != nil {
		return "", fmt.Errorf("generate demo token: %w", err)
	}
	tokenHash, err := demos.HashDemoToken(token)
	if err != nil {
		return "", fmt.Errorf("hash demo token: %w", err)
	}
	if err := resolver.Demos.UpdateTokenHash(ctx, demoSiteID, tokenHash); err != nil {
		return "", fmt.Errorf("refresh demo token: %w", err)
	}
	return token, nil
}
