package seoreport

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
)

func newReportTestService() *Service {
	return NewService(
		config.PlacesConfig{},
		nil,
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

type panicProfileRepository struct{ profiles.SiteRepository }

func TestNewServiceNeverUsesInternalProfilesForPublicSearchOrReport(t *testing.T) {
	// Any call through the embedded nil interface would panic. Passing this
	// repository proves the public constructor neither retains nor invokes it.
	service := NewService(
		config.PlacesConfig{},
		&panicProfileRepository{},
		nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	search, err := service.SearchRestaurants(context.Background(), "Thai", "Sydney", 8)
	if err != nil {
		t.Fatalf("SearchRestaurants: %v", err)
	}
	if search.Meta.InventoryEnabled || len(search.Results) != 0 {
		t.Fatalf("public search exposed internal inventory: %#v", search)
	}

	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "place-1",
			Name:    "Public Places Cafe",
			Source:  "places",
		}}, nil
	}
	payload, err := service.GetReport(context.Background(), "place-1")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if payload.Place.Source != "places" || payload.Place.Name != "Public Places Cafe" {
		t.Fatalf("public report source=%#v", payload.Place)
	}
}

func TestGetReportReturnsConservativePartialAtBudget(t *testing.T) {
	service := newReportTestService()
	service.reportBudget = 80 * time.Millisecond
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "place-2",
			Name:    "Slow Site Bistro",
			Address: "Sydney NSW",
			Website: "https://slow.example",
			Source:  "places",
		}}, nil
	}
	auditStopped := make(chan struct{})
	service.auditWebsite = func(ctx context.Context, website string, _ llmlib.Client) WebsiteAudit {
		<-ctx.Done()
		close(auditStopped)
		return fallbackWebsiteAudit(website, "test timeout")
	}

	startedAt := time.Now()
	payload, err := service.GetReport(context.Background(), "place-2")
	if err != nil {
		t.Fatalf("GetReport returned error: %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > 350*time.Millisecond {
		t.Fatalf("report exceeded bounded fallback window: %s", elapsed)
	}
	select {
	case <-auditStopped:
	default:
		t.Fatal("website audit did not observe cancellation before GetReport returned")
	}
	if payload.Report.AnalysisStatus != "partial" {
		t.Fatalf("expected partial status, got %q", payload.Report.AnalysisStatus)
	}
	if payload.Report.AnalysisSource != "automated" {
		t.Fatalf("fallback must not claim AI, got %q", payload.Report.AnalysisSource)
	}
	if payload.Report.WebsiteQualityScore != 0 {
		t.Fatalf("unreachable website received visual quality=%d", payload.Report.WebsiteQualityScore)
	}
	for _, metric := range payload.Report.Metrics {
		if metric.Key == "website" && metric.Score != 0 {
			t.Fatalf("unreachable website metric=%d, want zero", metric.Score)
		}
	}
}

func TestGetReportDoesNotAuditLinktreeAsDedicatedWebsite(t *testing.T) {
	service := newReportTestService()
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "linktree-listing",
			Name:    "Aggregator Cafe",
			Website: "https://linktr.ee/aggregator-cafe",
			Source:  "places",
		}}, nil
	}

	payload, err := service.GetReport(context.Background(), "linktree-listing")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	for _, metric := range payload.Report.Metrics {
		if metric.Key != "website" {
			continue
		}
		if metric.Score != 0 || !strings.Contains(metric.Rationale, "Linktree aggregator") {
			t.Fatalf("Linktree website metric=%#v", metric)
		}
		if len(metric.Evidence) == 0 || metric.Evidence[0] != "Listed destination: Linktree aggregator" {
			t.Fatalf("Linktree website evidence=%#v", metric.Evidence)
		}
		return
	}
	t.Fatal("website metric missing")
}

func TestGetReportInvalidVisionUsesReachabilityOnlyAutomatedPartial(t *testing.T) {
	service := newReportTestService()
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "invalid-vision",
			Name:    "Vision Truth Cafe",
			Website: "https://vision.example",
			Source:  "places",
		}}, nil
	}
	service.auditWebsite = func(context.Context, string, llmlib.Client) WebsiteAudit {
		// These are the capture properties AuditWebsite attaches after the exact
		// response parser rejects malformed provider output.
		audit := websiteAuditFromVisionResponse("https://vision.example", `{"score":"excellent"}`)
		audit.Reachable = true
		audit.ViewportCoverage = "desktop_and_mobile"
		audit.Screenshot = "data:image/jpeg;base64,desktop"
		audit.MobileScreenshot = "data:image/jpeg;base64,mobile"
		audit.Review = "The live homepage was reached, but automated visual-quality analysis did not complete. Only reachability was scored; no visual-quality claim was made."
		audit.MenuEvidence = MenuEvidence{Status: "not_found", Rationale: "No menu link."}
		audit.SocialPresence = SocialPresence{Status: "not_found"}
		return audit
	}

	payload, err := service.GetReport(context.Background(), "invalid-vision")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if payload.Report.AnalysisSource != "automated" || payload.Report.AnalysisStatus != "partial" {
		t.Fatalf("invalid vision source/status=%q/%q, want automated/partial", payload.Report.AnalysisSource, payload.Report.AnalysisStatus)
	}
	if payload.Report.WebsiteQualityScore != 0 {
		t.Fatalf("invalid vision published quality score=%d, want zero", payload.Report.WebsiteQualityScore)
	}
	for _, metric := range payload.Report.Metrics {
		if metric.Key == "website" {
			if metric.Score != 4 {
				t.Fatalf("invalid vision website score=%d, want reachability-only 4", metric.Score)
			}
			return
		}
	}
	t.Fatal("website metric missing")
}

