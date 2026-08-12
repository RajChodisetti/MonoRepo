package outreach

import (
	"strings"
	"testing"
)

func TestEligibleLeadQueryUsesStrictLifecycleAndFollowupPhaseGate(t *testing.T) {
	if strings.Contains(eligibleLeadsBaseQuery, "demo_ready") {
		t.Fatal("eligible query still accepts demo_ready")
	}
	if strings.Count(eligibleLeadsBaseQuery, "status IN ('lead', 'emailed')") < 2 {
		t.Fatal("eligible query does not enforce lead/emailed for both selected and blocking recipients")
	}
	gateAt := strings.Index(eligibleLeadsBaseQuery, "existing_campaign.current_step > 0")
	if gateAt < 0 {
		t.Fatal("eligible query has no unfinished-followup phase gate")
	}
	gate := eligibleLeadsBaseQuery[gateAt:]
	if strings.Contains(gate, "existing_campaign.next_send_at <= now()") {
		t.Fatal("future follow-ups must block new-recipient delivery")
	}
	for _, required := range []string{
		"outreach_consent_basis = 'inferred_business'",
		"outreach_consent_evidence <> '{}'::jsonb",
		"shown_interest = false",
	} {
		if !strings.Contains(eligibleLeadsBaseQuery, required) {
			t.Fatalf("eligible query missing policy guard %q", required)
		}
	}
}
