package seoreport

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildReportWeightedWeakAverageStrong(t *testing.T) {
	weakRating, weakCount := 2.8, 8
	averageRating, averageCount := 4.1, 120
	strongRating, strongCount := 4.8, 850

	weak := BuildReport(ScoreInput{
		Place: PlaceDetails{
			Name:            "Weak Cafe",
			Address:         "1 Main Road, Melbourne VIC",
			Rating:          &weakRating,
			UserRatingCount: &weakCount,
			BusinessStatus:  "OPERATIONAL",
			Types:           []string{"restaurant"},
		},
		Reviews:         []Review{{Rating: 2, RelativeTime: "2 years ago"}},
		PlaceKnown:      true,
		ReviewsKnown:    true,
		PhotoCountKnown: true,
		HoursKnown:      true,
		WebsiteKnown:    true,
		MenuKnown:       true,
		PhoneKnown:      true,
		EmailKnown:      true,
		DeliveryKnown:   true,
		TakeoutKnown:    true,
		ReservableKnown: true,
	})

	average := BuildReport(ScoreInput{
		Place: PlaceDetails{
			Name:            "Neighbourhood Table",
			Address:         "12 Smith Street, Fitzroy VIC",
			Phone:           "+61 3 9000 0000",
			Website:         "https://neighbourhood-table.example",
			MapsURI:         "https://maps.google.com/?cid=2",
			Rating:          &averageRating,
			UserRatingCount: &averageCount,
			BusinessStatus:  "OPERATIONAL",
			Types:           []string{"italian_restaurant", "restaurant"},
		},
		Reviews: []Review{
			{Rating: 4, RelativeTime: "2 weeks ago"},
			{Rating: 4, RelativeTime: "4 months ago"},
		},
		PhotoCount:      6,
		HasHours:        true,
		Delivery:        true,
		PlaceKnown:      true,
		ReviewsKnown:    true,
		PhotoCountKnown: true,
		HoursKnown:      true,
		WebsiteKnown:    true,
		MenuKnown:       true,
		PhoneKnown:      true,
		EmailKnown:      true,
		DeliveryKnown:   true,
		TakeoutKnown:    true,
		ReservableKnown: true,
		Enrichment: Enrichment{
			MenuItemCount: 3,
		},
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 40,
	})

	strong := BuildReport(ScoreInput{
		Place: PlaceDetails{
			Name:             "Melbourne Italian Kitchen",
			Address:          "123 Collins Street, Melbourne VIC",
			Phone:            "+61 3 9000 1111",
			Email:            "hello@kitchen.example",
			Website:          "https://kitchen.example/order-online",
			MapsURI:          "https://maps.google.com/?cid=3",
			Rating:           &strongRating,
			UserRatingCount:  &strongCount,
			BusinessStatus:   "OPERATIONAL",
			Types:            []string{"italian_restaurant", "pizza_restaurant"},
			EditorialSummary: "Italian dining and pizza in Melbourne.",
		},
		Reviews: []Review{
			{Rating: 5, RelativeTime: "2 days ago"},
			{Rating: 5, RelativeTime: "1 week ago"},
			{Rating: 4.8, RelativeTime: "3 weeks ago"},
		},
		PhotoCount:      10,
		HasHours:        true,
		Delivery:        true,
		Takeout:         true,
		Reservable:      true,
		PlaceKnown:      true,
		ReviewsKnown:    true,
		PhotoCountKnown: true,
		HoursKnown:      true,
		WebsiteKnown:    true,
		MenuKnown:       true,
		PhoneKnown:      true,
		EmailKnown:      true,
		DeliveryKnown:   true,
		TakeoutKnown:    true,
		ReservableKnown: true,
		Enrichment: Enrichment{
			MenuItemCount:  14,
			MenuImageCount: 4,
		},
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 90,
	})

	if weak.OverallScore >= 35 {
		t.Fatalf("weak score = %d, want below 35", weak.OverallScore)
	}
	if average.OverallScore < 45 || average.OverallScore >= 75 {
		t.Fatalf("average score = %d, want 45..74", average.OverallScore)
	}
	if strong.OverallScore < 80 {
		t.Fatalf("strong score = %d, want at least 80", strong.OverallScore)
	}
	if average.OverallScore-weak.OverallScore < 15 || strong.OverallScore-average.OverallScore < 15 {
		t.Fatalf("scores insufficiently separated: weak=%d average=%d strong=%d", weak.OverallScore, average.OverallScore, strong.OverallScore)
	}

	totalWeight := 0
	for _, metric := range strong.Metrics {
		totalWeight += metric.Max
		if metric.Score < 0 || metric.Score > metric.Max {
			t.Fatalf("metric %s score %d outside 0..%d", metric.Key, metric.Score, metric.Max)
		}
	}
	if totalWeight != 100 {
		t.Fatalf("metric weights total %d, want 100", totalWeight)
	}
	if len(strong.Competitors) != 0 {
		t.Fatalf("competitors = %#v, want no comparison without a verified competitor set", strong.Competitors)
	}
}

