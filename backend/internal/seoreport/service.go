package seoreport

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
	"golang.org/x/sync/singleflight"
)

const (
	reportGenerationBudget = 15 * time.Second
	reportSummaryBudget    = 2200 * time.Millisecond
	maxConcurrentReports   = 2
)

var (
	// ErrNotFound indicates the restaurant could not be resolved.
	ErrNotFound = errors.New("seo report restaurant not found")
	// ErrReportBusy fails fast before any provider work when public generation
	// capacity is exhausted. Callers can retry after an in-flight report finishes.
	ErrReportBusy = errors.New("seo report generation is at capacity")
)

var sharedReportGenerationSlots = make(chan struct{}, maxConcurrentReports)

// Service orchestrates public Places data, official-website evidence, scoring,
// and summary generation. It deliberately has no restaurant-profile repository:
// the public review surface must not consume unreviewed internal inventory data.
type Service struct {
	places     *PlacesClient
	summarizer Summarizer
	llm        llmlib.Client
	log        *slog.Logger

	interested InterestedRepository
	mailer     EmailSender
	appEnv     string
	unlockRate *unlockRequestLimiter

	reportCalls singleflight.Group
	reportSlots chan struct{}

	// Private hooks keep orchestration testable without provider calls. Production
	// constructors always wire the concrete Places and website clients.
	fetchPlaceDetails func(context.Context, string) (*placeSnapshot, error)
	// Deprecated compatibility hook for older focused tests. Public search and
	// report generation deliberately never invoke it, and NewService never wires it.
	fetchSiteContent       func(context.Context, string) (profiles.SiteContent, bool)
	fetchNearbyCompetitors func(context.Context, PlaceDetails, float64, int) ([]placeSnapshot, string, error)
	auditWebsite           func(context.Context, string, llmlib.Client) WebsiteAudit
	reportBudget           time.Duration
}

// NewService constructs the SEO report service.
func NewService(
	placesCfg config.PlacesConfig,
	profilesRepo profiles.SiteRepository,
	llmClient llmlib.Client,
	log *slog.Logger,
) *Service {
	// Preserve the constructor contract for current callers while intentionally
	// refusing to retain the internal profile repository on this public surface.
	_ = profilesRepo
	if log == nil {
		log = slog.Default()
	}
	service := &Service{
		places: NewPlacesClient(placesCfg),
		// Places listing and review content stays in the deterministic pipeline.
		// The LLM client is reserved for first-party website screenshot analysis.
		summarizer:   DeterministicSummarizer{},
		llm:          llmClient,
		log:          log,
		reportBudget: reportGenerationBudget,
		reportSlots:  sharedReportGenerationSlots,
		unlockRate:   newUnlockRequestLimiter(),
	}
	service.fetchPlaceDetails = service.places.GetPlaceDetails
	service.fetchNearbyCompetitors = service.places.SearchNearbyCuisine
	service.auditWebsite = AuditWebsite
	return service
}

// SearchRestaurants returns only public Google Places search results.
// location is a city name or postcode; empty defaults to Australia.
func (s *Service) SearchRestaurants(ctx context.Context, query, location string, limit int) (SearchResponse, error) {
	query = strings.TrimSpace(query)
	location = normalizeSearchLocation(location)
	if limit <= 0 {
		limit = 8
	}
	if limit > 12 {
		limit = 12
	}

	resp := SearchResponse{
		Results: []PlaceSummary{},
		Meta: SearchMeta{
			PlacesEnabled:    s.places.Enabled(),
			InventoryEnabled: false,
		},
	}
	if len(query) < 2 {
		return resp, nil
	}

	if s.places.Enabled() {
		places, err := s.places.SearchRestaurants(ctx, query, location, limit)
		if err != nil {
			s.log.WarnContext(ctx, "seo_places_search_failed", "error", err)
		} else {
			for _, item := range places {
				if strings.TrimSpace(item.PlaceID) == "" {
					continue
				}
				resp.Results = append(resp.Results, item)
				if len(resp.Results) >= limit {
					break
				}
			}
		}
	}

	return resp, nil
}

