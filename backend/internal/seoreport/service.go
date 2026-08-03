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
)

const reportCacheTTL = 12 * time.Minute

// ErrNotFound indicates the restaurant could not be resolved.
var ErrNotFound = errors.New("seo report restaurant not found")

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
	return &Service{
		places:     NewPlacesClient(placesCfg),
		profiles:   profilesRepo,
		summarizer: summarizer,
		llm:        llmClient,
		log:        log,
		cache:      make(map[string]cachedReport),
	}
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

	var snap *placeSnapshot
	var err error
	if s.places.Enabled() {
		snap, err = s.places.GetPlaceDetails(ctx, placeID)
		if err != nil {
			s.log.WarnContext(ctx, "seo_places_details_failed", "error", err, "place_id", placeID)
		}
	}

	enrichment := s.loadEnrichment(ctx, placeID)

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
	} else if content, ok := s.loadSiteContent(ctx, placeID); ok && strings.TrimSpace(content.Name) != "" {
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

	var siteMedia *profilesSiteMedia
	if content, ok := s.loadSiteContent(ctx, placeID); ok {
		siteMedia = siteMediaFromContent(content)
	}
	var photos []placePhoto
	if snap != nil {
		photos = snap.Photos
	}
	place.Media = buildPlaceMedia(place, photos, siteMedia)

	if strings.TrimSpace(place.Website) != "" {
		audit := AuditWebsite(ctx, place.Website, s.llm)
		if audit.QualityScore > 0 || audit.Source == "social" || audit.Source == "fallback" || audit.Source == "vision" {
			scoreIn.WebsiteQualityKnown = true
			scoreIn.WebsiteQualityScore = audit.QualityScore
			scoreIn.WebsiteReview = audit.Review
			scoreIn.WebsiteScreenshot = audit.Screenshot
		}
		if audit.Source == "fallback" {
			s.log.WarnContext(ctx, "seo_website_audit_fallback", "website", place.Website, "source", audit.Source)
		}
	}

	report := BuildReport(scoreIn)
	summary, sumErr := s.summarizer.Summarize(ctx, place, report, reviews)
	if sumErr != nil {
		s.log.WarnContext(ctx, "seo_summary_failed", "error", sumErr)
		summary = buildDeterministicSummary(place, report, reviews)
	}
	report.AISummary = summary

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

func siteMediaFromContent(content profiles.SiteContent) *profilesSiteMedia {
	out := &profilesSiteMedia{
		MenuItems: make([]siteMenuMedia, 0, len(content.MenuItems)),
		Gallery:   make([]siteGalleryMedia, 0, len(content.GalleryImages)),
	}
	for _, item := range content.MenuItems {
		img := strings.TrimSpace(item.ImageURL)
		if img == "" {
			continue
		}
		out.MenuItems = append(out.MenuItems, siteMenuMedia{
			Name:     strings.TrimSpace(item.Name),
			ImageURL: img,
		})
	}
	for _, g := range content.GalleryImages {
		out.Gallery = append(out.Gallery, siteGalleryMedia{
			Title:        g.Title,
			URL:          g.URL,
			ThumbnailURL: g.ThumbnailURL,
		})
	}
	return out
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
