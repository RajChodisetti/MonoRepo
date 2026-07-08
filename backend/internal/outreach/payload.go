package outreach

import (
	"encoding/json"
	"fmt"
)

type BulkJobPayload struct {
	TriggeredBy string          `json:"triggered_by,omitempty"`
	Summary     BulkSendSummary `json:"summary,omitempty"`
}

func decodeBulkSummary(payload []byte) (BulkSendSummary, error) {
	if len(payload) == 0 {
		return BulkSendSummary{}, nil
	}
	var job BulkJobPayload
	if err := json.Unmarshal(payload, &job); err != nil {
		return BulkSendSummary{}, fmt.Errorf("decode bulk job payload: %w", err)
	}
	return job.Summary, nil
}

func encodeBulkSummary(summary BulkSendSummary, triggeredBy string) ([]byte, error) {
	payload, err := json.Marshal(BulkJobPayload{
		TriggeredBy: triggeredBy,
		Summary:     summary,
	})
	if err != nil {
		return nil, fmt.Errorf("encode bulk job payload: %w", err)
	}
	return payload, nil
}
