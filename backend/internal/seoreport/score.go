package seoreport

import (
	"fmt"
	"math"
	"net/url"
	"strings"
)

const (
	colorGood   = "#1f9a4a"
	colorFair   = "#e8a33a"
	colorPoor   = "#e86a2d"
	colorOrange = "#d9772a"
)

var genericTypes = map[string]struct{}{
	"point_of_interest": {},
	"establishment":     {},
	"food":              {},
	"store":             {},
	"restaurant":        {},
	"meal_takeaway":     {},
	"meal_delivery":     {},
}

var socialHosts = map[string]struct{}{
	"facebook.com": {},
	"fb.com":       {},
	"instagram.com": {},
	"twitter.com":  {},
	"x.com":        {},
	"tiktok.com":   {},
	"linkedin.com": {},
	"linktr.ee":    {},
}

var orderOnlineHints = []string{
	"/order", "order-online", "online-order", "ubereats", "doordash",
	"menulog", "deliveroo", "grubhub", "skipthedishes", "toasttab", "square.site",
}

var menuKeywords = []string{"menu", "dish", "dishes", "food", "meal", "cuisine", "specials"}

// ScoreInput aggregates Places + enrichment signals for scoring.
type ScoreInput struct {
	Place      PlaceDetails
	Reviews    []Review
	PhotoCount int
	HasHours   bool
	Delivery   bool
	Takeout    bool
	Reservable bool
	DeliveryKnown  bool
	TakeoutKnown   bool
	ReservableKnown bool
	Enrichment Enrichment
}

// BuildReport computes the 100-point SEO report.
func BuildReport(in ScoreInput) Report {
	place := in.Place
	if in.Enrichment.Phone != "" && place.Phone == "" {
		place.Phone = in.Enrichment.Phone
	}
	if in.Enrichment.Email != "" {
		place.Email = in.Enrichment.Email
	}
	if in.Enrichment.Website != "" && place.Website == "" {
		place.Website = in.Enrichment.Website
	}
	hasHours := in.HasHours || in.Enrichment.HasHours

	seo := scoreSEO(place)
	reviews := scoreReviews(place, in.Reviews)
	website := scoreWebsite(place.Website)
	order := scoreOrderOnline(place.Website, in)
	menu := scoreMenu(in)
	contact := scoreContact(place.Phone, place.Email)
	listing := scoreListing(place, hasHours, in.PhotoCount)

	metrics := []Metric{
		metricOf("seo", "SEO keywords", seo, 20),
		metricOf("reviews", "Recent reviews", reviews, 25),
		metricOf("website", "Website", website, 20),
		metricOf("order_online", "Order online", order, 5),
		metricOf("menu", "Menu data", menu, 10),
		metricOf("contact", "Phone & email", contact, 10),
		metricOf("listing", "Listing completeness", listing, 10),
	}

	overall := seo + reviews + website + order + menu + contact + listing
	if overall < 8 {
		overall = 8
	}
	if overall > 98 {
		overall = 98
	}
	label, color := labelForScore(overall)

	issues := buildIssues(place, in, seo, reviews, website, order, menu, contact)
	loss := int(math.Round(float64(100-overall)*12)) + ternaryInt(place.Website == "", 180, 0)

	return Report{
		RestaurantName:       place.Name,
		Address:              place.Address,
		OverallScore:         overall,
		OverallLabel:         label,
		OverallColor:         color,
		Metrics:              metrics,
		Competitors:          buildCompetitors(place, reviews),
		Issues:               issues,
		EstimatedMonthlyLoss: loss,
		FullReportLocked:     true,
		UnlockCTA:            "Unlock the full SEO report by verifying your email.",
	}
}

func scoreSEO(place PlaceDetails) int {
	points := 0
	cuisines := cuisineLabels(place.Types)
	blob := strings.ToLower(strings.Join([]string{
		place.Name,
		place.Address,
		place.EditorialSummary,
		strings.Join(place.Types, " "),
	}, " "))

	if place.EditorialSummary != "" {
		points += 6
	}
	if len(cuisines) > 0 {
		points += 4
		hits := 0
		for _, cuisine := range cuisines {
			if strings.Contains(blob, strings.ToLower(cuisine)) {
				hits++
			}
		}
		if hits > 0 {
			points += 4
		}
		if hits >= 2 {
			points += 2
		}
	}
	// Locality signal: address suburb/city tokens appear with restaurant name context.
	if hasLocalityKeyword(place.Name, place.Address) {
		points += 4
	}
	return clamp(points, 0, 20)
}

