package outreach

import (
	"errors"
	"strings"
	"testing"
)

func validTestStep() SequenceStep {
	return SequenceStep{
		Position: 1, Enabled: true, DelayHours: 0,
		SubjectTemplate:  "A practical idea for {{restaurant_name}}",
		BodyTextTemplate: "{{greeting}}\n\nI had one practical idea for {{restaurant_name}}. Open to a quick note back?\n\nUnsubscribe: {{unsubscribe_url}}",
	}
}

func noOptOutTestStep() SequenceStep {
	step := validTestStep()
	step.BodyTextTemplate = "{{greeting}}\n\nI had one practical idea for {{restaurant_name}}. Open to a quick note back?"
	return step
}

func TestValidateSequenceStepsRequiresImmediateFirstEnabledStep(t *testing.T) {
	step := validTestStep()
	step.DelayHours = 72
	if err := validateSequenceSteps([]SequenceStep{step}); !errors.Is(err, ErrSequenceInvalid) {
		t.Fatalf("validateSequenceSteps() error = %v, want ErrSequenceInvalid", err)
	}
}

func TestValidateAndRenderPlainTextWithOptOutLink(t *testing.T) {
	step := validTestStep()
	if err := validateSequenceSteps([]SequenceStep{step}); err != nil {
		t.Fatalf("validateSequenceSteps() error = %v", err)
	}
	rendered, err := renderSequenceStep(step, "Harbour Cafe", "Ava", "https://api.tuvisolutions.com/t/unsubscribe/token")
	if err != nil {
		t.Fatalf("renderSequenceStep() error = %v", err)
	}
	if !strings.Contains(rendered.BodyText, "https://api.tuvisolutions.com/t/unsubscribe/token") {
		t.Fatalf("rendered = %#v, want opt-out link", rendered)
	}
	if !strings.HasPrefix(rendered.BodyText, "Hi Ava,") || strings.Contains(rendered.BodyText, "<") {
		t.Fatalf("body = %q, want owner greeting and plain text", rendered.BodyText)
	}
}

func TestRenderAppendsOptOutWhenTemplateOmitsPlaceholder(t *testing.T) {
	rendered, err := renderSequenceStep(
		noOptOutTestStep(), "Harbour Cafe", "Ava", "https://api.tuvisolutions.com/t/unsubscribe/token",
	)
	if err != nil {
		t.Fatalf("renderSequenceStep() error = %v", err)
	}
	if !strings.Contains(rendered.BodyText, "\n\nUnsubscribe: https://api.tuvisolutions.com/t/unsubscribe/token") {
		t.Fatalf("body = %q, want appended opt-out footer", rendered.BodyText)
	}
}

func TestValidateSequenceTemplateAllowsAdminManagedLinks(t *testing.T) {
	step := validTestStep()
	step.BodyTextTemplate += "\n\nTuvi overview: {{website_url}}\nhttps://example.com"
	if err := validateSequenceTemplate(step); err != nil {
		t.Fatalf("validateSequenceTemplate() error = %v", err)
	}
}

func TestRenderFallsBackToRestaurantGreeting(t *testing.T) {
	rendered, err := renderSequenceStep(
		validTestStep(), "Harbour Cafe", "", "https://api.tuvisolutions.com/t/unsubscribe/token",
	)
	if err != nil {
		t.Fatalf("renderSequenceStep() error = %v", err)
	}
	if !strings.HasPrefix(rendered.BodyText, "Hi Harbour Cafe team,") {
		t.Fatalf("body = %q, want restaurant fallback greeting", rendered.BodyText)
	}
}

func TestSequenceApprovalRebasesOnlyUntouchedEnrollments(t *testing.T) {
	for _, required := range []string{
		"campaign.current_step = 0",
		"campaign.last_sent_at IS NULL",
		"FROM email_delivery_attempts",
		"FROM email_events",
		"event.event_type IN ('sent', 'skipped', 'failed')",
	} {
		if !strings.Contains(rebaseUntouchedEnrollmentsQuery, required) {
			t.Fatalf("rebase query missing immutable-history guard %q", required)
		}
	}
}
