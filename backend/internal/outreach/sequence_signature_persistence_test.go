package outreach

import (
	"strings"
	"testing"
)

func TestUpdateSequenceDraftSignaturePersistenceContract(t *testing.T) {
	returningIndex := strings.Index(updateSequenceDraftQuery, "RETURNING")
	if returningIndex < 0 {
		t.Fatal("update sequence query does not return the saved row")
	}
	returningClause := updateSequenceDraftQuery[returningIndex:]
	for field, parameter := range map[string]string{
		"signature_name":    "$3",
		"signature_title":   "$4",
		"signature_details": "$5",
	} {
		if !strings.Contains(updateSequenceDraftQuery, field+" = "+parameter) {
			t.Fatalf("update sequence query does not persist %s from %s", field, parameter)
		}
		if !strings.Contains(returningClause, field) {
			t.Fatalf("update sequence query does not return persisted %s", field)
		}
		if !strings.Contains(listSequencesQuery, field) {
			t.Fatalf("list sequence query does not reload persisted %s", field)
		}
	}
	if !strings.Contains(updateSequenceDraftQuery, "updated_at = now()") {
		t.Fatal("update sequence query does not create a fresh updated_at concurrency token")
	}
	if !strings.Contains(returningClause, "updated_at") {
		t.Fatal("update sequence query does not return the fresh updated_at concurrency token")
	}
}

func TestNormalizeSequenceSignatureMatchesSavedResponseValues(t *testing.T) {
	got, err := normalizeSequenceSignature(SequenceSignature{
		Name:              "  Alex Morgan  ",
		Title:             "  Partnerships Manager  ",
		AdditionalDetails: " Phone: +61 400 000 000\r\n Available weekdays ",
	})
	if err != nil {
		t.Fatalf("normalizeSequenceSignature() error = %v", err)
	}
	want := SequenceSignature{
		Name:              "Alex Morgan",
		Title:             "Partnerships Manager",
		AdditionalDetails: "Phone: +61 400 000 000\n Available weekdays",
	}
	if got != want {
		t.Fatalf("normalizeSequenceSignature() = %#v, want %#v", got, want)
	}
}
