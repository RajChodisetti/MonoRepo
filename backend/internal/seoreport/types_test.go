package seoreport

import (
	"encoding/json"
	"testing"
)

func TestReportSerializesObservedZeroWebsiteQuality(t *testing.T) {
	data, err := json.Marshal(Report{
		WebsiteQualityScore:    0,
		WebsiteQualityAssessed: true,
	})
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode report JSON: %v", err)
	}
	value, ok := decoded["websiteQualityScore"]
	if !ok || value != float64(0) {
		t.Fatalf("observed zero website quality missing from JSON: %s", data)
	}
}
