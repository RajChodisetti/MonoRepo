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
		BodyTextTemplate: "{{greeting}}\n\nLearn more: {{website_url}}\n\nOpt out: {{unsubscribe_url}}",
	}
}

func TestValidateSequenceStepsRequiresImmediateFirstEnabledStep(t *testing.T) {
	step := validTestStep()
	step.DelayHours = 72
	if err := validateSequenceSteps([]SequenceStep{step}); !errors.Is(err, ErrSequenceInvalid) {
		t.Fatalf("validateSequenceSteps() error = %v, want ErrSequenceInvalid", err)
	}
}

func TestValidateAndRenderPlainTextWithExactlyTwoManagedLinks(t *testing.T) {
	step := validTestStep()
	if err := validateSequenceSteps([]SequenceStep{step}); err != nil {
		t.Fatalf("validateSequenceSteps() error = %v", err)
	}
	rendered, err := renderSequenceStep(step, "Harbour Cafe", "Ava", "https://api.tuvisolutions.com/t/unsubscribe/token")
	if err != nil {
		t.Fatalf("renderSequenceStep() error = %v", err)
	}
	if rendered.URLCount != 2 || strings.Count(rendered.BodyText, "https://") != 2 {
		t.Fatalf("rendered = %#v, want exactly two links", rendered)
	}
	if !strings.HasPrefix(rendered.BodyText, "Hi Ava,") || strings.Contains(rendered.BodyText, "<") {
		t.Fatalf("body = %q, want owner greeting and plain text", rendered.BodyText)
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

func TestValidateSequenceTemplateRejectsAdditionalURL(t *testing.T) {
	step := validTestStep()
	step.BodyTextTemplate += "\nhttps://example.com"
	if err := validateSequenceTemplate(step); !errors.Is(err, ErrSequenceInvalid) {
		t.Fatalf("validateSequenceTemplate() error = %v, want ErrSequenceInvalid", err)
	}
}

func TestValidateSequenceTemplateRejectsBareAutolinkDomain(t *testing.T) {
	step := validTestStep()
	step.BodyTextTemplate += "\nLearn more at example.com"
	if err := validateSequenceTemplate(step); !errors.Is(err, ErrSequenceInvalid) {
		t.Fatalf("validateSequenceTemplate() error = %v, want ErrSequenceInvalid", err)
	}
}

func TestValidateSequenceTemplateRejectsEmailAutolink(t *testing.T) {
	step := validTestStep()
	step.BodyTextTemplate += "\nReply to owner@example.com"
	if err := validateSequenceTemplate(step); !errors.Is(err, ErrSequenceInvalid) {
		t.Fatalf("validateSequenceTemplate() error = %v, want ErrSequenceInvalid", err)
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
