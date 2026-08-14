package outreach

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxGreetingRestaurantNameLength = 160
	maxGreetingOwnerNameLength      = 80
	maxGreetingCityLength           = 80
	maxGreetingCuisineLength        = 100
)

// GreetingFacts is the complete, non-inferred input to the deterministic
// first-email greeting renderer. Listing fields are considered only when the
// profile has both a Google place id and a successful scrape status.
type GreetingFacts struct {
	RestaurantName string
	OwnerFirstName string
	GooglePlaceID  string
	ScrapeStatus   string
	City           string
	Cuisines       json.RawMessage
	Rating         *float64
	ReviewCount    *int
}

type GreetingRender struct {
	Greeting01 string   `json:"greeting01"`
	FactsUsed  []string `json:"facts_used"`
}

// RenderGreeting01 returns exactly one salutation, one blank separator, and
// two deterministic greeting lines. Optional or unsafe facts select a fallback
// and never make delivery ineligible.
func RenderGreeting01(facts GreetingFacts) GreetingRender {
	restaurantName := safeGreetingValue(facts.RestaurantName, maxGreetingRestaurantNameLength)
	factsUsed := make([]string, 0, 6)
	if restaurantName == "" {
		restaurantName = "restaurant"
	} else {
		factsUsed = append(factsUsed, "restaurant_name")
	}

	ownerFirstName := safeOwnerFirstName(facts.OwnerFirstName)
	salutation := "Hi " + restaurantName + " team,"
	if ownerFirstName != "" {
		salutation = "Hi " + ownerFirstName + ","
		factsUsed = append([]string{"owner_first_name"}, factsUsed...)
	}

	verifiedListing := strings.TrimSpace(facts.GooglePlaceID) != "" &&
		strings.TrimSpace(facts.ScrapeStatus) == "success"
	city := ""
	cuisine := ""
	if verifiedListing {
		city = safeGreetingValue(facts.City, maxGreetingCityLength)
		cuisine = firstSafeRestaurantCuisine(facts.Cuisines)
	}

	firstLine := "I came across " + restaurantName + " while looking at local restaurants."
	switch {
	case cuisine != "" && city != "":
		firstLine = "I came across " + restaurantName + " while looking at " + cuisine + " restaurants in " + city + "."
		factsUsed = append(factsUsed, "cuisine", "city")
	case city != "":
		firstLine = "I came across " + restaurantName + " while looking at restaurants in " + city + "."
		factsUsed = append(factsUsed, "city")
	case cuisine != "":
		firstLine = "I came across " + restaurantName + " while looking at " + cuisine + " restaurants."
		factsUsed = append(factsUsed, "cuisine")
	}

	secondLine := "I thought it was worth reaching out directly to your team."
	if verifiedListing && facts.Rating != nil && facts.ReviewCount != nil &&
		!math.IsNaN(*facts.Rating) && !math.IsInf(*facts.Rating, 0) &&
		*facts.Rating >= 4.0 && *facts.Rating <= 5.0 && *facts.ReviewCount >= 10 {
		secondLine = "Your Google listing currently shows a " +
			strconv.FormatFloat(*facts.Rating, 'f', 1, 64) +
			"-star rating across " + strconv.Itoa(*facts.ReviewCount) + " reviews."
		factsUsed = append(factsUsed, "rating", "review_count")
	}

	return GreetingRender{
		Greeting01: salutation + "\n\n" + firstLine + "\n" + secondLine,
		FactsUsed:  factsUsed,
	}
}

func firstSafeRestaurantCuisine(raw json.RawMessage) string {
	var values []any
	if len(raw) == 0 || json.Unmarshal(raw, &values) != nil {
		return ""
	}
	for _, value := range values {
		candidate, ok := value.(string)
		if !ok {
			continue
		}
		candidate = safeGreetingValue(candidate, maxGreetingCuisineLength)
		lower := strings.ToLower(candidate)
		if candidate == "" || !strings.HasSuffix(lower, " restaurant") {
			continue
		}
		candidate = strings.TrimSpace(candidate[:len(candidate)-len(" restaurant")])
		if candidate = safeGreetingValue(candidate, maxGreetingCuisineLength); candidate != "" {
			return candidate
		}
	}
	return ""
}

func safeOwnerFirstName(value string) string {
	value = safeGreetingValue(value, maxGreetingOwnerNameLength)
	if value == "" {
		return ""
	}
	return cleanFirstName(value)
}

func safeGreetingValue(value string, maxLength int) string {
	if !utf8.ValidString(value) || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxLength {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "{{") || strings.Contains(lower, "}}") ||
		strings.Contains(lower, "http://") || strings.Contains(lower, "https://") ||
		strings.Contains(lower, "www.") ||
		isPlaceholderGreetingValue(lower) {
		return ""
	}
	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == ' ' {
			continue
		}
		switch char {
		case '-', '\'', '’', '.', '&':
			continue
		default:
			return ""
		}
	}
	return value
}

func isPlaceholderGreetingValue(value string) bool {
	switch strings.Trim(strings.TrimSpace(value), ".-_") {
	case "n/a", "na", "none", "null", "unknown", "undefined", "not available",
		"not provided", "placeholder", "city", "cuisine", "restaurant name":
		return true
	default:
		return false
	}
}
