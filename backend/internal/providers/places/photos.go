// Package places contains server-side Google Places adapters.
package places

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

var (
	ErrNotConfigured = errors.New("google places photos are not configured")
	photoNamePattern = regexp.MustCompile(`^places/[^/?#]+/photos/[^/?#]+$`)
)

type Attribution struct {
	DisplayName string `json:"display_name"`
	URI         string `json:"uri,omitempty"`
}

// Photo is a freshly resolved Places photo. URL is intentionally not stored in
// PostgreSQL because Google photo resource names and media URLs can expire.
type Photo struct {
	URL                string        `json:"url"`
	WidthPx            int           `json:"width_px,omitempty"`
	HeightPx           int           `json:"height_px,omitempty"`
	AuthorAttributions []Attribution `json:"author_attributions"`
}

type PhotoResolver interface {
	ListPhotoURLs(ctx context.Context, placeID string) ([]Photo, error)
}

type Client struct {
	cfg        config.PlacesConfig
	httpClient *http.Client
}

func NewClient(cfg config.PlacesConfig) *Client {
	return NewClientWithHTTP(cfg, &http.Client{Timeout: cfg.Timeout})
}

func NewClientWithHTTP(cfg config.PlacesConfig, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{cfg: cfg, httpClient: httpClient}
}

func (client *Client) ListPhotoURLs(ctx context.Context, placeID string) ([]Photo, error) {
	placeID = strings.TrimSpace(placeID)
	if strings.TrimSpace(client.cfg.APIKey) == "" {
		return nil, ErrNotConfigured
	}
	if placeID == "" {
		return []Photo{}, nil
	}

	resources, err := client.listPhotoResources(ctx, placeID)
	if err != nil {
		return nil, err
	}
	limit := min(client.cfg.PhotoLimit, len(resources))
	photos := make([]Photo, 0, limit)
	for _, resource := range resources[:limit] {
		photoURL, resolveErr := client.resolvePhotoURL(ctx, resource.Name)
		if resolveErr != nil {
			return nil, resolveErr
		}
		photos = append(photos, Photo{
			URL:                photoURL,
			WidthPx:            resource.WidthPx,
			HeightPx:           resource.HeightPx,
			AuthorAttributions: normalizeAttributions(resource.AuthorAttributions),
		})
	}
	return photos, nil
}

type photoResource struct {
	Name               string `json:"name"`
	WidthPx            int    `json:"widthPx"`
	HeightPx           int    `json:"heightPx"`
	AuthorAttributions []struct {
		DisplayName string `json:"displayName"`
		URI         string `json:"uri"`
	} `json:"authorAttributions"`
}

func (client *Client) listPhotoResources(ctx context.Context, placeID string) ([]photoResource, error) {
	endpoint := strings.TrimRight(client.cfg.APIBaseURL, "/") + "/places/" + url.PathEscape(placeID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create Places details request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Goog-Api-Key", client.cfg.APIKey)
	req.Header.Set("X-Goog-FieldMask", "photos")

	var payload struct {
		Photos []photoResource `json:"photos"`
	}
	if err := client.doJSON(req, &payload); err != nil {
		return nil, fmt.Errorf("refresh Places photo resources: %w", err)
	}

	resources := make([]photoResource, 0, len(payload.Photos))
	seen := make(map[string]struct{}, len(payload.Photos))
	for _, resource := range payload.Photos {
		resource.Name = strings.TrimSpace(resource.Name)
		if !photoNamePattern.MatchString(resource.Name) {
			continue
		}
		if _, exists := seen[resource.Name]; exists {
			continue
		}
		seen[resource.Name] = struct{}{}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (client *Client) resolvePhotoURL(ctx context.Context, photoName string) (string, error) {
	if !photoNamePattern.MatchString(photoName) {
		return "", errors.New("invalid Places photo resource name")
	}
	endpoint, err := url.Parse(strings.TrimRight(client.cfg.APIBaseURL, "/") + "/" + photoName + "/media")
	if err != nil {
		return "", fmt.Errorf("create Places photo media URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("maxWidthPx", fmt.Sprintf("%d", client.cfg.PhotoMaxWidth))
	query.Set("skipHttpRedirect", "true")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create Places photo media request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Goog-Api-Key", client.cfg.APIKey)

	var payload struct {
		PhotoURI string `json:"photoUri"`
	}
	if err := client.doJSON(req, &payload); err != nil {
		return "", fmt.Errorf("resolve Places photo media: %w", err)
	}
	photoURI := strings.TrimSpace(payload.PhotoURI)
	parsed, err := url.Parse(photoURI)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", errors.New("Places photo media returned an invalid URL")
	}
	return photoURI, nil
}

func (client *Client) doJSON(req *http.Request, destination any) error {
	response, err := client.httpClient.Do(req)
	if err != nil {
		return errors.New("Google Places request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Google Places HTTP %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(destination); err != nil {
		return errors.New("Google Places returned invalid JSON")
	}
	return nil
}

func normalizeAttributions(source []struct {
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
}) []Attribution {
	result := make([]Attribution, 0, len(source))
	for _, item := range source {
		uri := strings.TrimSpace(item.URI)
		if strings.HasPrefix(uri, "//") {
			uri = "https:" + uri
		}
		result = append(result, Attribution{
			DisplayName: strings.TrimSpace(item.DisplayName),
			URI:         uri,
		})
	}
	return result
}

var _ PhotoResolver = (*Client)(nil)
