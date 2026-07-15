package jobs

import (
	"context"

	"github.com/google/uuid"
)

type CampaignEnqueuer struct {
	Queue Queue
}

func (enqueuer *CampaignEnqueuer) EnqueueSendStep(ctx context.Context, campaignID uuid.UUID, step int) error {
	job, err := NewEmailSendJob(campaignID, step)
	if err != nil {
		return err
	}
	_, err = enqueuer.Queue.Enqueue(ctx, job)
	return err
}

var _ interface {
	EnqueueSendStep(ctx context.Context, campaignID uuid.UUID, step int) error
} = (*CampaignEnqueuer)(nil)
