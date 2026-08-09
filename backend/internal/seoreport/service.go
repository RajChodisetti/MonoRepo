package seoreport

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/profiles"
	llmlib "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/llm"
	"golang.org/x/sync/singleflight"
)

const (
	reportCacheTTL         = 12 * time.Minute
	reportGenerationBudget = 15 * time.Second
	reportSummaryBudget    = 2200 * time.Millisecond
	maxConcurrentReports   = 2
)

var (
	// ErrNotFound indicates the restaurant could not be resolved.
	ErrNotFound = errors.New("seo report restaurant not found")
	// ErrReportBusy fails fast before any provider work when public generation
	// capacity is exhausted. Callers can retry after an in-flight report lands
	// in the cache.
	ErrReportBusy = errors.New("seo report generation is at capacity")
)

var sharedReportGenerationSlots = make(chan struct{}, maxConcurrentReports)

// Service orchestrates Places + inventory enrichment + scoring + summary.
type Service struct {
	places     *PlacesClient
	profiles   profiles.SiteRepository
	summarizer Summarizer
	llm        llmlib.Client
	log        *slog.Logger

	interested    InterestedRepository
	leads         LeadUpserter
	mailer        EmailSender
	appEnv        string
	publicBaseURL string
	publicWebURL  string

	mu    sync.Mutex
	cache map[string]cachedReport

	reportCalls singleflight.Group
	reportSlots chan struct{}

	// Private hooks keep orchestration testable without provider calls. Production
	// constructors always wire the concrete Places, profile, and website clients.
	fetchPlaceDetails func(context.Context, string) (*placeSnapshot, error)
	fetchSiteContent  func(context.Context, string) (profiles.SiteContent, bool)
	auditWebsite      func(context.Context, string, llmlib.Client) WebsiteAudit
	reportBudget      time.Duration
}

type cachedReport struct {
	expires time.Time
	payload ReportResponse
}

// NewService constructs the SEO report service.
func NewService(
	placesCfg config.PlacesConfig,
	profilesRepo profiles.SiteRepository,
	llmClient llmlib.Client,
	log *slog.Logger,
) *Service {
	if log == nil {
		log = slog.Default()
	}
	var summarizer Summarizer = DeterministicSummarizer{}
	if llmClient != nil && llmClient.Enabled() {
		summarizer = LLMSummarizer{Client: llmClient, Fallback: DeterministicSummarizer{}}
	}
	service := &Service{
		places:       NewPlacesClient(placesCfg),
		profiles:     profilesRepo,
		summarizer:   summarizer,
		llm:          llmClient,
		log:          log,
		cache:        make(map[string]cachedReport),
		reportBudget: reportGenerationBudget,
		reportSlots:  sharedReportGenerationSlots,
	}
	service.fetchPlaceDetails = service.places.GetPlaceDetails
	service.fetchSiteContent = service.loadSiteContent
	service.auditWebsite = AuditWebsite
	return service
}

// SearchRestaurants merges inventory + Places search results.
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
			InventoryEnabled: s.profiles != nil,
		},
	}
	if len(query) < 2 {
		return resp, nil
	}

	seen := make(map[string]struct{})
	add := func(item PlaceSummary) {
		key := strings.ToLower(strings.TrimSpace(item.PlaceID))
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		resp.Results = append(resp.Results, item)
	}

	if s.profiles != nil {
		for _, item := range s.searchInventory(ctx, query, location, 5) {
			add(item)
			if len(resp.Results) >= limit {
				return resp, nil
			}
		}
	}

	if s.places.Enabled() {
		places, err := s.places.SearchRestaurants(ctx, query, location, 6)
		if err != nil {
			s.log.WarnContext(ctx, "seo_places_search_failed", "error", err)
		} else {
			for _, item := range places {
				add(item)
				if len(resp.Results) >= limit {
					break
				}
			}
		}
	}

	return resp, nil
}

