package seoreport

import (
	"context"
	"testing"

	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

func TestGetReportIncludesVerifiedWebsiteEvidenceAndGenuineCompetitorScan(t *testing.T) {
	service := newReportTestService()
	target := visibilityFixture("target", "Target Thai", -33.86, 151.20, 4.0, 50)
	target.Details.Website = "https://target.example"
	target.Details.Source = "places"
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &target, nil
	}
	service.auditWebsite = func(context.Context, string, llmlib.Client) WebsiteAudit {
		return WebsiteAudit{
			Listed:           true,
			Reachable:        true,
			ViewportCoverage: "desktop_and_mobile",
			Source:           "vision",
			QualityScore:     60,
			MenuEvidence:     MenuEvidence{Status: "present", HasWebsiteLink: true, MenuURL: "https://target.example/menu", Rationale: "verified"},
			SocialPresence: SocialPresence{
				Status: "present",
				Profiles: []SocialProfile{{
					Platform: "Instagram", Handle: "target", URL: "https://instagram.com/target", Source: "website_link",
				}},
			},
		}
	}
	called := false
	service.fetchNearbyCompetitors = func(_ context.Context, place PlaceDetails, radius float64, limit int) ([]placeSnapshot, string, error) {
		called = true
		if place.PrimaryType != "thai_restaurant" || radius != 10_000 || limit != 12 {
			t.Fatalf("nearby args place=%#v radius=%v limit=%d", place, radius, limit)
		}
		rival := visibilityFixture("rival", "Actual Nearby Thai", -33.85, 151.20, 4.9, 600)
		rival.Details.Website = "https://rival.example"
		rival.Details.Phone = "+61 2 0000 0000"
		rival.HasHours = true
		rival.PhotoCount = 10
		return []placeSnapshot{rival}, "thai_restaurant", nil
	}

	payload, err := service.GetReport(context.Background(), "target")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if !called {
		t.Fatal("nearby competitor provider was not called")
	}
	if payload.Report.MenuEvidence.Status != "present" || payload.Report.SocialPresence.Status != "present" {
		t.Fatalf("website evidence missing: %#v %#v", payload.Report.MenuEvidence, payload.Report.SocialPresence)
	}
	if len(payload.Report.CompetitorScan.Rows) != 1 || payload.Report.CompetitorScan.Rows[0].Name != "Actual Nearby Thai" {
		t.Fatalf("genuine competitor scan missing: %#v", payload.Report.CompetitorScan)
	}
}
