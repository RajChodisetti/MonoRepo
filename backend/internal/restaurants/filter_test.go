package restaurants

import "testing"

func TestMatchesFilterByRestaurantName(t *testing.T) {
	record := Restaurant{Name: "Thai Rama", Status: StatusLead}

	if !MatchesFilter(record, ListFilter{Restaurant: "thai"}) {
		t.Fatal("expected thai filter to match Thai Rama")
	}
	if MatchesFilter(record, ListFilter{Restaurant: "indian"}) {
		t.Fatal("expected indian filter not to match Thai Rama")
	}
}

func TestMatchesFilterExcludesArchivedByDefault(t *testing.T) {
	record := Restaurant{Name: "Closed", Status: StatusArchived}

	if MatchesFilter(record, ListFilter{}) {
		t.Fatal("archived restaurant should be excluded by default")
	}
	if !MatchesFilter(record, ListFilter{IncludeArchived: true}) {
		t.Fatal("archived restaurant should be included when requested")
	}
}

func TestMatchesFilterByOCRStatus(t *testing.T) {
	record := Restaurant{Name: "Verified Cafe", Status: StatusLead, OCRStatus: "verified"}

	if !MatchesFilter(record, ListFilter{OCRStatus: "verified"}) {
		t.Fatal("expected verified OCR filter to match")
	}
	if MatchesFilter(record, ListFilter{OCRStatus: "failed"}) {
		t.Fatal("expected failed OCR filter not to match verified restaurant")
	}
}
