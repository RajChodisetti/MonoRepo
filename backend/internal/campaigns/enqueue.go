package campaigns

import (
	"context"

	"github.com/google/uuid"
)

type SendJobEnqueuer interface {
	EnqueueSendStep(ctx context.Context, campaignID uuid.UUID, step int) error
}