func scoreReviews(place PlaceDetails, reviews []Review) int {
	rating := 0.0
	if place.Rating != nil {
		rating = *place.Rating
	}
	count := 0
	if place.UserRatingCount != nil {
		count = *place.UserRatingCount
	}

	volume := math.Min(1, float64(count)/200.0) * 10 // max 10
	stars := (rating / 5.0) * 8                      // max 8

	recent := 0.0
	if len(reviews) > 0 {
		sum := 0.0
		for _, r := range reviews {
			if r.Rating > 0 {
				sum += r.Rating
			} else {
				sum += rating
			}
		}
		avg := sum / float64(len(reviews))
		recent = (avg / 5.0) * 5 // max 5
		// Recency bonus from relative time strings.
		fresh := 0
		for _, r := range reviews {
			rt := strings.ToLower(r.RelativeTime)
			if strings.Contains(rt, "hour") || strings.Contains(rt, "day") || strings.Contains(rt, "week") {
				fresh++
			}
		}
		if fresh >= 2 {
			recent += 2
		} else if fresh >= 1 {
			recent += 1
		}
	} else if count > 0 {
		recent = 1
	}

	return clamp(int(math.Round(volume+stars+recent)), 0, 25)
}

func scoreWebsite(website string) int {
	website = strings.TrimSpace(website)
	if website == "" {
		return 0
	}
	points := 12
	parsed, err := url.Parse(website)
	if err != nil || parsed.Host == "" {
		return 8
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	if parsed.Scheme == "https" {
		points += 4
	}
	if _, social := socialHosts[host]; social {
		points -= 6
	} else {
		points += 4
	}
	return clamp(points, 0, 20)
}

func scoreOrderOnline(website string, in ScoreInput) int {
	points := 0
	if in.DeliveryKnown && in.Delivery {
		points += 2
	}
	if in.TakeoutKnown && in.Takeout {
		points += 2
	}
	if in.ReservableKnown && in.Reservable {
		points += 1
	}
	lower := strings.ToLower(website)
	for _, hint := range orderOnlineHints {
		if strings.Contains(lower, strings.ToLower(hint)) {
			points += 2
			break
		}
	}
	return clamp(points, 0, 5)
}

func scoreMenu(in ScoreInput) int {
	points := 0
	if in.Enrichment.MenuItemCount > 0 {
		points += 6
		if in.Enrichment.MenuItemCount >= 8 {
			points += 2
		}
	}
	if in.Enrichment.MenuImageCount > 0 {
		points += 2
	}
	if points >= 10 {
		return 10
	}
	// Weak Places proxies when inventory menu is missing.
	if in.PhotoCount >= 8 {
		points += 2
	} else if in.PhotoCount >= 3 {
		points += 1
	}
	menuMentions := 0
	for _, rev := range in.Reviews {
		lower := strings.ToLower(rev.Text)
		for _, kw := range menuKeywords {
			if strings.Contains(lower, kw) {
				menuMentions++
				break
			}
		}
	}
	if menuMentions >= 2 {
		points += 2
	} else if menuMentions == 1 {
		points += 1
	}
	return clamp(points, 0, 10)
}

func scoreContact(phone, email string) int {
	points := 0
	if strings.TrimSpace(phone) != "" {
		points += 5
	}
	if strings.TrimSpace(email) != "" {
		points += 5
	}
	return clamp(points, 0, 10)
}

func scoreListing(place PlaceDetails, hasHours bool, photoCount int) int {
	points := 0
	if place.MapsURI != "" || place.Address != "" {
		points += 3
	}
	if hasHours {
		points += 3
	}
	status := strings.ToUpper(place.BusinessStatus)
	if status == "" || status == "OPERATIONAL" {
		points += 2
	}
	if photoCount > 0 {
		points += 2
	}
	return clamp(points, 0, 10)
}

func buildIssues(place PlaceDetails, in ScoreInput, seo, reviews, website, order, menu, contact int) []Issue {
	issues := make([]Issue, 0, 4)
	if website < 10 || place.Website == "" {
		issues = append(issues, Issue{
			Title:       "No strong website on Google",
			Description: "Competitors with a dedicated website capture more direct orders and Google clicks.",
		})
	}
	if contact < 5 || place.Phone == "" {
		issues = append(issues, Issue{
			Title:       "Phone number missing from listing",
			Description: "Callers can't reach you from Google Maps — add a verified business number.",
		})
	}
	if place.Email == "" {
		issues = append(issues, Issue{
			Title:       "Business email not discoverable",
			Description: "Publish a public contact email so guests and partners can reach you without marketplace fees.",
		})
	}
	count := 0
	if place.UserRatingCount != nil {
		count = *place.UserRatingCount
	}
	if reviews < 14 || count < 50 {
		issues = append(issues, Issue{
			Title:       "Low recent review volume vs nearby rivals",
			Description: fmt.Sprintf("You have %d reviews. More recent 5-star reviews lift Map Pack rankings.", count),
		})
	}
	if order < 3 {
		issues = append(issues, Issue{
			Title:       "Order-online path is weak",
			Description: "Enable delivery/takeout signals and link a clear /order path so Google can promote direct ordering.",
		})
	}
	if menu < 5 {
		issues = append(issues, Issue{
			Title:       "Menu data underrepresented",
			Description: "Structured menu items and photos help guests decide faster and improve local SEO relevance.",
		})
	}
	if seo < 12 {
		issues = append(issues, Issue{
			Title:       "Local SEO keywords underused",
			Description: fmt.Sprintf("Add city + cuisine terms (e.g. \"%s in your suburb\") to titles and posts.", firstWord(place.Name)),
		})
	}
	if len(issues) == 0 {
		issues = append(issues, Issue{
			Title:       "Photo freshness gap",
			Description: "Weekly food and interior uploads keep you competitive in Google Maps photo carousels.",
		})
	}
	if len(issues) > 3 {
		issues = issues[:3]
	}
	return issues
}

func buildCompetitors(place PlaceDetails, reviewPoints int) []CompetitorRow {
	rating := "—"
	if place.Rating != nil {
		rating = fmt.Sprintf("%.1f", *place.Rating)
	}
	rank := "10th"
	if reviewPoints >= 20 {
		rank = "4th"
	} else if reviewPoints >= 14 {
		rank = "7th"
	}
	return []CompetitorRow{
		{Rank: "1st", Name: "Nearby competitor A", Rating: "4.8", Score: "23/25", ScoreColor: colorGood, Highlight: false},
		{Rank: "2nd", Name: "Nearby competitor B", Rating: "4.6", Score: "21/25", ScoreColor: colorGood, Highlight: false},
		{Rank: "3rd", Name: "Nearby competitor C", Rating: "4.5", Score: "20/25", ScoreColor: colorGood, Highlight: false},
		{Rank: rank, Name: place.Name, Rating: rating, Score: fmt.Sprintf("%d/25", reviewPoints), ScoreColor: colorOrange, Highlight: true},
	}
}

func metricOf(key, label string, score, max int) Metric {
	ratio := 0.0
	if max > 0 {
		ratio = float64(score) / float64(max)
	}
	status, color := metricStatus(ratio)
	return Metric{
		Key:         key,
		Label:       label,
		Score:       score,
		Max:         max,
		Status:      status,
		StatusColor: color,
		Value:       ratio,
	}
}

func labelForScore(score int) (string, string) {
	if score >= 75 {
		return "Good", colorGood
	}
	if score >= 50 {
		return "Fair", colorFair
	}
	return "Poor", colorPoor
}

func metricStatus(ratio float64) (string, string) {
	if ratio >= 0.75 {
		return "Good", colorGood
	}
	if ratio >= 0.45 {
		return "Fair", colorFair
	}
	return "Poor", colorPoor
}

func cuisineLabels(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		lower := strings.ToLower(t)
		if _, skip := genericTypes[lower]; skip {
			continue
		}
		label := strings.ReplaceAll(t, "_", " ")
		out = append(out, strings.TrimSpace(label))
	}
	return out
}

func hasLocalityKeyword(name, address string) bool {
	name = strings.ToLower(name)
	parts := strings.FieldsFunc(address, func(r rune) bool {
		return r == ',' || r == ' '
	})
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if len(part) < 4 {
			continue
		}
		if strings.Contains(name, part) {
			return true
		}
	}
	// Address present with a multi-token locality is still a weak local signal.
	return len(parts) >= 3
}

func firstWord(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return "restaurant"
	}
	return fields[0]
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ternaryInt(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}