func TestBuildReportDistinguishesUnknownFromConfirmedAbsent(t *testing.T) {
	zeroReviews := 0
	if got := scoreReviews(ScoreInput{Place: PlaceDetails{UserRatingCount: &zeroReviews}}); got != 0 {
		t.Fatalf("confirmed zero reviews scored %d, want 0 even when recent-review provider data is unavailable", got)
	}

	unknown := BuildReport(ScoreInput{Place: PlaceDetails{Name: "Partial Provider Result"}})
	absent := BuildReport(ScoreInput{
		Place:           PlaceDetails{Name: "Confirmed Empty Listing"},
		PlaceKnown:      true,
		ReviewsKnown:    true,
		PhotoCountKnown: true,
		HoursKnown:      true,
		WebsiteKnown:    true,
		MenuKnown:       true,
		PhoneKnown:      true,
		EmailKnown:      true,
		DeliveryKnown:   true,
		TakeoutKnown:    true,
		ReservableKnown: true,
	})

	if absent.OverallScore != 0 {
		t.Fatalf("confirmed absent score = %d, want 0 with no hard floor", absent.OverallScore)
	}
	if unknown.OverallScore != 0 {
		t.Fatalf("unknown score = %d, want unavailable evidence to earn no points", unknown.OverallScore)
	}
	if unknown.OverallLabel != "Not assessed" || absent.OverallLabel != "Poor" {
		t.Fatalf("unknown label=%q absent label=%q, want distinct evidence states", unknown.OverallLabel, absent.OverallLabel)
	}
	for _, metric := range unknown.Metrics {
		if metric.Status != "Not assessed" {
			t.Fatalf("unknown metric %s status=%q, want Not assessed", metric.Key, metric.Status)
		}
	}
	if len(unknown.Issues) != 0 {
		t.Fatalf("unknown evidence produced unsupported issues: %#v", unknown.Issues)
	}
}

