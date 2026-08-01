package seoreport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

const (
	defaultPlacesBase = "https://places.googleapis.com/v1"
	placesHTTPTimeout = 12 * time.Second
)

var searchFieldMask = strings.Join([]string{
	"places.id",
	"places.displayName",
	"places.formattedAddress",
	"places.rating",
	"places.userRatingCount",
	"places.businessStatus",
}, ",")

var detailFieldMask = strings.Join([]string{
	"id",
	"displayName",
	"formattedAddress",
	"nationalPhoneNumber",
	"internationalPhoneNumber",
	"websiteUri",
	"googleMapsUri",
	"rating",
	"userRatingCount",
	"priceLevel",
	"types",
	"primaryType",
	"businessStatus",
	"editorialSummary",
	"reviews",
	"photos",
	"regularOpeningHours",
	"delivery",
	"takeout",
	"reservable",
}, ",")

// PlacesClient talks to Google Places API (New).
type PlacesClient struct {
	apiKey     string
	baseURL    string
	regionCode string
	httpClient *http.Client
}

// NewPlacesClient builds a Places client from config. Empty API key disables remote calls.
func NewPlacesClient(cfg config.PlacesConfig) *PlacesClient {
	base := strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	if base == "" {
		base = defaultPlacesBase
	}
	region := strings.ToUpper(strings.TrimSpace(cfg.RegionCode))
	if region == "" {
		region = "AU"
	}
	return &PlacesClient{
		apiKey:     strings.TrimSpace(cfg.APIKey),
		baseURL:    base,
		regionCode: region,
		httpClient: &http.Client{Timeout: placesHTTPTimeout},
	}
}

// Enabled reports whether a Places API key is configured.
func (c *PlacesClient) Enabled() bool {
	return c != nil && c.apiKey != ""
}

