package outreach_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
)

func TestPreviewSequenceIncludesSelectedSavedSignature(t *testing.T) {
	sequenceID := uuid.New()
	wantSignature := outreach.SequenceSignature{
		Name:              "Alex Morgan",
		Title:             "Partnerships Manager",
		AdditionalDetails: "Phone: +61 400 000 000",
	}
	repo := &mockRepo{
		sequenceSteps: []outreach.SequenceStep{{
			ID: uuid.New(), SequenceID: sequenceID, Position: 1, Enabled: true,
			SubjectTemplate:  "A practical idea for {{restaurant_name}}",
			BodyTextTemplate: "[GREETING]\n\nA short saved message.",
		}},
		signatures: map[uuid.UUID]outreach.SequenceSignature{
			sequenceID: wantSignature,
		},
	}
	service := newSequenceService(t, repo, &mockEmailProvider{})

	preview, err := service.PreviewSequence(
		context.Background(),
		internalAdminPrincipal(),
		sequenceID,
		outreach.PreviewSequenceInput{RestaurantName: "Signature Cafe", OwnerFirstName: "Casey"},
	)
	if err != nil {
		t.Fatalf("PreviewSequence() error = %v", err)
	}
	if preview.Signature != wantSignature {
		t.Fatalf("PreviewSequence() signature = %#v, want %#v", preview.Signature, wantSignature)
	}

	payload, err := json.Marshal(preview)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var contract struct {
		Signature outreach.SequenceSignature `json:"signature"`
	}
	if err := json.Unmarshal(payload, &contract); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if contract.Signature != wantSignature {
		t.Fatalf("preview JSON signature = %#v, want %#v", contract.Signature, wantSignature)
	}
}