func TestScoreWebsiteUsesObservedQualityWithoutArtificialBand(t *testing.T) {
	absent := scoreWebsite("", false, 0)
	unknownWebsite := scoreWebsite("", false, 0)
	unaudited := scoreWebsite("https://restaurant.example", false, 0)
	observedZero := scoreWebsite("https://restaurant.example", true, 0)
	observedHigh := scoreWebsite("https://restaurant.example", true, 90)

	if absent != 0 || unknownWebsite != 0 {
		t.Fatalf("absent=%d unknown=%d, want no unsupported points", absent, unknownWebsite)
	}
	if observedZero != 8 {
		t.Fatalf("observed zero quality = %d, want only 8 verified-presence points", observedZero)
	}
	if unaudited != observedZero || observedHigh <= unaudited {
		t.Fatalf("website scores not ordered: observed-zero=%d unaudited=%d observed-high=%d", observedZero, unaudited, observedHigh)
	}
	if social := scoreWebsite("https://instagram.com/myrestaurant", true, 90); social >= observedZero {
		t.Fatalf("social profile score = %d, want below dedicated site baseline %d", social, observedZero)
	}
	if clampWebsiteQuality(7) != 7 || clampWebsiteQuality(93) != 93 {
		t.Fatal("website quality clamp must preserve valid observed values")
	}
	if clampWebsiteQuality(-1) != 0 || clampWebsiteQuality(101) != 100 {
		t.Fatal("website quality clamp must only enforce the public 0..100 range")
	}
	partialWebsite := BuildReport(ScoreInput{
		Place:        PlaceDetails{Website: "https://restaurant.example"},
		WebsiteKnown: true,
	})
	if partialWebsite.OverallLabel != "Partial" {
		t.Fatalf("unaudited dedicated website label=%q, want Partial", partialWebsite.OverallLabel)
	}
	for _, metric := range partialWebsite.Metrics {
		if metric.Key == "website" && metric.Status != "Partially assessed" {
			t.Fatalf("unaudited website metric status=%q, want Partially assessed", metric.Status)
		}
	}
	observedZeroReport := BuildReport(ScoreInput{
		Place:               PlaceDetails{Website: "https://restaurant.example"},
		WebsiteKnown:        true,
		WebsiteQualityKnown: true,
		WebsiteQualityScore: 0,
	})
	if !observedZeroReport.WebsiteQualityAssessed || observedZeroReport.WebsiteQualityScore != 0 {
		t.Fatalf("observed zero audit lost assessment state: assessed=%v score=%d", observedZeroReport.WebsiteQualityAssessed, observedZeroReport.WebsiteQualityScore)
	}
	unauditedReport := BuildReport(ScoreInput{
		Place:               PlaceDetails{Website: "https://instagram.com/myrestaurant"},
		WebsiteKnown:        true,
		WebsiteQualityScore: 28,
	})
	if unauditedReport.WebsiteQualityAssessed || unauditedReport.WebsiteQualityScore != 0 {
		t.Fatalf("rule-based website quality surfaced as visual assessment: assessed=%v score=%d", unauditedReport.WebsiteQualityAssessed, unauditedReport.WebsiteQualityScore)
	}
}

func TestReviewScoreRewardsQualityVolumeAndRecency(t *testing.T) {
	lowRating, lowCount := 3.2, 12
	highRating, highCount := 4.7, 420
	low := scoreReviews(ScoreInput{
		Place:        PlaceDetails{Rating: &lowRating, UserRatingCount: &lowCount},
		PlaceKnown:   true,
		ReviewsKnown: true,
		Reviews:      []Review{{Rating: 3, RelativeTime: "2 years ago"}},
	})
	high := scoreReviews(ScoreInput{
		Place:        PlaceDetails{Rating: &highRating, UserRatingCount: &highCount},
		PlaceKnown:   true,
		ReviewsKnown: true,
		Reviews: []Review{
			{Rating: 5, RelativeTime: "2 days ago"},
			{Rating: 4.5, RelativeTime: "3 weeks ago"},
		},
	})
	if high-low < 12 {
		t.Fatalf("review evidence insufficiently discriminating: low=%d high=%d", low, high)
	}
}

func TestReviewIssueDoesNotMislabelLowStrengthAsLowVolume(t *testing.T) {
	rating, count := 2.5, 1000
	report := BuildReport(ScoreInput{
		Place: PlaceDetails{
			Name:            "High Volume Restaurant",
			Rating:          &rating,
			UserRatingCount: &count,
		},
		Reviews:      []Review{{Rating: 2, RelativeTime: "2 years ago"}},
		PlaceKnown:   true,
		ReviewsKnown: true,
	})
	foundStrength := false
	for _, issue := range report.Issues {
		if issue.Title == "Low review volume" {
			t.Fatalf("high-volume venue received false volume issue: %#v", issue)
		}
		if issue.Title == "Review strength needs attention" {
			foundStrength = true
		}
	}
	if !foundStrength {
		t.Fatalf("expected evidence-specific review strength issue, got %#v", report.Issues)
	}
}

