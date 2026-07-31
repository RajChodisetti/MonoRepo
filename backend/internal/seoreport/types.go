package seoreport

// PlaceSummary is a lightweight search hit.
type PlaceSummary struct {
	PlaceID         string   `json:"placeId"`
	Name            string   `json:"name"`
	Address         string   `json:"address"`
	Rating          *float64 `json:"rating,omitempty"`
	UserRatingCount *int     `json:"userRatingCount,omitempty"`
	Source          string   `json:"source"`
}

// PlaceDetails is the restaurant identity returned with a report.
type PlaceDetails struct {
	PlaceID           string   `json:"placeId"`
	Name              string   `json:"name"`
	Address           string   `json:"address"`
	Phone             string   `json:"phone,omitempty"`
	Email             string   `json:"email,omitempty"`
	Website           string   `json:"website,omitempty"`
	MapsURI           string   `json:"mapsUri,omitempty"`
	Rating            *float64 `json:"rating,omitempty"`
	UserRatingCount   *int     `json:"userRatingCount,omitempty"`
	PriceLevel        string   `json:"priceLevel,omitempty"`
	BusinessStatus    string   `json:"businessStatus,omitempty"`
	Types             []string `json:"types,omitempty"`
	EditorialSummary  string   `json:"editorialSummary,omitempty"`
	Source            string   `json:"source"`
}

// Review is a single Google review used for scoring/summary.
type Review struct {
	Author  string  `json:"author,omitempty"`
	Text    string  `json:"text,omitempty"`
	Rating  float64 `json:"rating,omitempty"`
	RelativeTime string `json:"relativeTime,omitempty"`
	PublishTime  string `json:"publishTime,omitempty"`
}

// Metric is one scored bucket of the SEO report.
type Metric struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	Score       int     `json:"score"`
	Max         int     `json:"max"`
	Status      string  `json:"status"`
	StatusColor string  `json:"statusColor"`
	Value       float64 `json:"value"`
}

// Issue is a constructive improvement item.
type Issue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CompetitorRow is a simplified comparison row for the UI.
type CompetitorRow struct {
	Rank       string `json:"rank"`
	Name       string `json:"name"`
	Rating     string `json:"rating"`
	Score      string `json:"score"`
	ScoreColor string `json:"scoreColor"`
	Highlight  bool   `json:"highlight"`
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
	Issues               []Issue         `json:"issues"`
	AISummary            string          `json:"aiSummary"`
	EstimatedMonthlyLoss int             `json:"estimatedMonthlyLoss"`
	FullReportLocked     bool            `json:"fullReportLocked"`
	UnlockCTA            string          `json:"unlockCta"`
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
	PlacesEnabled   bool `json:"placesEnabled"`
	InventoryEnabled bool `json:"inventoryEnabled"`
}

// Enrichment holds optional MonoRepo profile data used to fill scoring gaps.
type Enrichment struct {
	Email         string
	Phone         string
	Website       string
	MenuItemCount int
	MenuImageCount int
	HasHours      bool
}
