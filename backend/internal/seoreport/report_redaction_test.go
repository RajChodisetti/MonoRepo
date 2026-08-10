package seoreport

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactLockedReportRemovesGatedServerData(t *testing.T) {
	report := Report{
		FullReportLocked: true,
		Competitors:      []CompetitorRow{{Name: "Secret Rival"}},
		CompetitorScan: CompetitorScan{
			Status: "complete",
			Rows:   []CompetitorRow{{Name: "Secret Rival"}},
		},
		MenuEvidence: MenuEvidence{
			Status:    "present",
			MenuURL:   "https://restaurant.example/private-menu",
			Rationale: "gated menu rationale",
		},
		SocialPresence: SocialPresence{
			Status:    "present",
			Profiles:  []SocialProfile{{Platform: "Instagram", URL: "https://instagram.com/secret"}},
			Rationale: "gated social rationale",
		},
		Metrics: []Metric{{
			Key:            "menu",
			Rationale:      "gated metric rationale",
			Evidence:       []string{"gated evidence"},
			Recommendation: "gated recommendation",
		}},
		Issues:        []Issue{{Title: "gated issue"}},
		AISummary:     "gated summary",
		WebsiteReview: "gated website review",
	}

	redacted := RedactLockedReport(report)
	if len(redacted.Competitors) != 0 || len(redacted.CompetitorScan.Rows) != 0 {
		t.Fatalf("competitor identities survived redaction: %#v", redacted.CompetitorScan.Rows)
	}
	if redacted.AISummary != "" || redacted.WebsiteReview != "" || len(redacted.Issues) != 0 {
		t.Fatalf("gated narrative survived redaction: %#v", redacted)
	}
	if redacted.MenuEvidence.MenuURL != "" || redacted.MenuEvidence.Rationale != "" ||
		len(redacted.SocialPresence.Profiles) != 0 || redacted.SocialPresence.Rationale != "" {
		t.Fatalf("gated evidence survived redaction: %#v %#v", redacted.MenuEvidence, redacted.SocialPresence)
	}
	if redacted.Metrics[0].Rationale != "" || len(redacted.Metrics[0].Evidence) != 0 || redacted.Metrics[0].Recommendation != "" {
		t.Fatalf("metric explanation survived redaction: %#v", redacted.Metrics[0])
	}
	if len(report.CompetitorScan.Rows) != 1 || report.AISummary == "" {
		t.Fatal("redaction mutated the full report value")
	}

	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal locked report: %v", err)
	}
	for _, secret := range []string{
		"Secret Rival", "private-menu", "instagram.com/secret", "gated summary",
		"gated website review", "gated metric rationale", "gated evidence", "gated recommendation",
	} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("locked JSON exposed %q: %s", secret, encoded)
		}
	}
}

func TestUnlockedReportSerializationRetainsFullEvidence(t *testing.T) {
	report := Report{
		FullReportLocked: false,
		CompetitorScan:   CompetitorScan{Rows: []CompetitorRow{{Name: "Real Rival"}}},
		AISummary:        "full analysis",
		Metrics:          []Metric{{Rationale: "why", Evidence: []string{"proof"}, Recommendation: "next"}},
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal unlocked report: %v", err)
	}
	for _, expected := range []string{"Real Rival", "full analysis", "why", "proof", "next"} {
		if !strings.Contains(string(encoded), expected) {
			t.Fatalf("unlocked JSON missing %q: %s", expected, encoded)
		}
	}
}
