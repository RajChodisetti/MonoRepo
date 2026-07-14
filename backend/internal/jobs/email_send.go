package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const EmailSendJobType = "email.send"

type EmailSendPayload struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	Step       int       `json:"step"`
}

func NewEmailSendJob(campaignID uuid.UUID, step int) (Job, error) {
	payload, err := json.Marshal(EmailSendPayload{
		CampaignID: campaignID,
		Step:       step,
	})
	if err != nil {
		return Job{}, err
	}
	return Job{
		Type:           EmailSendJobType,
		Payload:        payload,
		IdempotencyKey: fmt.Sprintf("email.send:%s:step:%d", campaignID, step),
		// Providers do not currently expose a cross-provider idempotency contract.
		// Retrying after a successful send and failed state update could duplicate outreach.
		MaxAttempts: 1,
	}, nil
}