// SearchRestaurants runs Places text search for restaurants near a location.
// location may be a city name or postcode; empty defaults to Australia.
func (c *PlacesClient) SearchRestaurants(ctx context.Context, query, location string, limit int) ([]PlaceSummary, error) {
	if !c.Enabled() || strings.TrimSpace(query) == "" {
		return nil, nil
	}
	if limit < 1 {
		limit = 6
	}
	if limit > 10 {
		limit = 10
	}

	body := map[string]any{
		"textQuery":           buildPlacesTextQuery(query, location, c.regionCode),
		"includedType":        "restaurant",
		"strictTypeFiltering": true,
		"languageCode":        "en",
		"regionCode":          c.regionCode,
		"pageSize":            limit,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal places search body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/places:searchText", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create places search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", searchFieldMask)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("places searchText: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("places searchText failed (%d): %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed struct {
		Places []map[string]any `json:"places"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode places searchText: %w", err)
	}

	out := make([]PlaceSummary, 0, len(parsed.Places))
	for _, place := range parsed.Places {
		if strings.EqualFold(asString(place["businessStatus"]), "CLOSED_PERMANENTLY") {
			continue
		}
		placeID := strings.TrimSpace(asString(place["id"]))
		if placeID == "" {
			continue
		}
		summary := PlaceSummary{
			PlaceID: placeID,
			Name:    localizedText(place["displayName"]),
			Address: asString(place["formattedAddress"]),
			Source:  "places",
		}
		if summary.Name == "" {
			summary.Name = "Restaurant"
		}
		if rating, ok := asFloat(place["rating"]); ok {
			summary.Rating = &rating
		}
		if count, ok := asInt(place["userRatingCount"]); ok {
			summary.UserRatingCount = &count
		}
		out = append(out, summary)
	}
	return out, nil
}

// placeSnapshot is the internal Places details model used for scoring.
type placeSnapshot struct {
	Details          PlaceDetails
	Reviews          []Review
	PhotoCount       int
	HasHours         bool
	Delivery         bool
	Takeout          bool
	Reservable       bool
	DeliveryKnown    bool
	TakeoutKnown     bool
	ReservableKnown  bool
}

// GetPlaceDetails loads Places details including reviews and order attrs.
func (c *PlacesClient) GetPlaceDetails(ctx context.Context, placeID string) (*placeSnapshot, error) {
	if !c.Enabled() {
		return nil, nil
	}
	id := sanitizePlaceID(placeID)
	if id == "" {
		return nil, nil
	}

	endpoint := c.baseURL + "/places/" + url.PathEscape(id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create places details request: %w", err)
	}
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", detailFieldMask)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("places details: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("places details failed (%d): %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var place map[string]any
	if err := json.Unmarshal(raw, &place); err != nil {
		return nil, fmt.Errorf("decode places details: %w", err)
	}

	snap := &placeSnapshot{
		Details: PlaceDetails{
			PlaceID:          firstNonEmpty(asString(place["id"]), id),
			Name:             firstNonEmpty(localizedText(place["displayName"]), "Restaurant"),
			Address:          asString(place["formattedAddress"]),
			Phone:            firstNonEmpty(asString(place["nationalPhoneNumber"]), asString(place["internationalPhoneNumber"])),
			Website:          asString(place["websiteUri"]),
			MapsURI:          asString(place["googleMapsUri"]),
			PriceLevel:       asString(place["priceLevel"]),
			BusinessStatus:   asString(place["businessStatus"]),
			Types:            asStringSlice(place["types"]),
			EditorialSummary: localizedText(place["editorialSummary"]),
			Source:           "places",
		},
	}
	if rating, ok := asFloat(place["rating"]); ok {
		snap.Details.Rating = &rating
	}
	if count, ok := asInt(place["userRatingCount"]); ok {
		snap.Details.UserRatingCount = &count
	}
	if photos, ok := place["photos"].([]any); ok {
		snap.PhotoCount = len(photos)
	}
	if hours, ok := place["regularOpeningHours"].(map[string]any); ok && len(hours) > 0 {
		snap.HasHours = true
	}
	if v, ok := place["delivery"].(bool); ok {
		snap.Delivery = v
		snap.DeliveryKnown = true
	}
	if v, ok := place["takeout"].(bool); ok {
		snap.Takeout = v
		snap.TakeoutKnown = true
	}
	if v, ok := place["reservable"].(bool); ok {
		snap.Reservable = v
		snap.ReservableKnown = true
	}

	if reviews, ok := place["reviews"].([]any); ok {
		for i, item := range reviews {
			if i >= 5 {
				break
			}
			revMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			author := ""
			if attr, ok := revMap["authorAttribution"].(map[string]any); ok {
				author = asString(attr["displayName"])
			}
			text := localizedText(revMap["originalText"])
			if text == "" {
				text = localizedText(revMap["text"])
			}
			rating, _ := asFloat(revMap["rating"])
			snap.Reviews = append(snap.Reviews, Review{
				Author:       author,
				Text:         text,
				Rating:       rating,
				RelativeTime: asString(revMap["relativePublishTimeDescription"]),
				PublishTime:  asString(revMap["publishTime"]),
			})
		}
	}

	return snap, nil
}

func sanitizePlaceID(placeID string) string {
	id := strings.TrimSpace(placeID)
	id = strings.TrimPrefix(id, "places/")
	id = strings.TrimSpace(id)
	if id == "" || strings.ContainsAny(id, " \t\n\r") {
		return ""
	}
	return id
}

func localizedText(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		return strings.TrimSpace(asString(v["text"]))
	default:
		return ""
	}
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return ""
	}
}

func asStringSlice(value any) []string {
	arr, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s := asString(item)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func asFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func asInt(value any) (int, bool) {
	f, ok := asFloat(value)
	if !ok {
		return 0, false
	}
	return int(f), true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// buildPlacesTextQuery scopes restaurant search to a city/postcode (AU default).
func buildPlacesTextQuery(query, location, regionCode string) string {
	q := strings.TrimSpace(query)
	loc := strings.TrimSpace(location)
	if loc == "" {
		loc = "Australia"
	}
	lower := strings.ToLower(loc)
	region := strings.ToUpper(strings.TrimSpace(regionCode))
	if region == "" {
		region = "AU"
	}
	if region == "AU" && lower != "australia" && !strings.Contains(lower, "australia") {
		loc = loc + ", Australia"
	}
	return q + " restaurant in " + loc
}
