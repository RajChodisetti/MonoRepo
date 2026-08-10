package seoreport

import "testing"

func TestBuildCompetitorScanReturnsOnlyGenuinelyStrongerEligibleRows(t *testing.T) {
	target := visibilityFixture("target", "Target Thai", 0, 0, 4.2, 100)
	target.PhotoCount = 5
	target.Details.Website = "https://target.example"
	target.Details.Phone = "+61 2 0000 0000"
	target.HasHours = true

	strongest := visibilityFixture("strong-1", "Real Thai One", 0.01, 0.01, 4.9, 700)
	strongest.PhotoCount = 10
	strongest.Details.Website = "https://one.example"
	strongest.Details.Phone = "+61 2 1111 1111"
	strongest.HasHours = true
	strongest.DeliveryKnown = true
	strongest.Delivery = true

	stronger := visibilityFixture("strong-2", "Real Thai Two", 0.02, 0.01, 4.7, 450)
	stronger.PhotoCount = 10
	stronger.Details.Website = "https://two.example"
	stronger.Details.Phone = "+61 2 2222 2222"
	stronger.HasHours = true

	weak := visibilityFixture("weak", "Small Thai", 0.01, -0.01, 3.5, 20)
	closed := visibilityFixture("closed", "Closed Thai", 0.01, 0, 5, 1000)
	closed.Details.BusinessStatus = "CLOSED_PERMANENTLY"
	outside := visibilityFixture("outside", "Far Thai", 0.2, 0, 5, 1000)
	otherCuisine := visibilityFixture("other", "Italian Nearby", 0.01, 0, 5, 1000)
	otherCuisine.Details.PrimaryType = "italian_restaurant"
	otherCuisine.Details.Types = []string{"italian_restaurant", "restaurant"}
	duplicate := strongest

	scan := buildCompetitorScan(
		target,
		[]placeSnapshot{weak, stronger, closed, outside, strongest, otherCuisine, duplicate},
		"thai_restaurant",
	)
	if scan.Status != "complete" {
		t.Fatalf("status=%q, want complete", scan.Status)
	}
	if scan.SampleSize != 3 {
		t.Fatalf("sampleSize=%d, want three eligible unique candidates", scan.SampleSize)
	}
	if scan.CurrentRestaurantLeading {
		t.Fatal("target should not lead stronger sampled restaurants")
	}
	if scan.CurrentPosition != 3 || len(scan.Rows) != 2 {
		t.Fatalf("position=%d rows=%d, want position 3 with two stronger rows", scan.CurrentPosition, len(scan.Rows))
	}
	if scan.Rows[0].Name != "Real Thai One" || scan.Rows[0].Rank != "1st" {
		t.Fatalf("first row=%#v", scan.Rows[0])
	}
	if scan.Rows[1].Name != "Real Thai Two" || scan.Rows[1].Rank != "2nd" {
		t.Fatalf("second row=%#v", scan.Rows[1])
	}
	for _, row := range scan.Rows {
		if row.VisibilityScore <= scan.CurrentScore {
			t.Fatalf("row %q score=%d not stronger than current=%d", row.Name, row.VisibilityScore, scan.CurrentScore)
		}
		if row.ScoreMax != 100 || row.DistanceKM > 10 {
			t.Fatalf("invalid comparable row: %#v", row)
		}
	}
}

func TestBuildCompetitorScanHonestlyReportsLeadingAndNoData(t *testing.T) {
	target := visibilityFixture("target", "Leading Thai", 0, 0, 5, 1000)
	target.PhotoCount = 10
	target.Details.Website = "https://leading.example"
	target.Details.Phone = "+61 2 0000 0000"
	target.HasHours = true
	target.DeliveryKnown = true
	target.Delivery = true

	weaker := visibilityFixture("weaker", "Weaker Thai", 0.01, 0, 3, 10)
	scan := buildCompetitorScan(target, []placeSnapshot{weaker}, "thai_restaurant")
	if scan.Status != "complete" || !scan.CurrentRestaurantLeading || len(scan.Rows) != 0 || scan.CurrentPosition != 1 {
		t.Fatalf("unexpected leading scan: %#v", scan)
	}

	noData := buildCompetitorScan(target, nil, "thai_restaurant")
	if noData.Status != "no_data" || noData.CurrentRestaurantLeading || noData.SampleSize != 0 {
		t.Fatalf("no-data scan made an unsupported leading claim: %#v", noData)
	}
}

func visibilityFixture(id, name string, lat, lng, rating float64, reviews int) placeSnapshot {
	return placeSnapshot{
		Details: PlaceDetails{
			PlaceID:         id,
			Name:            name,
			Latitude:        floatPointer(lat),
			Longitude:       floatPointer(lng),
			Rating:          floatPointer(rating),
			UserRatingCount: intPointer(reviews),
			PrimaryType:     "thai_restaurant",
			Types:           []string{"thai_restaurant", "restaurant"},
			BusinessStatus:  "OPERATIONAL",
		},
	}
}

func floatPointer(value float64) *float64 { return &value }
func intPointer(value int) *int           { return &value }
