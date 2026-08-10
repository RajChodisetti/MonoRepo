package seoreport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func TestGetPlaceDetailsRetainsPrimaryTypeAndGooglePhotoAttribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/places/target" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		mask := r.Header.Get("X-Goog-FieldMask")
		for _, field := range []string{"primaryType", "photos", "googleMapsUri", "attributions"} {
			if !strings.Contains(mask, field) {
				t.Fatalf("field mask %q missing %q", mask, field)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"target",
			"displayName":{"text":"Target Thai"},
			"formattedAddress":"1 Main St",
			"primaryType":"thai_restaurant",
			"types":["thai_restaurant","restaurant"],
			"googleMapsUri":"https://maps.google.com/target",
			"attributions":[
				{"provider":"Local data partner","providerUri":"https://provider.example/place"},
				{"provider":"Second provider"}
			],
			"photos":[{
				"name":"places/target/photos/one",
				"googleMapsUri":"https://maps.google.com/photo/one",
				"flagContentUri":"https://maps.google.com/photo/one/flag",
				"authorAttributions":[
					{"displayName":"Google contributor","uri":"https://maps.google.com/contrib/one","photoUri":"https://images.example/contrib/one"},
					{"displayName":"Restaurant owner","uri":"https://maps.google.com/contrib/two","photoUri":"https://images.example/contrib/two"}
				]
			}],
			"reviews":[{
				"authorAttribution":{"displayName":"Recent guest","uri":"https://maps.google.com/contrib/reviewer","photoUri":"https://images.example/reviewer"},
				"googleMapsUri":"https://maps.google.com/review/one",
				"flagContentUri":"https://maps.google.com/review/one/flag",
				"rating":5,
				"text":{"text":"Wonderful service"},
				"visitDate":{"year":2026,"month":7}
			}]
		}`))
	}))
	defer server.Close()

	client := NewPlacesClient(config.PlacesConfig{APIKey: "test", APIBaseURL: server.URL})
	snapshot, err := client.GetPlaceDetails(context.Background(), "target")
	if err != nil {
		t.Fatalf("GetPlaceDetails: %v", err)
	}
	if snapshot == nil || snapshot.Details.PrimaryType != "thai_restaurant" {
		t.Fatalf("primary type not retained: %#v", snapshot)
	}
	if snapshot.Details.MapsURI != "https://maps.google.com/target" || len(snapshot.Photos) != 1 {
		t.Fatalf("Google source data not retained: %#v", snapshot)
	}
	if got := snapshot.Details.Attributions; len(got) != 2 || got[0].Provider != "Local data partner" ||
		got[0].ProviderURI != "https://provider.example/place" || got[1].Provider != "Second provider" {
		t.Fatalf("place attributions not retained: %#v", got)
	}
	photo := snapshot.Photos[0]
	if photo.Attribution != "Google contributor, Restaurant owner" || photo.GoogleMapsURI != "https://maps.google.com/photo/one" ||
		photo.FlagContentURI != "https://maps.google.com/photo/one/flag" || len(photo.AuthorAttributions) != 2 {
		t.Fatalf("photo attribution not retained: %#v", photo)
	}
	if got := photo.AuthorAttributions[0]; got.DisplayName != "Google contributor" ||
		got.URI != "https://maps.google.com/contrib/one" || got.PhotoURI != "https://images.example/contrib/one" {
		t.Fatalf("first photo author attribution not retained: %#v", got)
	}
	if len(snapshot.Reviews) != 1 {
		t.Fatalf("reviews not retained: %#v", snapshot.Reviews)
	}
	review := snapshot.Reviews[0]
	if review.Author != "Recent guest" || review.AuthorURI != "https://maps.google.com/contrib/reviewer" ||
		review.AuthorPhotoURI != "https://images.example/reviewer" || review.GoogleMapsURI != "https://maps.google.com/review/one" ||
		review.FlagContentURI != "https://maps.google.com/review/one/flag" {
		t.Fatalf("review attribution not retained: %#v", review)
	}
	if review.VisitDate == nil || review.VisitDate.Year != 2026 || review.VisitDate.Month != 7 || review.VisitDate.Day != 0 {
		t.Fatalf("review visit date not retained: %#v", review.VisitDate)
	}
}

func TestSearchRestaurantsPreservesPlaceAttributions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/places:searchText" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		if !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.attributions") {
			t.Fatalf("search field mask=%q", r.Header.Get("X-Goog-FieldMask"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"places":[{
			"id":"search-result",
			"displayName":{"text":"Attributed Restaurant"},
			"formattedAddress":"2 Main St",
			"attributions":[
				{"provider":"Directory partner","providerUri":"https://provider.example/search-result"},
				{"providerUri":"https://provider.example/unnamed"},
				{}
			]
		}]}`))
	}))
	defer server.Close()

	client := NewPlacesClient(config.PlacesConfig{APIKey: "test", APIBaseURL: server.URL, RegionCode: "AU"})
	results, err := client.SearchRestaurants(context.Background(), "restaurant", "Sydney", 6)
	if err != nil {
		t.Fatalf("SearchRestaurants: %v", err)
	}
	if len(results) != 1 || len(results[0].Attributions) != 2 {
		t.Fatalf("search attributions not retained: %#v", results)
	}
	if got := results[0].Attributions[0]; got.Provider != "Directory partner" || got.ProviderURI != "https://provider.example/search-result" {
		t.Fatalf("first search attribution=%#v", got)
	}
	if got := results[0].Attributions[1]; got.Provider != "" || got.ProviderURI != "https://provider.example/unnamed" {
		t.Fatalf("URI-only search attribution=%#v", got)
	}
}

