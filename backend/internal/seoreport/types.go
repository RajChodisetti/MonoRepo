package seoreport

import "encoding/json"

// PlaceSummary is a lightweight search hit.
type PlaceSummary struct {
	PlaceID         string   `json:"placeId"`
	Name            string   `json:"name"`
	Address         string   `json:"address"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
	Rating          *float64 `json:"rating,omitempty"`
	UserRatingCount *int     `json:"userRatingCount,omitempty"`
	Source          string   `json:"source"`
}

// PlaceDetails is the restaurant identity returned with a report.
type PlaceDetails struct {
	PlaceID string `json:"placeId"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
	MapsURI string `json:"mapsUri,omitempty"`
	// Latitude / Longitude from Places location (WGS84) for map pin.
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
	Rating           *float64 `json:"rating,omitempty"`
	UserRatingCount  *int     `json:"userRatingCount,omitempty"`
	PriceLevel       string   `json:"priceLevel,omitempty"`
	BusinessStatus   string   `json:"businessStatus,omitempty"`
	Types            []string `json:"types,omitempty"`
	PrimaryType      string   `json:"primaryType,omitempty"`
	EditorialSummary string   `json:"editorialSummary,omitempty"`
	Source           string   `json:"source"`
	// Media contains only live Google Places photo resource names and attribution links.
	Media *PlaceMedia `json:"media,omitempty"`
}

// MediaCard is one tile in the listing media carousels.
type MediaCard struct {
	Kind      string `json:"kind"` // menu | highlight | photo | latest | video
	Label     string `json:"label"`
	Subtitle  string `json:"subtitle,omitempty"`
	ImageURL  string `json:"imageUrl,omitempty"`
	PhotoName string `json:"photoName,omitempty"`
	Href      string `json:"href,omitempty"`
}

// PlaceMedia groups scraped listing visuals for the report UI.
type PlaceMedia struct {
	MenuAndHighlights []MediaCard `json:"menuAndHighlights,omitempty"`
	PhotosAndVideos   []MediaCard `json:"photosAndVideos,omitempty"`
	MapsURI           string      `json:"mapsUri,omitempty"`
}

// Review is a single Google review used for scoring/summary.
type Review struct {
	Author       string  `json:"author,omitempty"`
	Text         string  `json:"text,omitempty"`
	Rating       float64 `json:"rating,omitempty"`
	RelativeTime string  `json:"relativeTime,omitempty"`
	PublishTime  string  `json:"publishTime,omitempty"`
	// Sentiment is a coarse label for scan UI: positive | mixed | negative.
	Sentiment string `json:"sentiment,omitempty"`
}

// Metric is one scored bucket of the SEO report.
type Metric struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	Score          int      `json:"score"`
	Max            int      `json:"max"`
	Status         string   `json:"status"`
	StatusColor    string   `json:"statusColor"`
	Value          float64  `json:"value"`
	Rationale      string   `json:"rationale,omitempty"`
	Evidence       []string `json:"evidence,omitempty"`
	Recommendation string   `json:"recommendation,omitempty"`
}

// Issue is a constructive improvement item.
type Issue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CompetitorRow is a simplified comparison row for the UI.
type CompetitorRow struct {
	Rank            string   `json:"rank"`
	PlaceID         string   `json:"placeId,omitempty"`
	Name            string   `json:"name"`
	Rating          string   `json:"rating"`
	UserRatingCount int      `json:"userRatingCount,omitempty"`
	Score           string   `json:"score"`
	VisibilityScore int      `json:"visibilityScore"`
	ScoreMax        int      `json:"scoreMax"`
	DistanceKM      float64  `json:"distanceKm"`
	Reasons         []string `json:"reasons,omitempty"`
	ScoreColor      string   `json:"scoreColor"`
	Highlight       bool     `json:"highlight"`
}

// CompetitorScan describes a bounded, same-cuisine Google Places comparison.
// ScoreKind is deliberately not named "Google rank": Nearby Search POPULARITY
// is a discovery order, not evidence of a live Search or Map Pack position.
type CompetitorScan struct {
	Status                   string          `json:"status"`
	RadiusKM                 float64         `json:"radiusKm"`
	Cuisine                  string          `json:"cuisine,omitempty"`
	ScoreKind                string          `json:"scoreKind"`
	SampleSize               int             `json:"sampleSize"`
	CurrentScore             int             `json:"currentScore"`
	CurrentPosition          int             `json:"currentPosition,omitempty"`
	CurrentRestaurantLeading bool            `json:"currentRestaurantLeading"`
	Notice                   string          `json:"notice,omitempty"`
	Rows                     []CompetitorRow `json:"rows,omitempty"`
}