func TestGetReportFillsOnlyMissingPlaceContactsFromOfficialWebsiteLinks(t *testing.T) {
	tests := []struct {
		name      string
		place     PlaceDetails
		wantEmail string
		wantPhone string
	}{
		{
			name: "fills missing contacts",
			place: PlaceDetails{
				PlaceID: "contact-missing",
				Name:    "Contact Cafe",
				Website: "https://contact.example",
				Source:  "places",
			},
			wantEmail: "hello@contact.example",
			wantPhone: "+61 3 9000 0000",
		},
		{
			name: "preserves Places contacts",
			place: PlaceDetails{
				PlaceID: "contact-existing",
				Name:    "Existing Contact Cafe",
				Website: "https://contact.example",
				Email:   "places@contact.example",
				Phone:   "+61 2 8111 1111",
				Source:  "places",
			},
			wantEmail: "places@contact.example",
			wantPhone: "+61 2 8111 1111",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newReportTestService()
			service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
				return &placeSnapshot{Details: test.place}, nil
			}
			service.auditWebsite = func(context.Context, string, llmlib.Client) WebsiteAudit {
				return WebsiteAudit{
					Listed:           true,
					Reachable:        true,
					ViewportCoverage: "desktop_and_mobile",
					Source:           "vision",
					QualityScore:     50,
					PublicEmail:      "hello@contact.example",
					PublicPhone:      "+61 3 9000 0000",
					MenuEvidence:     MenuEvidence{Status: "not_found", Rationale: "No menu link."},
					SocialPresence:   SocialPresence{Status: "not_found"},
				}
			}

			payload, err := service.GetReport(context.Background(), test.place.PlaceID)
			if err != nil {
				t.Fatalf("GetReport: %v", err)
			}
			if payload.Place.Email != test.wantEmail || payload.Place.Phone != test.wantPhone {
				t.Fatalf("contacts email=%q phone=%q, want %q / %q", payload.Place.Email, payload.Place.Phone, test.wantEmail, test.wantPhone)
			}
			contactScored := false
			for _, metric := range payload.Report.Metrics {
				if metric.Key == "contact" {
					contactScored = true
					if metric.Score != 10 {
						t.Fatalf("contact score=%d, want contacts applied before scoring", metric.Score)
					}
				}
			}
			if !contactScored {
				t.Fatal("contact metric missing")
			}
		})
	}
}

func TestGetReportMarksSingleViewportVisionAsPartial(t *testing.T) {
	service := newReportTestService()
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "mobile-only",
			Name:    "Mobile Cafe",
			Website: "https://mobile.example",
			Source:  "places",
		}}, nil
	}
	service.auditWebsite = func(context.Context, string, llmlib.Client) WebsiteAudit {
		return WebsiteAudit{
			Listed:           true,
			Reachable:        true,
			ViewportCoverage: "mobile",
			Source:           "vision",
			QualityScore:     55,
			MobileScreenshot: "data:image/jpeg;base64,mobile",
			MenuEvidence:     MenuEvidence{Status: "not_found", Rationale: "No menu link."},
			SocialPresence:   SocialPresence{Status: "not_found"},
		}
	}
	payload, err := service.GetReport(context.Background(), "mobile-only")
	if err != nil {
		t.Fatalf("GetReport: %v", err)
	}
	if payload.Report.AnalysisStatus != "partial" || !strings.Contains(payload.Report.AnalysisNotice, "Only the mobile") {
		t.Fatalf("single viewport status=%q notice=%q", payload.Report.AnalysisStatus, payload.Report.AnalysisNotice)
	}
	if payload.Report.WebsiteQualityScore != 0 {
		t.Fatalf("single viewport quality score=%d, want no visual-quality score", payload.Report.WebsiteQualityScore)
	}
	for _, metric := range payload.Report.Metrics {
		if metric.Key == "website" && metric.Score != 4 {
			t.Fatalf("single viewport website score=%d, want reachability-only 4", metric.Score)
		}
	}
}