func TestSearchNearbyCuisineUsesTenKilometrePopularityAndFiltersResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/places:searchNearby" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		if !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.primaryType") ||
			!strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.attributions") {
			t.Fatalf("nearby field mask=%q", r.Header.Get("X-Goog-FieldMask"))
		}
		var body struct {
			IncludedTypes  []string `json:"includedTypes"`
			RankPreference string   `json:"rankPreference"`
			MaxResultCount int      `json:"maxResultCount"`
			Location       struct {
				Circle struct {
					Radius float64 `json:"radius"`
				} `json:"circle"`
			} `json:"locationRestriction"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(body.IncludedTypes) != 1 || body.IncludedTypes[0] != "thai_restaurant" ||
			body.RankPreference != "POPULARITY" || body.Location.Circle.Radius != 10_000 || body.MaxResultCount != 12 {
			t.Fatalf("unexpected nearby request: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"places":[
			{"id":"target","displayName":{"text":"Target Thai"},"primaryType":"thai_restaurant","types":["thai_restaurant"],"location":{"latitude":-33.86,"longitude":151.20}},
			{"id":"near","displayName":{"text":"Nearby Thai"},"primaryType":"thai_restaurant","types":["thai_restaurant","restaurant"],"businessStatus":"OPERATIONAL","location":{"latitude":-33.85,"longitude":151.20},"attributions":[{"provider":"Nearby data partner","providerUri":"https://provider.example/near"}]},
			{"id":"closed","displayName":{"text":"Closed Thai"},"primaryType":"thai_restaurant","types":["thai_restaurant"],"businessStatus":"CLOSED_TEMPORARILY","location":{"latitude":-33.85,"longitude":151.20}},
			{"id":"far","displayName":{"text":"Far Thai"},"primaryType":"thai_restaurant","types":["thai_restaurant"],"businessStatus":"OPERATIONAL","location":{"latitude":-33.65,"longitude":151.20}},
			{"id":"italian","displayName":{"text":"Italian Place"},"primaryType":"italian_restaurant","types":["italian_restaurant"],"businessStatus":"OPERATIONAL","location":{"latitude":-33.85,"longitude":151.20}},
			{"id":"nameless","primaryType":"thai_restaurant","types":["thai_restaurant"],"businessStatus":"OPERATIONAL","location":{"latitude":-33.85,"longitude":151.20}}
		]}`))
	}))
	defer server.Close()

	lat, lng := -33.86, 151.20
	client := NewPlacesClient(config.PlacesConfig{APIKey: "test", APIBaseURL: server.URL, RegionCode: "AU"})
	results, cuisine, err := client.SearchNearbyCuisine(context.Background(), PlaceDetails{
		PlaceID:     "target",
		PrimaryType: "thai_restaurant",
		Types:       []string{"thai_restaurant", "restaurant"},
		Latitude:    &lat,
		Longitude:   &lng,
	}, competitorRadiusMeters, 12)
	if err != nil {
		t.Fatalf("SearchNearbyCuisine: %v", err)
	}
	if cuisine != "thai_restaurant" || len(results) != 1 || results[0].Details.Name != "Nearby Thai" {
		t.Fatalf("filtered results cuisine=%q rows=%#v", cuisine, results)
	}
	if results[0].DistanceKM <= 0 || results[0].DistanceKM > 10 {
		t.Fatalf("distance=%f, want inside 10km", results[0].DistanceKM)
	}
	if got := results[0].Details.Attributions; len(got) != 1 || got[0].Provider != "Nearby data partner" || got[0].ProviderURI != "https://provider.example/near" {
		t.Fatalf("nearby attribution not retained: %#v", got)
	}
}
