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
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 52, // strict visual band → modest website metric points
		WebsiteReview:       "Clean layout but weak order CTA and menu prominence.",
		WebsiteScreenshot:   "data:image/jpeg;base64,test",
	}

	report := BuildReport(in)
	// Strict scoring: even a relatively complete listing should land ~20–60, not 75+.
	if report.OverallScore < 28 || report.OverallScore > 60 {
		t.Fatalf("expected strict overall in 28–60, got %d", report.OverallScore)
	}
	if report.WebsiteQualityScore != 52 || report.WebsiteScreenshot == "" {
		t.Fatalf("expected website audit fields on report")
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
	if report.OverallScore < 12 || report.OverallScore > 35 {
		t.Fatalf("expected weak strict score, got %d", report.OverallScore)
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
	social := scoreWebsite("https://instagram.com/myrestaurant", true, 40)
	dedicated := scoreWebsite("https://myrestaurant.com.au", true, 40)
	if social >= dedicated {
		t.Fatalf("social website should score lower than dedicated site (%d >= %d)", social, dedicated)
	}
}

func TestScoreWebsiteNoFreeFullMarks(t *testing.T) {
	// Presence alone (no visual audit) must not award the old free ~20 points.
	presenceOnly := scoreWebsite("https://myrestaurant.com.au", false, 0)
	if presenceOnly >= 12 {
		t.Fatalf("presence-only website score too high: %d", presenceOnly)
	}
	// Strict mid quality (typical 20–60 band) maps into modest metric points.
	mid := scoreWebsite("https://myrestaurant.com.au", true, 45)
	if mid < 4 || mid > 10 {
		t.Fatalf("expected mid visual quality to map ~4–10/20, got %d", mid)
	}
	// Even a high AI score is hard-capped well below a perfect 20.
	high := scoreWebsite("https://myrestaurant.com.au", true, 90)
	if high > 11 {
		t.Fatalf("expected hard cap at 11/20, got %d", high)
	}
	if scoreWebsite("", true, 90) != 0 {
		t.Fatal("empty website must score 0")
	}
}

func TestStrictOverallScoreBand(t *testing.T) {
	if got := strictOverallScore(90); got > 60 {
		t.Fatalf("raw 90 should compress to <=60, got %d", got)
	}
	if got := strictOverallScore(75); got > 55 {
		t.Fatalf("raw 75 should compress well below 75, got %d", got)
	}
	if got := strictOverallScore(40); got < 20 || got > 45 {
		t.Fatalf("raw 40 should land mid-low band, got %d", got)
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