func TestGetReportCoalescesConcurrentSamePlaceGeneration(t *testing.T) {
	service := newReportTestService()
	service.reportBudget = time.Second
	service.reportSlots = make(chan struct{}, 1)

	release := make(chan struct{})
	placeStarted := make(chan struct{})
	var placeCalls atomic.Int32
	service.fetchPlaceDetails = func(ctx context.Context, _ string) (*placeSnapshot, error) {
		placeCalls.Add(1)
		close(placeStarted)
		select {
		case <-release:
			return &placeSnapshot{Details: PlaceDetails{
				PlaceID: "shared-place",
				Name:    "Shared Cafe",
				Source:  "places",
			}}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	type response struct {
		payload ReportResponse
		err     error
	}
	first := make(chan response, 1)
	go func() {
		payload, err := service.GetReport(context.Background(), "shared-place")
		first <- response{payload: payload, err: err}
	}()
	select {
	case <-placeStarted:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("first Places call did not start")
	}
	second := make(chan response, 1)
	go func() {
		payload, err := service.GetReport(context.Background(), "shared-place")
		second <- response{payload: payload, err: err}
	}()
	select {
	case result := <-second:
		t.Fatalf("same-place caller did not share in-flight work: %v", result.err)
	case <-time.After(40 * time.Millisecond):
	}
	close(release)

	for name, resultCh := range map[string]<-chan response{"first": first, "second": second} {
		select {
		case result := <-resultCh:
			if result.err != nil {
				t.Fatalf("%s GetReport error: %v", name, result.err)
			}
			if result.payload.Place.PlaceID != "shared-place" {
				t.Fatalf("%s place=%q", name, result.payload.Place.PlaceID)
			}
		case <-time.After(time.Second):
			t.Fatalf("%s GetReport did not return", name)
		}
	}
	if placeCalls.Load() != 1 {
		t.Fatalf("Places calls=%d, want one", placeCalls.Load())
	}
}

func TestGetReportDoesNotCachePlacesContentAfterGeneration(t *testing.T) {
	service := newReportTestService()
	service.reportSlots = make(chan struct{}, 1)
	var placeCalls atomic.Int32
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		placeCalls.Add(1)
		return &placeSnapshot{Details: PlaceDetails{
			PlaceID: "fresh-place",
			Name:    "Fresh Cafe",
			Source:  "places",
		}}, nil
	}
	for i := 0; i < 2; i++ {
		if _, err := service.GetReport(context.Background(), "fresh-place"); err != nil {
			t.Fatalf("GetReport call %d: %v", i+1, err)
		}
	}
	if placeCalls.Load() != 2 {
		t.Fatalf("Places detail calls=%d, want one fresh call per completed report request", placeCalls.Load())
	}
}

func TestGetReportCapacityExhaustionFailsBeforeProviderCalls(t *testing.T) {
	service := newReportTestService()
	service.reportSlots = make(chan struct{}, 1)
	service.reportSlots <- struct{}{}
	var providerCalls atomic.Int32
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		providerCalls.Add(1)
		return nil, errors.New("provider should not run")
	}
	started := time.Now()
	_, err := service.GetReport(context.Background(), "busy-place")
	if !errors.Is(err, ErrReportBusy) {
		t.Fatalf("GetReport error=%v, want ErrReportBusy", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("capacity rejection took %s, want fast failure", elapsed)
	}
	if providerCalls.Load() != 0 {
		t.Fatalf("provider calls=%d, want 0", providerCalls.Load())
	}
}

func TestGetReportPublishesOnlyLivePlacesPhotoResources(t *testing.T) {
	service := newReportTestService()
	service.reportSlots = make(chan struct{}, 1)
	service.fetchPlaceDetails = func(context.Context, string) (*placeSnapshot, error) {
		return &placeSnapshot{
			Details: PlaceDetails{
				PlaceID: "media-place",
				Name:    "Media Cafe",
				MapsURI: "https://maps.google.com/?cid=123",
				Source:  "places",
			},
			Photos: []placePhoto{{
				Name:          "places/media-place/photos/live-photo",
				GoogleMapsURI: "https://maps.google.com/?cid=123",
			}},
		}, nil
	}
	payload, err := service.GetReport(context.Background(), "media-place")
	if err != nil {
		t.Fatalf("GetReport error: %v", err)
	}
	if payload.Place.Media == nil || len(payload.Place.Media.PhotosAndVideos) == 0 {
		t.Fatal("expected live Places photo media")
	}
	if got := payload.Place.Media.PhotosAndVideos[0].PhotoName; got != "places/media-place/photos/live-photo" {
		t.Fatalf("photoName=%q, want live Places resource", got)
	}
}