// GetReport builds a fresh SEO report for a place ID. Concurrent callers for
// the same place share one in-flight generation, but completed Places content
// is not retained or reused.
func (s *Service) GetReport(ctx context.Context, placeID string) (ReportResponse, error) {
	placeID = sanitizePlaceID(placeID)
	if placeID == "" {
		return ReportResponse{}, ErrNotFound
	}

	// DoChan lets each HTTP caller stop waiting independently while the shared,
	// 15-second-bounded generation finishes for remaining concurrent callers.
	result := s.reportCalls.DoChan(placeID, func() (any, error) {
		return s.generateReport(context.WithoutCancel(ctx), placeID)
	})
	select {
	case <-ctx.Done():
		return ReportResponse{}, ctx.Err()
	case completed := <-result:
		if completed.Err != nil {
			return ReportResponse{}, completed.Err
		}
		payload, ok := completed.Val.(ReportResponse)
		if !ok {
			return ReportResponse{}, errors.New("invalid shared report result")
		}
		return payload, nil
	}
}

func (s *Service) generateReport(ctx context.Context, placeID string) (ReportResponse, error) {
	reportSlots := s.reportSlots
	if reportSlots == nil {
		reportSlots = sharedReportGenerationSlots
	}
	select {
	case reportSlots <- struct{}{}:
		defer func() { <-reportSlots }()
	default:
		return ReportResponse{}, ErrReportBusy
	}

	startedAt := time.Now()

	budget := s.reportBudget
	if budget <= 0 {
		budget = reportGenerationBudget
	}
	reportCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	type placeResult struct {
		snapshot *placeSnapshot
		err      error
	}
	placeCh := make(chan placeResult, 1)
	fetchPlaceDetails := s.fetchPlaceDetails
	if fetchPlaceDetails == nil && s.places != nil {
		fetchPlaceDetails = s.places.GetPlaceDetails
	}
	go func() {
		if fetchPlaceDetails == nil {
			placeCh <- placeResult{}
			return
		}
		snapshot, err := fetchPlaceDetails(reportCtx, placeID)
		placeCh <- placeResult{snapshot: snapshot, err: err}
	}()

	var (
		snap       *placeSnapshot
		detailsErr error
	)
	select {
	case result := <-placeCh:
		snap = result.snapshot
		detailsErr = result.err
		placeCh = nil
	case <-reportCtx.Done():
	}
	// A source can finish in the same scheduler tick as the deadline. Prefer an
	// already-buffered useful result over dropping it because Done won the select.
	if placeCh != nil {
		select {
		case result := <-placeCh:
			snap = result.snapshot
			detailsErr = result.err
		default:
		}
	}
	if detailsErr != nil {
		s.log.WarnContext(ctx, "seo_places_details_failed", "error", detailsErr, "place_id", placeID)
	}

	place := PlaceDetails{PlaceID: placeID, Source: "places"}
	reviews := []Review(nil)
	photoCount := 0
	hasHours := false
	scoreIn := ScoreInput{}

	if snap != nil {
		place = snap.Details
		reviews = snap.Reviews
		photoCount = snap.PhotoCount
		hasHours = snap.HasHours
		scoreIn.Delivery = snap.Delivery
		scoreIn.Takeout = snap.Takeout
		scoreIn.Reservable = snap.Reservable
		scoreIn.DeliveryKnown = snap.DeliveryKnown
		scoreIn.TakeoutKnown = snap.TakeoutKnown
		scoreIn.ReservableKnown = snap.ReservableKnown
	} else {
		return ReportResponse{}, ErrNotFound
	}

	scoreIn.Reviews = reviews
	scoreIn.PhotoCount = photoCount
	scoreIn.HasHours = hasHours

	var photos []placePhoto
	if snap != nil {
		photos = snap.Photos
	}
	// Public report visuals come only from live Places photo resource names.
	// Historical profile URLs can contain legacy/third-party assets without a
	// current public-use contract and must never cross this boundary.
	place.Media = buildPlaceMedia(place, photos)

	type competitorResult struct {
		candidates  []placeSnapshot
		cuisineType string
		err         error
	}
	var competitorCh chan competitorResult
	fetchNearbyCompetitors := s.fetchNearbyCompetitors
	if fetchNearbyCompetitors == nil && s.places != nil {
		fetchNearbyCompetitors = s.places.SearchNearbyCuisine
	}
	if snap != nil && place.Latitude != nil && place.Longitude != nil && fetchNearbyCompetitors != nil {
		competitorCh = make(chan competitorResult, 1)
		go func() {
			candidates, cuisineType, err := fetchNearbyCompetitors(
				reportCtx,
				place,
				competitorRadiusMeters,
				12,
			)
			competitorCh <- competitorResult{candidates: candidates, cuisineType: cuisineType, err: err}
		}()
	}

	audit := WebsiteAudit{
		Source:         "none",
		MenuEvidence:   noWebsiteMenuEvidence(),
		SocialPresence: noWebsiteSocialPresence(),
	}
	if strings.TrimSpace(place.Website) != "" {
		auditWebsite := s.auditWebsite
		if auditWebsite == nil {
			auditWebsite = AuditWebsite
		}
		if reportCtx.Err() == nil {
			audit = auditWebsite(reportCtx, place.Website, s.llm)
		} else {
			audit = fallbackWebsiteAudit(place.Website, "report budget exhausted")
		}
		scoreIn.WebsiteReachable = audit.Reachable
		if audit.Reachable && audit.Source == "vision" && audit.ViewportCoverage == "desktop_and_mobile" {
			scoreIn.WebsiteQualityKnown = true
			scoreIn.WebsiteQualityScore = audit.QualityScore
		}
		scoreIn.WebsiteReview = audit.Review
		scoreIn.WebsiteScreenshot = audit.Screenshot
		scoreIn.WebsiteMobileScreenshot = audit.MobileScreenshot
		scoreIn.WebsitePageEvidence = audit.PageEvidence
		if strings.TrimSpace(place.Email) == "" && strings.TrimSpace(audit.PublicEmail) != "" {
			place.Email = strings.TrimSpace(audit.PublicEmail)
		}
		if strings.TrimSpace(place.Phone) == "" && strings.TrimSpace(audit.PublicPhone) != "" {
			place.Phone = strings.TrimSpace(audit.PublicPhone)
		}
		if audit.Source == "fallback" {
			s.log.WarnContext(ctx, "seo_website_audit_fallback", "website", place.Website, "source", audit.Source, "reason", audit.FailureReason)
		}
	}
	scoreIn.Place = place
	scoreIn.MenuEvidence = audit.MenuEvidence
	scoreIn.SocialPresence = audit.SocialPresence
	competitorFailed := false
	if competitorCh == nil {
		scoreIn.CompetitorScan = CompetitorScan{
			Status:       "unavailable",
			RadiusKM:     competitorRadiusMeters / 1000,
			Cuisine:      strings.ReplaceAll(competitorCuisineType(place), "_", " "),
			ScoreKind:    "google_visibility",
			CurrentScore: ternaryVisibilityScore(snap),
			Notice:       "Nearby comparison requires live Places coordinates and a specific restaurant listing.",
		}
	} else {
		select {
		case result := <-competitorCh:
			if result.err != nil {
				competitorFailed = true
				s.log.WarnContext(ctx, "seo_competitor_scan_failed", "error", result.err, "place_id", placeID)
				scoreIn.CompetitorScan = CompetitorScan{
					Status:       "partial",
					RadiusKM:     competitorRadiusMeters / 1000,
					Cuisine:      strings.ReplaceAll(result.cuisineType, "_", " "),
					ScoreKind:    "google_visibility",
					CurrentScore: googleVisibilityScore(*snap),
					Notice:       "Nearby same-cuisine listings did not finish, so no competitor claim is shown.",
				}
			} else {
				scoreIn.CompetitorScan = buildCompetitorScan(*snap, result.candidates, result.cuisineType)
			}
		case <-reportCtx.Done():
			competitorFailed = true
			scoreIn.CompetitorScan = CompetitorScan{
				Status:       "partial",
				RadiusKM:     competitorRadiusMeters / 1000,
				Cuisine:      strings.ReplaceAll(competitorCuisineType(place), "_", " "),
				ScoreKind:    "google_visibility",
				CurrentScore: googleVisibilityScore(*snap),
				Notice:       "Nearby same-cuisine listings exceeded the report time budget, so no competitor claim is shown.",
			}
		}
	}

	report := BuildReport(scoreIn)
	summaryResult := SummaryResult{
		Text:   buildDeterministicSummary(place, report, reviews),
		Source: "automated",
	}
	var sumErr error
	if s.summarizer != nil {
		summaryCtx, summaryCancel := context.WithTimeout(reportCtx, reportSummaryBudget)
		summaryResult, sumErr = s.summarizer.Summarize(summaryCtx, place, report, reviews)
		summaryCancel()
	}
	if sumErr != nil {
		s.log.WarnContext(ctx, "seo_summary_failed", "error", sumErr)
		summaryResult = SummaryResult{
			Text:   buildDeterministicSummary(place, report, reviews),
			Source: "automated",
		}
	}
	if strings.TrimSpace(summaryResult.Text) == "" {
		summaryResult.Text = buildDeterministicSummary(place, report, reviews)
		summaryResult.Source = "automated"
	}
	report.AISummary = summaryResult.Text
	report.AnalysisSource = "automated"
	if audit.Source == "vision" || summaryResult.Source == "ai-assisted" {
		report.AnalysisSource = "ai-assisted"
	}
	partial := detailsErr != nil || reportCtx.Err() != nil || competitorFailed
	if strings.TrimSpace(place.Website) != "" && (audit.Source == "fallback" || audit.Source == "none") {
		partial = true
	}
	if audit.Source == "vision" && audit.ViewportCoverage != "desktop_and_mobile" {
		partial = true
	}
	if partial {
		report.AnalysisStatus = "partial"
		report.AnalysisNotice = "Some live signals did not finish within the scan window, so conservative fallback scoring was used."
		if audit.Source == "vision" {
			switch audit.ViewportCoverage {
			case "desktop":
				report.AnalysisNotice = "Only the desktop homepage capture was available; mobile presentation was not visually assessed."
			case "mobile":
				report.AnalysisNotice = "Only the mobile homepage capture was available; desktop presentation was not visually assessed."
			}
		}
	} else {
		report.AnalysisStatus = "complete"
		if audit.Source == "vision" {
			report.AnalysisNotice = "AI-assisted analysis used live listing signals and the captured website homepage."
		} else if summaryResult.Source == "ai-assisted" {
			report.AnalysisNotice = "An AI-assisted summary was generated from the live listing signals; website visuals were not analyzed by AI."
		} else {
			report.AnalysisNotice = "Live listing signals were scored automatically; AI visual analysis was not used for this run."
		}
	}
	report.GeneratedInMS = time.Since(startedAt).Milliseconds()

	payload := ReportResponse{Place: place, Report: report}
	return payload, nil
}

func normalizeSearchLocation(location string) string {
	loc := strings.TrimSpace(location)
	if loc == "" {
		return "Australia"
	}
	return loc
}

// FetchPlacePhoto proxies a Google Places photo through the server.
func (s *Service) FetchPlacePhoto(ctx context.Context, photoName string, maxPx int) ([]byte, string, error) {
	if s == nil || s.places == nil || !s.places.Enabled() {
		return nil, "", ErrNotFound
	}
	return s.places.FetchPhotoMedia(ctx, photoName, maxPx)
}