func TestListingScoreReachesFullWeightAtProviderPhotoCap(t *testing.T) {
	in := ScoreInput{
		Place: PlaceDetails{
			Address:        "12 Smith Street, Fitzroy VIC",
			MapsURI:        "https://maps.google.com/?cid=4",
			BusinessStatus: "OPERATIONAL",
		},
		PhotoCount:      10,
		HasHours:        true,
		PlaceKnown:      true,
		PhotoCountKnown: true,
		HoursKnown:      true,
	}
	if got := scoreListing(in.Place, true, in); got != listingWeight {
		t.Fatalf("listing score at 10-photo provider cap = %d, want %d", got, listingWeight)
	}
}

func TestParsePlacePhotosHonorsProviderMaximum(t *testing.T) {
	raw := make([]any, 12)
	for index := range raw {
		raw[index] = map[string]any{"name": fmt.Sprintf("places/example/photos/%d", index)}
	}
	if got := len(parsePlacePhotos(raw)); got != 10 {
		t.Fatalf("parsed photos = %d, want official Places maximum 10", got)
	}
}

func TestWeightedScoresDoNotCluster(t *testing.T) {
	ratings := []float64{2.4, 3.1, 3.8, 4.4, 4.9}
	counts := []int{2, 12, 55, 220, 900}
	seen := make(map[int]struct{}, len(ratings))
	minimum, maximum := 101, -1
	for i := range ratings {
		report := BuildReport(ScoreInput{
			Place: PlaceDetails{
				Name:            "Calibration Restaurant",
				Address:         "Melbourne VIC",
				Rating:          &ratings[i],
				UserRatingCount: &counts[i],
				BusinessStatus:  "OPERATIONAL",
			},
			Reviews:         []Review{{Rating: ratings[i], RelativeTime: []string{"3 years ago", "1 year ago", "6 months ago", "2 months ago", "2 days ago"}[i]}},
			PlaceKnown:      true,
			ReviewsKnown:    true,
			WebsiteKnown:    true,
			MenuKnown:       true,
			PhoneKnown:      true,
			EmailKnown:      true,
			PhotoCountKnown: true,
			HoursKnown:      true,
			DeliveryKnown:   true,
			TakeoutKnown:    true,
			ReservableKnown: true,
		})
		seen[report.OverallScore] = struct{}{}
		minimum = min(minimum, report.OverallScore)
		maximum = max(maximum, report.OverallScore)
	}
	if len(seen) != len(ratings) {
		t.Fatalf("scores clustered despite distinct evidence: %#v", seen)
	}
	if maximum-minimum < 18 {
		t.Fatalf("score spread = %d (%d..%d), want at least 18", maximum-minimum, minimum, maximum)
	}
}

func TestBuildReportLeavesSummaryForSummarizer(t *testing.T) {
	report := BuildReport(ScoreInput{Place: PlaceDetails{Name: "Quiet Cafe"}})
	if report.AISummary != "" {
		t.Fatal("BuildReport should leave AISummary empty for summarizer")
	}
	if !report.FullReportLocked {
		t.Fatal("expected full report locked")
	}
	summary := buildDeterministicSummary(reportPlace(report), report, nil)
	lines := nonEmptyLines(summary)
	if len(lines) < 3 || len(lines) > 4 {
		t.Fatalf("expected 3-4 summary lines, got %d: %q", len(lines), summary)
	}
}

func reportPlace(report Report) PlaceDetails {
	return PlaceDetails{Name: report.RestaurantName, Address: report.Address}
}

func nonEmptyLines(value string) []string {
	parts := strings.Split(value, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
