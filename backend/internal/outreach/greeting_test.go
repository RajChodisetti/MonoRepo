package outreach

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func float64Pointer(value float64) *float64 { return &value }
func intPointer(value int) *int             { return &value }

func TestRenderGreeting01DeterministicFallbacks(t *testing.T) {
	tests := []struct {
		name      string
		facts     GreetingFacts
		want      string
		factsUsed []string
	}{
		{
			name: "owner cuisine city and qualifying rating",
			facts: GreetingFacts{
				RestaurantName: "Spice Garden", OwnerFirstName: "Maya Patel",
				GooglePlaceID: "place-1", ScrapeStatus: "success", City: "Plano",
				Cuisines: json.RawMessage(`["Indian Restaurant"]`),
				Rating:   float64Pointer(4.7), ReviewCount: intPointer(380),
			},
			want:      "Morning Maya,\n\nI noticed Spice Garden serves some of the most popular Indian dishes in Plano with a 4.7-star rating across over 380 reviews.",
			factsUsed: []string{"owner_first_name", "restaurant_name", "cuisine", "city", "rating", "review_count"},
		},
		{
			name: "city only",
			facts: GreetingFacts{
				RestaurantName: "Spice Garden", GooglePlaceID: "place-1",
				ScrapeStatus: "success", City: "Plano",
				Cuisines: json.RawMessage(`["Indian"]`),
			},
			want:      "Morning Spice Garden team,\n\nI noticed Spice Garden is serving guests in Plano.",
			factsUsed: []string{"restaurant_name", "city"},
		},
		{
			name: "cuisine only and first safe restaurant cuisine",
			facts: GreetingFacts{
				RestaurantName: "Spice Garden", GooglePlaceID: "place-1", ScrapeStatus: "success",
				City: "unknown", Cuisines: json.RawMessage(`[42, "Indian", "{{bad}} Restaurant", "South Indian Restaurant", "Thai Restaurant"]`),
			},
			want:      "Morning Spice Garden team,\n\nI noticed Spice Garden serves popular South Indian dishes.",
			factsUsed: []string{"restaurant_name", "cuisine"},
		},
		{
			name: "verified profile with low rating",
			facts: GreetingFacts{
				RestaurantName: "Corner Cafe", GooglePlaceID: "place-2", ScrapeStatus: "success",
				City: "Austin", Rating: float64Pointer(3.9), ReviewCount: intPointer(500),
			},
			want:      "Morning Corner Cafe team,\n\nI noticed Corner Cafe is serving guests in Austin.",
			factsUsed: []string{"restaurant_name", "city"},
		},
		{
			name: "verified profile with too few reviews",
			facts: GreetingFacts{
				RestaurantName: "Corner Cafe", GooglePlaceID: "place-2", ScrapeStatus: "success",
				Rating: float64Pointer(4.9), ReviewCount: intPointer(9),
			},
			want:      "Morning Corner Cafe team,\n\nI noticed Corner Cafe has been building a local following.",
			factsUsed: []string{"restaurant_name"},
		},
		{
			name: "missing Google profile ignores listing fields",
			facts: GreetingFacts{
				RestaurantName: "Corner Cafe", ScrapeStatus: "success", City: "Austin",
				Cuisines: json.RawMessage(`["Cafe Restaurant"]`),
				Rating:   float64Pointer(5.0), ReviewCount: intPointer(100),
			},
			want:      "Morning Corner Cafe team,\n\nI noticed Corner Cafe has been building a local following.",
			factsUsed: []string{"restaurant_name"},
		},
		{
			name: "failed scrape ignores listing fields",
			facts: GreetingFacts{
				RestaurantName: "Corner Cafe", GooglePlaceID: "place-2", ScrapeStatus: "failed",
				City: "Austin", Cuisines: json.RawMessage(`["Cafe Restaurant"]`),
				Rating: float64Pointer(5.0), ReviewCount: intPointer(100),
			},
			want:      "Morning Corner Cafe team,\n\nI noticed Corner Cafe has been building a local following.",
			factsUsed: []string{"restaurant_name"},
		},
		{
			name: "malformed multiline and oversized fields are discarded",
			facts: GreetingFacts{
				RestaurantName: "Unsafe\nName", OwnerFirstName: "Maya\nInjected",
				GooglePlaceID: "place-3", ScrapeStatus: "success", City: "Plano\nTexas",
				Cuisines: json.RawMessage(`["` + strings.Repeat("A", 101) + ` Restaurant"]`),
			},
			want:      "Morning restaurant team,\n\nI noticed restaurant has been building a local following.",
			factsUsed: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RenderGreeting01(test.facts)
			if got.Greeting01 != test.want {
				t.Fatalf("Greeting01 = %q, want %q", got.Greeting01, test.want)
			}
			if strings.Join(got.FactsUsed, ",") != strings.Join(test.factsUsed, ",") {
				t.Fatalf("FactsUsed = %#v, want %#v", got.FactsUsed, test.factsUsed)
			}
			lines := strings.Split(got.Greeting01, "\n")
			if len(lines) != 3 || lines[0] == "" || lines[1] != "" || lines[2] == "" {
				t.Fatalf("Greeting01 must contain one salutation and one personalized line: %#v", lines)
			}
			if !strings.HasPrefix(lines[0], "Morning ") || !strings.HasSuffix(lines[0], ",") || strings.Contains(got.Greeting01, "{{") {
				t.Fatalf("Greeting01 contains an unsafe salutation or unresolved tag: %q", got.Greeting01)
			}
		})
	}
}

