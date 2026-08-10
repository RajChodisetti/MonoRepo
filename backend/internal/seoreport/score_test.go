package seoreport

import (
	"strings"
	"testing"
)

func TestBuildReportPointBudgets(t *testing.T) {
	rating := 5.0
	count := 500
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
			PrimaryType:      "italian_restaurant",
			EditorialSummary: "Italian restaurant in Melbourne serving pasta.",
		},
		Reviews: []Review{
			{Text: "Amazing pasta menu", Rating: 5, RelativeTime: "2 days ago"},
			{Text: "Great dishes", Rating: 5, RelativeTime: "1 week ago"},
		},
		PhotoCount:      12,
		HasHours:        true,
		Delivery:        true,
		Takeout:         true,
		Reservable:      true,
		DeliveryKnown:   true,
		TakeoutKnown:    true,
		ReservableKnown: true,
		Enrichment: Enrichment{
			MenuItemCount:  12,
			MenuImageCount: 3,
			HasHours:       true,
			Email:          "hello@bistro.test",
			Phone:          "+61 3 9000 0000",
		},
		WebsiteReachable:        true,
		WebsiteQualityKnown:     true,
		WebsiteQualityScore:     100,
		WebsiteReview:           "Clean layout but weak order CTA and menu prominence.",
		WebsiteScreenshot:       "data:image/jpeg;base64,test",
		WebsiteMobileScreenshot: "data:image/jpeg;base64,mobile",
		MenuEvidence: MenuEvidence{
			Status:            "present",
			HasWebsiteLink:    true,
			HasStructuredData: true,
			MenuURL:           "https://bistromelbourne.com.au/menu",
			Rationale:         "Verified menu evidence.",
		},
		SocialPresence: SocialPresence{
			Status: "present",
			Profiles: []SocialProfile{
				{Platform: "Instagram", URL: "https://instagram.com/bistro"},
				{Platform: "Facebook", URL: "https://facebook.com/bistro"},
			},
		},
	}

	report := BuildReport(in)
	if report.OverallScore != 100 {
		t.Fatalf("expected complete report to reach 100, got %d", report.OverallScore)
	}
	if report.WebsiteQualityScore != 100 || report.WebsiteScreenshot == "" {
		t.Fatalf("expected website audit fields on report")
	}

	totals := 0
	for _, m := range report.Metrics {
		totals += m.Max
		if m.Score < 0 || m.Score > m.Max {
			t.Fatalf("metric %s score %d out of range 0..%d", m.Key, m.Score, m.Max)
		}
		if m.Score != m.Max {
			t.Fatalf("metric %s should reach declared max: %d/%d", m.Key, m.Score, m.Max)
		}
		if m.Rationale == "" || len(m.Evidence) == 0 || m.Recommendation == "" {
			t.Fatalf("metric %s missing report/PDF explanation fields: %#v", m.Key, m)
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
	if report.OverallScore < 0 || report.OverallScore > 20 {
		t.Fatalf("expected weak evidence score, got %d", report.OverallScore)
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

func TestScoreSEOGenericAddressAloneEarnsNoLocalityBonus(t *testing.T) {
	place := PlaceDetails{
		Name:    "Quiet Cafe",
		Address: "123 Main Street, Springfield NSW 2000, Australia",
	}
	if got := scoreSEO(place, ""); got != 0 {
		t.Fatalf("generic address-only SEO score=%d, want zero without corroborated locality content", got)
	}

	place.Name = "Quiet Cafe Springfield"
	if got := scoreSEO(place, ""); got != 6 {
		t.Fatalf("corroborated listing-name locality score=%d, want 6", got)
	}
}

func TestScoreWebsiteSocialPenalty(t *testing.T) {
	social := scoreWebsite("https://instagram.com/myrestaurant", true, true, 40)
	dedicated := scoreWebsite("https://myrestaurant.com.au", true, true, 40)
	if social >= dedicated {
		t.Fatalf("social website should score lower than dedicated site (%d >= %d)", social, dedicated)
	}
}

func TestScoreWebsiteNoFreeFullMarks(t *testing.T) {
	// Verified reachability alone (no visual audit) must not award visual points.
	reachabilityOnly := scoreWebsite("https://myrestaurant.com.au", true, false, 0)
	if reachabilityOnly >= 12 {
		t.Fatalf("reachability-only website score too high: %d", reachabilityOnly)
	}
	// Mid quality maps proportionally into the explicit 20-point budget.
	mid := scoreWebsite("https://myrestaurant.com.au", true, true, 45)
	if mid < 4 || mid > 10 {
		t.Fatalf("expected mid visual quality to map ~4–10/20, got %d", mid)
	}
	if got := scoreWebsite("https://myrestaurant.com.au", true, true, 100); got != 20 {
		t.Fatalf("expected website metric to reach 20/20, got %d", got)
	}
	if scoreWebsite("", true, true, 90) != 0 {
		t.Fatal("empty website must score 0")
	}
}

func TestScoreWebsiteListedButUnreachableGetsNoPoints(t *testing.T) {
	if got := scoreWebsite("https://dead.example", false, true, 100); got != 0 {
		t.Fatalf("unreachable listed website score=%d, want zero", got)
	}
	report := BuildReport(ScoreInput{
		Place:               PlaceDetails{Name: "Dead Link Cafe", Website: "https://dead.example"},
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 100,
	})
	for _, metric := range report.Metrics {
		if metric.Key != "website" {
			continue
		}
		if metric.Score != 0 || !strings.Contains(metric.Rationale, "could not be verified as reachable") {
			t.Fatalf("unreachable website metric=%#v", metric)
		}
		if len(metric.Evidence) < 2 || metric.Evidence[0] != "Official website: present" || metric.Evidence[1] != "Live homepage reachable: not found" {
			t.Fatalf("listed/reachable evidence=%#v", metric.Evidence)
		}
		return
	}
	t.Fatal("website metric missing")
}

func TestOverallScorePreservesHundredPointContract(t *testing.T) {
	for _, raw := range []int{0, 40, 75, 90, 100} {
		if got := strictOverallScore(raw); got != raw {
			t.Fatalf("strictOverallScore(%d)=%d, want explicit raw total", raw, got)
		}
	}
	if strictOverallScore(-1) != 0 || strictOverallScore(101) != 100 {
		t.Fatal("overall score must clamp outside 0..100")
	}
}

func TestGenericPlacesPhotosNeverCountAsMenu(t *testing.T) {
	report := BuildReport(ScoreInput{
		Place:      PlaceDetails{Name: "Photo Cafe", Website: "https://photo.example"},
		PhotoCount: 20,
		Reviews:    []Review{{Text: "Great menu and dishes", Rating: 5}},
		MenuEvidence: MenuEvidence{
			Status:    "not_found",
			Rationale: placesMenuEvidenceLimitation,
		},
	})
	for _, metric := range report.Metrics {
		if metric.Key == "menu" && metric.Score != 0 {
			t.Fatalf("generic photos/review words produced menu score %d", metric.Score)
		}
	}
}

func TestSocialPresenceContributesExactlyThreeListingPoints(t *testing.T) {
	place := PlaceDetails{BusinessStatus: "CLOSED_TEMPORARILY"}
	without := scoreListing(place, false, 0, SocialPresence{})
	presence := scoreSocialPresence(SocialPresence{
		Status: "present",
		Profiles: []SocialProfile{
			{Platform: "Instagram", URL: "https://instagram.com/one"},
			{Platform: "Facebook", URL: "https://facebook.com/one"},
		},
	})
	with := scoreListing(place, false, 0, presence)
	if presence.Score != 3 || with-without != 3 {
		t.Fatalf("social score=%d listing delta=%d, want 3", presence.Score, with-without)
	}
}

func TestLinkAggregatorNeverEarnsDedicatedWebsiteOrOrderPoints(t *testing.T) {
	report := BuildReport(ScoreInput{
		Place:               PlaceDetails{Name: "Aggregator Cafe", Website: "https://linktr.ee/order-online"},
		WebsiteReachable:    true,
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 100,
		WebsitePageEvidence: WebsitePageEvidence{HasOrderCTA: true},
		SocialPresence: SocialPresence{
			Status:   "present",
			Profiles: []SocialProfile{{Platform: "Linktree", Handle: "order-online", URL: "https://linktr.ee/order-online"}},
		},
	})

	wantScores := map[string]int{"website": 0, "order_online": 0}
	for _, metric := range report.Metrics {
		want, checked := wantScores[metric.Key]
		if !checked {
			continue
		}
		if metric.Score != want {
			t.Fatalf("%s score=%d, want %d", metric.Key, metric.Score, want)
		}
		delete(wantScores, metric.Key)
	}
	if len(wantScores) != 0 {
		t.Fatalf("missing metrics: %#v", wantScores)
	}
	if report.SocialPresence.Status != "not_found" || report.SocialPresence.Score != 0 {
		t.Fatalf("Linktree social presence=%#v", report.SocialPresence)
	}
	if len(report.Issues) == 0 || report.Issues[0].Title != "No dedicated restaurant website" {
		t.Fatalf("Linktree issue classification=%#v", report.Issues)
	}
}

func TestUnknownBusinessStatusDoesNotEarnOperationalPoint(t *testing.T) {
	unknown := scoreListing(PlaceDetails{}, false, 0, SocialPresence{})
	operational := scoreListing(PlaceDetails{BusinessStatus: "OPERATIONAL"}, false, 0, SocialPresence{})
	if unknown != 0 || operational != 1 {
		t.Fatalf("listing status scores unknown=%d operational=%d, want 0/1", unknown, operational)
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