// MenuEvidence is fail-closed evidence from the restaurant's own website.
// Google Places photos are intentionally excluded because the official Photo
// resource does not expose Google Maps UI categories such as "Menu".
type MenuEvidence struct {
	Status            string `json:"status"` // present | not_found | unknown
	HasWebsiteLink    bool   `json:"hasWebsiteLink"`
	HasStructuredData bool   `json:"hasStructuredData"`
	MenuURL           string `json:"menuUrl,omitempty"`
	Source            string `json:"source,omitempty"`
	Rationale         string `json:"rationale"`
}

// SocialProfile is a canonical public profile linked from the restaurant's
// official website or used as the listing's website URI.
type SocialProfile struct {
	Platform string `json:"platform"`
	Handle   string `json:"handle,omitempty"`
	URL      string `json:"url"`
	Source   string `json:"source"`
}

// SocialPresence is a listing-completeness signal worth at most three points.
type SocialPresence struct {
	Status    string          `json:"status"` // present | not_found | unknown
	Score     int             `json:"score"`
	Max       int             `json:"max"`
	Profiles  []SocialProfile `json:"profiles,omitempty"`
	Rationale string          `json:"rationale,omitempty"`
}

// Report is the scored SEO report payload.
type Report struct {
	RestaurantName       string          `json:"restaurantName"`
	Address              string          `json:"address"`
	OverallScore         int             `json:"overallScore"`
	OverallLabel         string          `json:"overallLabel"`
	OverallColor         string          `json:"overallColor"`
	Metrics              []Metric        `json:"metrics"`
	Competitors          []CompetitorRow `json:"competitors"`
	CompetitorScan       CompetitorScan  `json:"competitorScan"`
	MenuEvidence         MenuEvidence    `json:"menuEvidence"`
	SocialPresence       SocialPresence  `json:"socialPresence"`
	Issues               []Issue         `json:"issues"`
	AISummary            string          `json:"aiSummary"`
	EstimatedMonthlyLoss int             `json:"estimatedMonthlyLoss"`
	FullReportLocked     bool            `json:"fullReportLocked"`
	UnlockCTA            string          `json:"unlockCta"`
	// Website visual audit (screenshot + AI design review).
	WebsiteScreenshot       string `json:"websiteScreenshot,omitempty"`
	WebsiteMobileScreenshot string `json:"websiteMobileScreenshot,omitempty"`
	WebsiteQualityScore     int    `json:"websiteQualityScore,omitempty"`
	WebsiteReview           string `json:"websiteReview,omitempty"`
	// AnalysisSource is "ai-assisted" only when an LLM actually contributed.
	// Rule-based fallbacks are reported as "automated" so the public UI does
	// not imply that AI ran when the provider is disabled or timed out.
	AnalysisSource string `json:"analysisSource"`
	// AnalysisStatus is "complete" or "partial". Partial reports remain useful
	// but use conservative scoring when a live dependency misses the time budget.
	AnalysisStatus string `json:"analysisStatus"`
	AnalysisNotice string `json:"analysisNotice,omitempty"`
	GeneratedInMS  int64  `json:"generatedInMs"`
	// RecentReviews are live Google reviews shown during the scan experience.
	RecentReviews []Review `json:"recentReviews,omitempty"`
}

// RedactLockedReport removes verification-gated data on the server. The full
// values remain in the cached in-memory Report so the existing unlock flow can
// reveal them simply by setting FullReportLocked=false.
func RedactLockedReport(report Report) Report {
	if !report.FullReportLocked {
		return report
	}
	report.Competitors = nil
	report.CompetitorScan.Rows = nil
	report.Issues = nil
	report.AISummary = ""
	report.WebsiteReview = ""
	report.MenuEvidence.MenuURL = ""
	report.MenuEvidence.Rationale = ""
	report.SocialPresence.Profiles = nil
	report.SocialPresence.Rationale = ""
	for index := range report.Metrics {
		report.Metrics[index].Rationale = ""
		report.Metrics[index].Evidence = nil
		report.Metrics[index].Recommendation = ""
	}
	return report
}

// MarshalJSON enforces locked-report redaction at the serialization boundary.
// This prevents CSS blur or client code from becoming the access-control layer.
func (report Report) MarshalJSON() ([]byte, error) {
	view := RedactLockedReport(report)
	type reportJSON Report
	return json.Marshal(reportJSON(view))
}

// ReportResponse is the public API envelope for a place report.
type ReportResponse struct {
	Place  PlaceDetails `json:"place"`
	Report Report       `json:"report"`
}

// SearchResponse is the public API envelope for restaurant search.
type SearchResponse struct {
	Results []PlaceSummary `json:"results"`
	Meta    SearchMeta     `json:"meta"`
}

// SearchMeta describes which backends contributed to search.
type SearchMeta struct {
	PlacesEnabled    bool `json:"placesEnabled"`
	InventoryEnabled bool `json:"inventoryEnabled"`
}

// Enrichment holds optional MonoRepo profile data used to fill scoring gaps.
type Enrichment struct {
	Email          string
	Phone          string
	Website        string
	MenuItemCount  int
	MenuImageCount int
	HasHours       bool
}