func TestRenderGreeting01EveryOwnerCityCuisineRatingCombination(t *testing.T) {
	for combination := 0; combination < 16; combination++ {
		combination := combination
		t.Run(strconv.Itoa(combination), func(t *testing.T) {
			facts := GreetingFacts{
				RestaurantName: "Spice Garden",
				GooglePlaceID:  "place-1",
				ScrapeStatus:   "success",
				Cuisines:       json.RawMessage(`[]`),
			}
			if combination&1 != 0 {
				facts.OwnerFirstName = "Maya"
			}
			if combination&2 != 0 {
				facts.City = "Plano"
			}
			if combination&4 != 0 {
				facts.Cuisines = json.RawMessage(`["Indian Restaurant"]`)
			}
			if combination&8 != 0 {
				facts.Rating = float64Pointer(4.7)
				facts.ReviewCount = intPointer(380)
			}

			got := RenderGreeting01(facts)
			if combination&1 != 0 && !strings.HasPrefix(got.Greeting01, "Morning Maya,") {
				t.Fatalf("owner combination did not use owner salutation: %q", got.Greeting01)
			}
			if combination&1 == 0 && !strings.HasPrefix(got.Greeting01, "Morning Spice Garden team,") {
				t.Fatalf("owner fallback did not use restaurant team: %q", got.Greeting01)
			}
			if combination&2 != 0 && !strings.Contains(got.Greeting01, "Plano") {
				t.Fatalf("city combination did not use city: %q", got.Greeting01)
			}
			if combination&2 == 0 && strings.Contains(got.Greeting01, "Plano") {
				t.Fatalf("city appeared without a city fact: %q", got.Greeting01)
			}
			if combination&4 != 0 && !strings.Contains(got.Greeting01, "Indian dishes") {
				t.Fatalf("cuisine combination did not use cuisine: %q", got.Greeting01)
			}
			if combination&4 == 0 && strings.Contains(got.Greeting01, "Indian") {
				t.Fatalf("cuisine appeared without a cuisine fact: %q", got.Greeting01)
			}
			if combination&8 != 0 && !strings.Contains(got.Greeting01, "4.7-star rating across over 380 reviews") {
				t.Fatalf("rating combination did not use rating/review count: %q", got.Greeting01)
			}
			if combination&8 == 0 && strings.Contains(got.Greeting01, "star rating") {
				t.Fatalf("rating appeared without qualifying rating facts: %q", got.Greeting01)
			}
		})
	}
}
