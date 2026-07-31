package seoreport

import (
	"strings"
	"testing"
)

func TestBuildReportPointBudgets(t *testing.T) {
	rating := 4.6
	count := 180
	in := ScoreInput{
		Place: PlaceDetails{
			PlaceID:          "abc",
			Name:             "Bistro Melbourne",
			Address:          "123 Collins St, Melbourne VIC",
			Phone:            "+61 3 9000 0000",
			Email:            "hello@bistro.test",
			Website:          "https://bistromelbourne.com.au/order",
			MapsURI:          "https://maps.google.com/?cid=1",
			Rating:           &rating,
			UserRatingCount:  &count,
			BusinessStatus:   "OPERATIONAL",
			Types:            []string{"italian_restaurant", "restaurant"},
			EditorialSummary: "Italian restaurant in Melbourne serving pasta.",
		},
		Reviews: []Review{
			{Text: "Amazing pasta menu", Rating: 5, RelativeTime: "2 days ago"},
			{Text: "Great dishes", Rating: 5, RelativeTime: "1 week ago"},
		},
		PhotoCount:    12,
		HasHours:      true,
		Delivery:      true,
		Takeout:       true,
		DeliveryKnown: true,
		TakeoutKnown:  true,
		Enrichment: Enrichment{
			MenuItemCount:  12,
			MenuImageCount: 3,
			HasHours:       true,
			Email:          "hello@bistro.test",
			Phone:          "+61 3 9000 0000",
		},
	}

	report := BuildReport(in)
	if report.OverallScore < 70 {
		t.Fatalf("expected strong overall score, got %d", report.OverallScore)
	}

	totals := 0
	for _, m := range report.Metrics {
		totals += m.Max
		if m.Score < 0 || m.Score > m.Max {
			t.Fatalf("metric %s score %d out of range 0..%d", m.Key, m.Score, m.Max)
		}
	}
	if totals != 100 {
		t.Fatalf("expected metric max sum 100, got %d", totals)
	}
	if report.AISummary != "" {
		t.Fatalf("BuildReport should leave AISummary empty for summarizer")
	}
	if !report.FullReportLocked {
		t.Fatal("expected full report locked")
	}
}

func TestBuildReportMissingSignals(t *testing.T) {
	in := ScoreInput{
		Place: PlaceDetails{
			PlaceID: "xyz",
			Name:    "Quiet Cafe",
			Address: "Somewhere",
		},
	}
	report := BuildReport(in)
	if report.OverallScore < 8 || report.OverallScore > 40 {
		t.Fatalf("expected weak score, got %d", report.OverallScore)
	}
	if len(report.Issues) == 0 {
		t.Fatal("expected issues for weak listing")
	}
	summary := buildDeterministicSummary(in.Place, report, nil)
	lines := nonEmptyLines(summary)
	if len(lines) < 3 || len(lines) > 4 {
		t.Fatalf("expected 3-4 summary lines, got %d: %q", len(lines), summary)
	}
}

func TestScoreWebsiteSocialPenalty(t *testing.T) {
	social := scoreWebsite("https://instagram.com/myrestaurant")
	dedicated := scoreWebsite("https://myrestaurant.com.au")
	if social >= dedicated {
		t.Fatalf("social website should score lower than dedicated site (%d >= %d)", social, dedicated)
	}
}

func nonEmptyLines(s string) []string {
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
