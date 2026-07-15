package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const OutreachBulkSendJobType = "outreach.bulk_send"

type OutreachBulkSendPayload struct {
	TriggeredBy uuid.UUID `json:"triggered_by"`
}

func NewOutreachBulkSendJob(triggeredBy uuid.UUID) (Job, error) {
	payload, err := json.Marshal(OutreachBulkSendPayload{
		TriggeredBy: triggeredBy,
	})
	if err != nil {
		return Job{}, err
	}
	return Job{
		Type:           OutreachBulkSendJobType,
		Payload:        payload,
		IdempotencyKey: fmt.Sprintf("outreach.bulk_send:manual:%s", uuid.New()),
		// Provider and delivery failures are terminal for this workflow. Automatic
		// retries could drain paced slots through a broken credential; normal
		// pacing continuation uses explicit durable deferral instead.
		MaxAttempts: 1,
	}, nil
}
