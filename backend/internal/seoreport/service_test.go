package seoreport

import (
	"context"
	"encoding/json"
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

func TestGetReportStartsIndependentSourcesConcurrently(t *testing.T) {
	service := newReportTestService()
	service.reportBudget = time.Second

	release := make(chan struct{})
	placeStarted := make(chan struct{})
	contentStarted := make(chan struct{})
	service.fetchPlaceDetails = func(ctx context.Context, _ string) (*placeSnapshot, error) {
		close(placeStarted)
		select {
		case <-release:
			return &placeSnapshot{Details: PlaceDetails{
				PlaceID: "place-1",
				Name:    "Parallel Cafe",
				Address: "Melbourne VIC",
				Source:  "places",
			}}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	service.fetchSiteContent = func(ctx context.Context, _ string) (profiles.SiteContent, bool) {
		close(contentStarted)
		select {
		case <-release:
			return profiles.SiteContent{}, false
		case <-ctx.Done():
			return profiles.SiteContent{}, false
		}
	}

	type response struct {
		payload ReportResponse
		err     error
	}
	resultCh := make(chan response, 1)
	go func() {
		payload, err := service.GetReport(context.Background(), "place-1")
		resultCh <- response{payload: payload, err: err}
	}()

	for name, started := range map[string]<-chan struct{}{
		"Places":  placeStarted,
		"profile": contentStarted,
	} {
		select {
		case <-started:
		case <-time.After(300 * time.Millisecond):
			t.Fatalf("%s source did not start concurrently", name)
		}
	}
	close(release)

	select {
	case result := <-resultCh:
		if result.err != nil {
			t.Fatalf("GetReport returned error: %v", result.err)
		}
		if result.payload.Report.AnalysisSource != "automated" {
			t.Fatalf("expected truthful automated source, got %q", result.payload.Report.AnalysisSource)
		}
		if result.payload.Report.AnalysisStatus != "complete" {
			t.Fatalf("expected complete report, got %q", result.payload.Report.AnalysisStatus)
		}
	case <-time.After(time.Second):
		t.Fatal("GetReport did not complete after both sources were released")
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
	service.fetchSiteContent = func(context.Context, string) (profiles.SiteContent, bool) {
		return profiles.SiteContent{}, false
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
	if payload.Report.WebsiteQualityScore == 0 {
		t.Fatal("expected conservative website fallback score")
	}
}

func TestGetReportCoalescesConcurrentSamePlaceGeneration(t *testing.T) {
	service := newReportTestService()
	service.reportBudget = time.Second
	service.reportSlots = make(chan struct{}, 1)

	release := make(chan struct{})
	placeStarted := make(chan struct{})
	contentStarted := make(chan struct{})
	var placeCalls, contentCalls atomic.Int32
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
	service.fetchSiteContent = func(ctx context.Context, _ string) (profiles.SiteContent, bool) {
		contentCalls.Add(1)
		close(contentStarted)
		select {
		case <-release:
			return profiles.SiteContent{}, false
		case <-ctx.Done():
			return profiles.SiteContent{}, false
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
	select {
	case <-contentStarted:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("first profile call did not start")
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
	if placeCalls.Load() != 1 || contentCalls.Load() != 1 {
		t.Fatalf("provider calls: Places=%d profile=%d, want one each", placeCalls.Load(), contentCalls.Load())
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
	service.fetchSiteContent = func(context.Context, string) (profiles.SiteContent, bool) {
		providerCalls.Add(1)
		return profiles.SiteContent{}, false
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

func TestGetReportNeverPublishesLegacyInventoryImageURLs(t *testing.T) {
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
	service.fetchSiteContent = func(context.Context, string) (profiles.SiteContent, bool) {
		return profiles.SiteContent{
			Name: "Media Cafe",
			MenuItems: []profiles.SiteMenuItem{{
				Name:     "Legacy dish",
				ImageURL: "https://legacy.example/tripadvisor-menu.jpg",
			}},
			GalleryImages: []profiles.GalleryImage{{
				URL:          "https://legacy.example/gallery.jpg",
				ThumbnailURL: "https://legacy.example/gallery-thumb.jpg",
			}},
		}, true
	}

	payload, err := service.GetReport(context.Background(), "media-place")
	if err != nil {
		t.Fatalf("GetReport error: %v", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	for _, forbidden := range []string{
		"tripadvisor-menu.jpg",
		"gallery.jpg",
		"gallery-thumb.jpg",
		"legacy.example",
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("public report exposed legacy inventory URL fragment %q", forbidden)
		}
	}
	if payload.Place.Media == nil || len(payload.Place.Media.PhotosAndVideos) == 0 {
		t.Fatal("expected live Places photo media")
	}
	if got := payload.Place.Media.PhotosAndVideos[0].PhotoName; got != "places/media-place/photos/live-photo" {
		t.Fatalf("photoName=%q, want live Places resource", got)
	}
}