// GetReport builds (or returns cached) SEO report for a place ID.
func (s *Service) GetReport(ctx context.Context, placeID string) (ReportResponse, error) {
	placeID = sanitizePlaceID(placeID)
	if placeID == "" {
		return ReportResponse{}, ErrNotFound
	}

	if cached, ok := s.getCached(placeID); ok {
		return cached, nil
	}

	// DoChan lets each HTTP caller stop waiting independently while the shared,
	// 15-second-bounded generation finishes for remaining callers and cache.
	result := s.reportCalls.DoChan(placeID, func() (any, error) {
		if cached, ok := s.getCached(placeID); ok {
			return cached, nil
		}
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
	if cached, ok := s.getCached(placeID); ok {
		return cached, nil
	}
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

	// Places and profile inventory are independent. Starting both immediately
	// removes a full network/database waterfall from the critical path.
	type placeResult struct {
		snapshot *placeSnapshot
		err      error
	}
	type contentResult struct {
		content profiles.SiteContent
		ok      bool
	}
	placeCh := make(chan placeResult, 1)
	contentCh := make(chan contentResult, 1)
	fetchPlaceDetails := s.fetchPlaceDetails
	if fetchPlaceDetails == nil && s.places != nil {
		fetchPlaceDetails = s.places.GetPlaceDetails
	}
	fetchSiteContent := s.fetchSiteContent
	if fetchSiteContent == nil {
		fetchSiteContent = s.loadSiteContent
	}

	go func() {
		if fetchPlaceDetails == nil {
			placeCh <- placeResult{}
			return
		}
		snapshot, err := fetchPlaceDetails(reportCtx, placeID)
		placeCh <- placeResult{snapshot: snapshot, err: err}
	}()
	go func() {
		content, ok := fetchSiteContent(reportCtx, placeID)
		contentCh <- contentResult{content: content, ok: ok}
	}()

	var (
		snap       *placeSnapshot
		detailsErr error
		content    profiles.SiteContent
		hasContent bool
	)
	for pending := 2; pending > 0; {
		select {
		case result := <-placeCh:
			snap = result.snapshot
			detailsErr = result.err
			placeCh = nil
			pending--
		case result := <-contentCh:
			content = result.content
			hasContent = result.ok
			contentCh = nil
			pending--
		case <-reportCtx.Done():
			pending = 0
		}
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
	if contentCh != nil {
		select {
		case result := <-contentCh:
			content = result.content
			hasContent = result.ok
		default:
		}
	}
	if detailsErr != nil {
		s.log.WarnContext(ctx, "seo_places_details_failed", "error", detailsErr, "place_id", placeID)
	}

	enrichment := enrichmentFromSiteContent(content, hasContent)

	place := PlaceDetails{PlaceID: placeID, Source: "places"}
	reviews := []Review(nil)
	photoCount := 0
	hasHours := false
	scoreIn := ScoreInput{Enrichment: enrichment}

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
	} else if hasContent && strings.TrimSpace(content.Name) != "" {
		place = placeFromSite(content)
	} else {
		return ReportResponse{}, ErrNotFound
	}

	if enrichment.Phone != "" && place.Phone == "" {
		place.Phone = enrichment.Phone
	}
	if enrichment.Email != "" {
		place.Email = enrichment.Email
	}
	if enrichment.Website != "" && place.Website == "" {
		place.Website = enrichment.Website
	}

	scoreIn.Place = place
	scoreIn.Reviews = reviews
	scoreIn.PhotoCount = photoCount
	scoreIn.HasHours = hasHours || enrichment.HasHours

	var photos []placePhoto
	if snap != nil {
		photos = snap.Photos
	}
	// Public report visuals come only from live Places photo resource names.
	// Historical profile URLs can contain legacy/third-party assets without a
	// current public-use contract and must never cross this boundary.
	place.Media = buildPlaceMedia(place, photos)

	audit := WebsiteAudit{Source: "none"}
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
		if audit.QualityScore > 0 || audit.Source == "social" || audit.Source == "fallback" || audit.Source == "vision" {
			scoreIn.WebsiteQualityKnown = true
			scoreIn.WebsiteQualityScore = audit.QualityScore
			scoreIn.WebsiteReview = audit.Review
			scoreIn.WebsiteScreenshot = audit.Screenshot
			scoreIn.WebsiteMobileScreenshot = audit.MobileScreenshot
		}
		if audit.Source == "fallback" {
			s.log.WarnContext(ctx, "seo_website_audit_fallback", "website", place.Website, "source", audit.Source, "reason", audit.FailureReason)
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
	partial := detailsErr != nil || reportCtx.Err() != nil
	if strings.TrimSpace(place.Website) != "" && (audit.Source == "fallback" || audit.Source == "none") {
		partial = true
	}
	if s.places != nil && s.places.Enabled() && snap == nil && hasContent {
		partial = true
	}
	if partial {
		report.AnalysisStatus = "partial"
		report.AnalysisNotice = "Some live signals did not finish within the scan window, so conservative fallback scoring was used."
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
	s.putCached(placeID, payload)
	return payload, nil
}

func (s *Service) searchInventory(ctx context.Context, query, location string, limit int) []PlaceSummary {
	if s.profiles == nil {
		return nil
	}
	list, err := s.profiles.ListSiteRestaurants(ctx)
	if err != nil {
		s.log.WarnContext(ctx, "seo_inventory_search_failed", "error", err)
		return nil
	}
	q := strings.ToLower(query)
	loc := strings.ToLower(strings.TrimSpace(location))
	filterLoc := loc != "" && loc != "australia"
	out := make([]PlaceSummary, 0, limit)
	for _, item := range list {
		name := strings.TrimSpace(item.Name)
		if name == "" || !strings.Contains(strings.ToLower(name), q) {
			continue
		}
		city := strings.TrimSpace(item.City)
		if filterLoc {
			hay := strings.ToLower(city + " " + name)
			if !strings.Contains(hay, loc) {
				// Allow numeric postcode matches only against city/name text we have.
				continue
			}
		}
		placeID := strings.TrimSpace(item.PlaceID)
		if placeID == "" {
			placeID = item.ID.String()
		}
		out = append(out, PlaceSummary{
			PlaceID: placeID,
			Name:    name,
			Address: city,
			Source:  "monorepo",
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func normalizeSearchLocation(location string) string {
	loc := strings.TrimSpace(location)
	if loc == "" {
		return "Australia"
	}
	return loc
}

func (s *Service) loadEnrichment(ctx context.Context, placeID string) Enrichment {
	content, ok := s.loadSiteContent(ctx, placeID)
	return enrichmentFromSiteContent(content, ok)
}

func enrichmentFromSiteContent(content profiles.SiteContent, ok bool) Enrichment {
	if !ok {
		return Enrichment{}
	}
	return Enrichment{
		Email:          strings.TrimSpace(content.Email),
		Phone:          strings.TrimSpace(content.Phone),
		Website:        strings.TrimSpace(content.Website),
		MenuItemCount:  len(content.MenuItems),
		MenuImageCount: len(content.GalleryImages),
		HasHours:       len(content.Hours) > 2,
	}
}

func (s *Service) loadSiteContent(ctx context.Context, placeID string) (profiles.SiteContent, bool) {
	if s.profiles == nil {
		return profiles.SiteContent{}, false
	}
	content, err := s.profiles.GetSiteContentByPlaceID(ctx, placeID)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.WarnContext(ctx, "seo_site_content_failed", "error", err, "place_id", placeID)
		}
		return profiles.SiteContent{}, false
	}
	return content, true
}

func placeFromSite(content profiles.SiteContent) PlaceDetails {
	place := PlaceDetails{
		PlaceID: firstNonEmpty(content.PlaceID, content.RestaurantID.String()),
		Name:    firstNonEmpty(content.Name, "Restaurant"),
		Address: strings.TrimSpace(strings.Join(filterEmpty([]string{content.Address, content.City, content.State}), ", ")),
		Phone:   content.Phone,
		Email:   content.Email,
		Website: content.Website,
		Source:  "monorepo",
	}
	if content.Rating != nil {
		place.Rating = content.Rating
	}
	if content.ReviewsCount != nil {
		place.UserRatingCount = content.ReviewsCount
	}
	return place
}

// FetchPlacePhoto proxies a Google Places photo through the server.
func (s *Service) FetchPlacePhoto(ctx context.Context, photoName string, maxPx int) ([]byte, string, error) {
	if s == nil || s.places == nil || !s.places.Enabled() {
		return nil, "", ErrNotFound
	}
	return s.places.FetchPhotoMedia(ctx, photoName, maxPx)
}

func filterEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, strings.TrimSpace(v))
		}
	}
	return out
}

func (s *Service) getCached(placeID string) (ReportResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[placeID]
	if !ok || time.Now().After(entry.expires) {
		if ok {
			delete(s.cache, placeID)
		}
		return ReportResponse{}, false
	}
	return entry.payload, true
}

func (s *Service) putCached(placeID string, payload ReportResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Simple bound to avoid unbounded growth.
	if len(s.cache) > 512 {
		now := time.Now()
		for k, v := range s.cache {
			if now.After(v.expires) {
				delete(s.cache, k)
			}
		}
	}
	s.cache[placeID] = cachedReport{expires: time.Now().Add(reportCacheTTL), payload: payload}
}
