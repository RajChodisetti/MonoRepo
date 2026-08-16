package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
)

func TestDecodeOutreachSequenceUpdatePreservesSignature(t *testing.T) {
	request := httptest.NewRequest("PUT", "/api/v1/outreach/sequences/sequence-id", strings.NewReader(`{
		"name":"Saved template",
		"expected_updated_at":"2026-08-16T12:34:56Z",
		"signature":{
			"name":"Alex Morgan",
			"title":"Partnerships Manager",
			"additional_details":"Phone: +61 400 000 000"
		},
		"steps":[{
			"position":1,
			"enabled":true,
			"delay_hours":0,
			"subject_template":"Hello",
			"body_text_template":"Saved body"
		}]
	}`))
	var input outreach.UpdateSequenceInput
	if err := decodeOutreachJSON(request, &input); err != nil {
		t.Fatalf("decodeOutreachJSON() error = %v", err)
	}
	want := outreach.SequenceSignature{
		Name:              "Alex Morgan",
		Title:             "Partnerships Manager",
		AdditionalDetails: "Phone: +61 400 000 000",
	}
	if input.Signature != want {
		t.Fatalf("decoded signature = %#v, want %#v", input.Signature, want)
	}
	if input.ExpectedUpdatedAt.IsZero() {
		t.Fatal("expected_updated_at concurrency token was not decoded")
	}
	if len(input.Steps) != 1 || input.Steps[0].BodyTextTemplate != "Saved body" {
		t.Fatalf("decoded steps = %#v", input.Steps)
	}
}
