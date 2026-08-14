package outreach

import (
	"strings"
	"testing"
)

func TestEligibleLeadQueryUsesStrictLifecycleDueFollowupGateAndSharedEmailLimit(t *testing.T) {
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
	if !strings.Contains(gate, "existing_campaign.next_send_at <= now()") {
		t.Fatal("only due follow-ups may block new-recipient delivery")
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
	if strings.Count(eligibleLeadsBaseQuery, ") <= 3") < 2 {
		t.Fatal("eligible query must reject shared emails for selected and blocking recipients")
	}
}

func TestGreetingOwnerPrecedenceRemainsApolloFirstNameThenApolloNameThenOwners(t *testing.T) {
	wantOrder := []string{
		"apollo_lead #>> '{contact,first_name}'",
		"apollo_lead #>> '{contact,name}'",
		"profile.owners ->> 0",
	}
	last := -1
	for _, fragment := range wantOrder {
		position := strings.Index(ownerFirstNameSelectExpression, fragment)
		if position <= last {
			t.Fatalf("owner precedence expression %q does not preserve order %#v", ownerFirstNameSelectExpression, wantOrder)
		}
		last = position
	}
}

func TestSharedEmailGroupsQueryListsOnlyRepeatedValidEmails(t *testing.T) {
	for _, required := range []string{
		"GROUP BY lower(trim(email))",
		"HAVING count(*) > 1",
		"ORDER BY restaurant_count DESC",
	} {
		if !strings.Contains(sharedEmailGroupsQuery, required) {
			t.Fatalf("shared email groups query missing %q", required)
		}
	}
}
