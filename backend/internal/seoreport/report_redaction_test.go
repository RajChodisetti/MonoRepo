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
			Status:                   "complete",
			RadiusKM:                 10,
			Cuisine:                  "thai restaurant",
			ScoreKind:                "google_visibility",
			SampleSize:               12,
			CurrentScore:             72,
			CurrentPosition:          3,
			CurrentRestaurantLeading: false,
			Notice:                   "Two competitors have stronger visibility.",
			Rows: []CompetitorRow{{
				Name:    "Secret Rival",
				Reasons: []string{"More recent reviews"},
			}},
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
	if redacted.CompetitorScan.Status != "locked" || redacted.CompetitorScan.RadiusKM != 0 ||
		redacted.CompetitorScan.Cuisine != "" || redacted.CompetitorScan.ScoreKind != "" ||
		redacted.CompetitorScan.SampleSize != 0 || redacted.CompetitorScan.CurrentScore != 0 ||
		redacted.CompetitorScan.CurrentPosition != 0 || redacted.CompetitorScan.CurrentRestaurantLeading ||
		redacted.CompetitorScan.Notice != "" {
		t.Fatalf("competitor metadata or conclusion survived redaction: %#v", redacted.CompetitorScan)
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
	if len(report.CompetitorScan.Rows) != 1 || report.CompetitorScan.Rows[0].Reasons[0] != "More recent reviews" ||
		report.AISummary == "" || report.Metrics[0].Rationale == "" || len(report.Metrics[0].Evidence) != 1 {
		t.Fatal("redaction mutated the full report value")
	}

	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal locked report: %v", err)
	}
	for _, emptyCollection := range []string{`"competitors":[]`, `"issues":[]`} {
		if !strings.Contains(string(encoded), emptyCollection) {
			t.Fatalf("locked JSON must preserve array contracts for %s: %s", emptyCollection, encoded)
		}
	}
	for _, secret := range []string{
		"Secret Rival", "private-menu", "instagram.com/secret", "gated summary",
		"gated website review", "gated metric rationale", "gated evidence", "gated recommendation",
		"thai restaurant", "google_visibility", "Two competitors", "More recent reviews",
	} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("locked JSON exposed %q: %s", secret, encoded)
		}
	}
	for _, competitorField := range []string{
		`"radiusKm"`, `"cuisine"`, `"scoreKind"`, `"sampleSize"`, `"currentScore"`,
		`"currentPosition"`, `"currentRestaurantLeading"`, `"notice"`, `"rows"`,
	} {
		if strings.Contains(string(encoded), competitorField) {
			t.Fatalf("locked JSON retained competitor field %s: %s", competitorField, encoded)
		}
	}
}

func TestUnlockedReportSerializationRetainsFullEvidence(t *testing.T) {
	report := Report{
		FullReportLocked: false,
		CompetitorScan: CompetitorScan{
			Status:                   "complete",
			RadiusKM:                 10,
			ScoreKind:                "google_visibility",
			CurrentScore:             0,
			CurrentRestaurantLeading: false,
			Rows:                     []CompetitorRow{{Name: "Real Rival"}},
		},
		AISummary: "full analysis",
		Metrics:   []Metric{{Rationale: "why", Evidence: []string{"proof"}, Recommendation: "next"}},
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
	for _, zeroValue := range []string{`"currentScore":0`, `"currentRestaurantLeading":false`} {
		if !strings.Contains(string(encoded), zeroValue) {
			t.Fatalf("unlocked JSON omitted meaningful zero value %s: %s", zeroValue, encoded)
		}
	}
}
