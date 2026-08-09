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
	"places.location",
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
	"location",
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
		if lat, lng, ok := placeLatLng(place["location"]); ok {
			summary.Latitude = &lat
			summary.Longitude = &lng
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

// placePhoto is a raw Places photo resource used for media URLs.
type placePhoto struct {
	Name          string
	WidthPx       int
	HeightPx      int
	Attribution   string
	GoogleMapsURI string
}

// placeSnapshot is the internal Places details model used for scoring.
type placeSnapshot struct {
	Details         PlaceDetails
	Reviews         []Review
	Photos          []placePhoto
	PhotoCount      int
	HasHours        bool
	Delivery        bool
	Takeout         bool
	Reservable      bool
	DeliveryKnown   bool
	TakeoutKnown    bool
	ReservableKnown bool
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
	if lat, lng, ok := placeLatLng(place["location"]); ok {
		snap.Details.Latitude = &lat
		snap.Details.Longitude = &lng
	}
	if rating, ok := asFloat(place["rating"]); ok {
		snap.Details.Rating = &rating
	}
	if count, ok := asInt(place["userRatingCount"]); ok {
		snap.Details.UserRatingCount = &count
	}
	if photos, ok := place["photos"].([]any); ok {
		snap.Photos = parsePlacePhotos(photos)
		snap.PhotoCount = len(snap.Photos)
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

// FetchPhotoMedia streams a Places photo binary (keeps the API key server-side).
func (c *PlacesClient) FetchPhotoMedia(ctx context.Context, photoName string, maxPx int) (body []byte, contentType string, err error) {
	if !c.Enabled() {
		return nil, "", fmt.Errorf("places api not configured")
	}
	name := sanitizePhotoName(photoName)
	if name == "" {
		return nil, "", fmt.Errorf("invalid photo name")
	}
	if maxPx < 64 {
		maxPx = 64
	}
	if maxPx > 1600 {
		maxPx = 1600
	}

	endpoint := fmt.Sprintf("%s/%s/media?maxHeightPx=%d&maxWidthPx=%d&skipHttpRedirect=false",
		c.baseURL, name, maxPx, maxPx)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create photo media request: %w", err)
	}
	req.Header.Set("X-Goog-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("places photo media: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("places photo media failed (%d): %s", resp.StatusCode, truncate(string(raw), 160))
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	return raw, ct, nil
}

func parsePlacePhotos(photos []any) []placePhoto {
	out := make([]placePhoto, 0, len(photos))
	for i, item := range photos {
		if i >= 10 {
			break
		}
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(asString(m["name"]))
		if name == "" {
			continue
		}
		p := placePhoto{Name: name, GoogleMapsURI: asString(m["googleMapsUri"])}
		if w, ok := asInt(m["widthPx"]); ok {
			p.WidthPx = w
		}
		if h, ok := asInt(m["heightPx"]); ok {
			p.HeightPx = h
		}
		if attrs, ok := m["authorAttributions"].([]any); ok && len(attrs) > 0 {
			if attr, ok := attrs[0].(map[string]any); ok {
				p.Attribution = asString(attr["displayName"])
			}
		}
		out = append(out, p)
	}
	return out
}

func sanitizePhotoName(photoName string) string {
	name := strings.TrimSpace(photoName)
	name = strings.TrimPrefix(name, "/")
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, " \t\n\r") {
		return ""
	}
	// Expected: places/{placeId}/photos/{photoId}
	if !strings.HasPrefix(name, "places/") || !strings.Contains(name, "/photos/") {
		return ""
	}
	return name
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

func placeLatLng(value any) (lat float64, lng float64, ok bool) {
	m, ok := value.(map[string]any)
	if !ok {
		return 0, 0, false
	}
	lat, okLat := asFloat(m["latitude"])
	lng, okLng := asFloat(m["longitude"])
	if !okLat || !okLng {
		return 0, 0, false
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return 0, 0, false
	}
	return lat, lng, true
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
