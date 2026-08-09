package seoreport

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	colorGood    = "#1f9a4a"
	colorFair    = "#e8a33a"
	colorPoor    = "#e86a2d"
	colorUnknown = "#6b7280"

	seoWeight     = 10
	reviewsWeight = 30
	websiteWeight = 20
	orderWeight   = 10
	menuWeight    = 10
	contactWeight = 10
	listingWeight = 10
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
	"facebook.com":  {},
	"fb.com":        {},
	"instagram.com": {},
	"twitter.com":   {},
	"x.com":         {},
	"tiktok.com":    {},
	"linkedin.com":  {},
	"linktr.ee":     {},
}

var orderOnlineHints = []string{
	"/order", "order-online", "online-order", "ubereats", "doordash",
	"menulog", "deliveroo", "grubhub", "skipthedishes", "toasttab", "square.site",
}

// ScoreInput aggregates Places + enrichment signals for scoring.
type ScoreInput struct {
	Place           PlaceDetails
	Reviews         []Review
	PhotoCount      int
	HasHours        bool
	Delivery        bool
	Takeout         bool
	Reservable      bool
	PlaceKnown      bool
	ReviewsKnown    bool
	PhotoCountKnown bool
	HoursKnown      bool
	WebsiteKnown    bool
	MenuKnown       bool
	PhoneKnown      bool
	EmailKnown      bool
	DeliveryKnown   bool
	TakeoutKnown    bool
	ReservableKnown bool
	Enrichment      Enrichment
	// Website visual audit (from screenshot + AI). When QualityKnown is false,
	// scoreWebsite awards only the verified-site baseline, not inferred quality.
	WebsiteQualityKnown     bool
	WebsiteQualityScore     int // 0–100
	WebsiteReview           string
	WebsiteScreenshot       string
	WebsiteMobileScreenshot string
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
	reviews := scoreReviews(in)
	website := scoreWebsite(place.Website, in.WebsiteQualityKnown, in.WebsiteQualityScore)
	order := scoreOrderOnline(place.Website, in)
	menu := scoreMenu(in)
	contact := scoreContact(place.Phone, place.Email)
	listing := scoreListing(place, hasHours, in)
	seoAssessed := in.PlaceKnown || len(place.Types) > 0 || strings.TrimSpace(place.EditorialSummary) != "" || strings.TrimSpace(place.Address) != ""
	reviewsAssessed := in.PlaceKnown || place.Rating != nil || place.UserRatingCount != nil || len(in.Reviews) > 0
	websiteAssessed := in.WebsiteKnown || strings.TrimSpace(place.Website) != ""
	orderAssessed := in.DeliveryKnown || in.TakeoutKnown || in.ReservableKnown || in.Delivery || in.Takeout || in.Reservable || containsAnyFold(place.Website, orderOnlineHints)
	menuAssessed := in.MenuKnown || in.Enrichment.MenuItemCount > 0 || in.Enrichment.MenuImageCount > 0 || containsAnyFold(place.Website, []string{"/menu", "menu."})
	contactAssessed := in.PhoneKnown || in.EmailKnown || strings.TrimSpace(place.Phone) != "" || strings.TrimSpace(place.Email) != ""
	listingAssessed := in.PlaceKnown || in.HoursKnown || in.PhotoCountKnown || hasHours || in.PhotoCount > 0 || strings.TrimSpace(place.MapsURI) != "" || strings.TrimSpace(place.Address) != "" || strings.TrimSpace(place.BusinessStatus) != ""
	seoComplete := in.PlaceKnown
	reviewsComplete := in.PlaceKnown && in.ReviewsKnown
	websiteComplete := websiteAssessed && (strings.TrimSpace(place.Website) == "" || isSocialWebsite(place.Website) || in.WebsiteQualityKnown)
	orderComplete := in.DeliveryKnown && in.TakeoutKnown && in.ReservableKnown
	menuComplete := in.MenuKnown
	phoneObserved := in.PhoneKnown || strings.TrimSpace(place.Phone) != ""
	emailObserved := in.EmailKnown || strings.TrimSpace(place.Email) != ""
	contactComplete := phoneObserved && emailObserved
	listingComplete := in.PlaceKnown && in.HoursKnown && in.PhotoCountKnown

	metrics := []Metric{
		metricOf("seo", "SEO keywords", seo, seoWeight, seoAssessed, seoComplete),
		metricOf("reviews", "Rating & reviews", reviews, reviewsWeight, reviewsAssessed, reviewsComplete),
		metricOf("website", "Website design", website, websiteWeight, websiteAssessed, websiteComplete),
		metricOf("order_online", "Order online", order, orderWeight, orderAssessed, orderComplete),
		metricOf("menu", "Menu data", menu, menuWeight, menuAssessed, menuComplete),
		metricOf("contact", "Phone & email", contact, contactWeight, contactAssessed, contactComplete),
		metricOf("listing", "Listing completeness", listing, listingWeight, listingAssessed, listingComplete),
	}

	// Every metric already carries its final weight. Summing the evidence-backed
	// buckets keeps the result explainable and avoids the previous hard floor and
	// power curve that collapsed materially different restaurants near 20.
	overall := clamp(seo+reviews+website+order+menu+contact+listing, 0, 100)
	label, color := labelForScore(overall)
	assessedCount := countTrue(seoAssessed, reviewsAssessed, websiteAssessed, orderAssessed, menuAssessed, contactAssessed, listingAssessed)
	completeCount := countTrue(seoComplete, reviewsComplete, websiteComplete, orderComplete, menuComplete, contactComplete, listingComplete)
	if assessedCount == 0 {
		label, color = "Not assessed", colorUnknown
	} else if completeCount < len(metrics) {
		label, color = "Partial", colorUnknown
	}

	issues := buildIssues(place, in, seo, reviews, website, order, menu)
	websiteQualityScore := 0
	if in.WebsiteQualityKnown {
		websiteQualityScore = clamp(in.WebsiteQualityScore, 0, 100)
	}

	return Report{
		RestaurantName: place.Name,
		Address:        place.Address,
		OverallScore:   overall,
		OverallLabel:   label,
		OverallColor:   color,
		Metrics:        metrics,
		Competitors:    make([]CompetitorRow, 0),
		Issues:         issues,
		// No revenue, traffic, conversion, or average-order-value evidence is
		// available in this report. Zero means "not estimated", not "$0 loss".
		EstimatedMonthlyLoss:    0,
		FullReportLocked:        true,
		UnlockCTA:               "Unlock the full SEO report by verifying your email.",
		WebsiteScreenshot:       in.WebsiteScreenshot,
		WebsiteMobileScreenshot: in.WebsiteMobileScreenshot,
		WebsiteQualityScore:     websiteQualityScore,
		WebsiteQualityAssessed:  in.WebsiteQualityKnown,
		WebsiteReview:           in.WebsiteReview,
		RecentReviews:           decorateReviewsForScan(in.Reviews),
	}
}

func decorateReviewsForScan(reviews []Review) []Review {
	if len(reviews) == 0 {
		return nil
	}
	out := make([]Review, 0, len(reviews))
	for i, r := range reviews {
		if i >= 5 {
			break
		}
		text := strings.TrimSpace(r.Text)
		if text == "" && r.Rating <= 0 {
			continue
		}
		if len([]rune(text)) > 160 {
			runes := []rune(text)
			text = string(runes[:157]) + "…"
		}
		r.Text = text
		r.Sentiment = reviewSentiment(r)
		out = append(out, r)
	}
	return out
}

func reviewSentiment(r Review) string {
	text := strings.ToLower(r.Text)
	negHits := 0
	posHits := 0
	for _, w := range []string{"worst", "terrible", "awful", "rude", "dirty", "slow", "overpriced", "never again", "disappoint"} {
		if strings.Contains(text, w) {
			negHits++
		}
	}
	for _, w := range []string{"amazing", "excellent", "love", "delicious", "friendly", "recommend", "great", "perfect", "wonderful"} {
		if strings.Contains(text, w) {
			posHits++
		}
	}
	switch {
	case r.Rating >= 4.5 && negHits == 0:
		return "positive"
	case r.Rating > 0 && r.Rating < 3:
		return "negative"
	case negHits > posHits:
		return "negative"
	case posHits > negHits || r.Rating >= 4:
		return "positive"
	default:
		return "mixed"
	}
}

func scoreSEO(place PlaceDetails) int {
	points := 0.0
	cuisines := cuisineLabels(place.Types)
	if len(cuisines) > 0 {
		// A specific category such as "Italian restaurant" is materially more
		// useful than generic restaurant/establishment tags.
		points += 4
	}

	if strings.TrimSpace(place.EditorialSummary) != "" {
		points += 3
	}

	if strings.TrimSpace(place.Address) != "" {
		points += 2
		if hasLocalityKeyword(place.Name+" "+place.EditorialSummary, place.Address) {
			points++
		}
	}
	return clamp(int(math.Round(points)), 0, seoWeight)
}

// scoreReviews assigns 30 points to reputation: 12 for the aggregate rating,
// 10 for review volume, four for recent-review quality, and four for recency.
// Unavailable provider evidence earns no points and is marked Not assessed in
// the public metric; confirmed weak evidence remains an assessed low score.
func scoreReviews(in ScoreInput) int {
	place := in.Place
	points := 0.0
	if place.Rating != nil {
		// Ratings at or below 2.0 earn no quality points; 5.0 earns all 12.
		points += clampFloat((*place.Rating-2.0)/3.0, 0, 1) * 12
	}

	count := 0
	if place.UserRatingCount != nil {
		count = max(*place.UserRatingCount, 0)
		// Logarithmic volume avoids treating 20 and 500 reviews as equivalent
		// while keeping very large venues from dominating the whole score.
		points += math.Min(1, math.Log1p(float64(count))/math.Log1p(500)) * 10
	}

	recentQuality, recentQualityKnown := recentReviewQuality(in.Reviews)
	if recentQualityKnown {
		points += recentQuality * 4
	}

	freshness, freshnessKnown := recentReviewFreshness(in.Reviews, time.Now().UTC())
	if freshnessKnown {
		points += freshness * 4
	}

	return clamp(int(math.Round(points)), 0, reviewsWeight)
}

// scoreWebsite reserves eight points for a valid dedicated site (six for
// presence and two for HTTPS) and twelve for an observed visual audit.
// Unavailable visual evidence earns no quality points.
func scoreWebsite(website string, qualityKnown bool, qualityScore int) int {
	website = strings.TrimSpace(website)
	if website == "" {
		return 0
	}
	parsed, err := url.Parse(website)
	if err != nil || parsed.Host == "" {
		return 3
	}
	host := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
	if _, social := socialHosts[host]; social {
		return 4
	}

	points := 6.0
	if strings.EqualFold(parsed.Scheme, "https") {
		points += 2
	}
	if qualityKnown {
		points += clampFloat(float64(qualityScore)/100, 0, 1) * 12
	}
	return clamp(int(math.Round(points)), 0, websiteWeight)
}

func scoreOrderOnline(website string, in ScoreInput) int {
	points := booleanEvidencePoints(in.Delivery, 3) +
		booleanEvidencePoints(in.Takeout, 3) +
		booleanEvidencePoints(in.Reservable, 2)
	if containsAnyFold(website, orderOnlineHints) {
		points += 2
	}
	return clamp(int(math.Round(points)), 0, orderWeight)
}

func scoreMenu(in ScoreInput) int {
	points := 0
	menuKnown := in.MenuKnown || in.Enrichment.MenuItemCount > 0 || in.Enrichment.MenuImageCount > 0
	if menuKnown {
		switch {
		case in.Enrichment.MenuItemCount >= 8:
			points += 7
		case in.Enrichment.MenuItemCount >= 3:
			points += 5
		case in.Enrichment.MenuItemCount > 0:
			points += 3
		}
		switch {
		case in.Enrichment.MenuImageCount >= 3:
			points += 2
		case in.Enrichment.MenuImageCount > 0:
			points++
		}
	}
	if containsAnyFold(in.Place.Website, []string{"/menu", "menu."}) {
		points++
	}
	return clamp(points, 0, menuWeight)
}

func scoreContact(phone, email string) int {
	points := 0.0
	if strings.TrimSpace(phone) != "" {
		points += 6
	}
	if strings.TrimSpace(email) != "" {
		points += 4
	}
	return clamp(int(math.Round(points)), 0, contactWeight)
}

func scoreListing(place PlaceDetails, hasHours bool, in ScoreInput) int {
	points := 0.0
	if strings.TrimSpace(place.MapsURI) != "" || strings.TrimSpace(place.Address) != "" {
		points += 2
	}

	status := strings.ToUpper(strings.TrimSpace(place.BusinessStatus))
	if status == "OPERATIONAL" {
		points++
	}

	if hasHours {
		points += 3
	}

	photoKnown := in.PhotoCountKnown || in.PhotoCount > 0
	if photoKnown {
		switch {
		case in.PhotoCount >= 10:
			points += 4
		case in.PhotoCount >= 8:
			points += 3
		case in.PhotoCount >= 3:
			points += 2
		case in.PhotoCount > 0:
			points++
		}
	}
	return clamp(int(math.Round(points)), 0, listingWeight)
}

func buildIssues(place PlaceDetails, in ScoreInput, seo, reviews, website, order, menu int) []Issue {
	issues := make([]Issue, 0, 4)
	if strings.TrimSpace(place.Website) == "" && in.WebsiteKnown {
		issues = append(issues, Issue{
			Title:       "No website on Google",
			Description: "Add a dedicated website so guests have a direct path to menus, reservations, and orders.",
		})
	} else if strings.TrimSpace(place.Website) != "" && in.WebsiteQualityKnown && website < 12 {
		desc := "The observed homepage has gaps in design, clarity, or booking calls to action."
		if strings.TrimSpace(in.WebsiteReview) != "" {
			desc = in.WebsiteReview
		}
		issues = append(issues, Issue{
			Title:       "Website design needs work",
			Description: desc,
		})
	}
	if strings.TrimSpace(place.Phone) == "" && in.PhoneKnown {
		issues = append(issues, Issue{
			Title:       "Phone number missing from listing",
			Description: "Callers can't reach you from Google Maps — add a verified business number.",
		})
	}
	if strings.TrimSpace(place.Email) == "" && in.EmailKnown {
		issues = append(issues, Issue{
			Title:       "Business email not discoverable",
			Description: "Publish a public contact email so guests and partners can reach you without marketplace fees.",
		})
	}
	count := 0
	if place.UserRatingCount != nil {
		count = *place.UserRatingCount
	}
	reviewSummaryKnown := in.PlaceKnown || place.Rating != nil || place.UserRatingCount != nil
	if reviewSummaryKnown && count < 50 {
		issues = append(issues, Issue{
			Title:       "Low review volume",
			Description: fmt.Sprintf("You have %d reviews. More guest feedback provides broader evidence for people evaluating the listing.", count),
		})
	} else if reviewSummaryKnown && reviews < 18 {
		issues = append(issues, Issue{
			Title:       "Review strength needs attention",
			Description: "The aggregate rating, volume, or available dated-review evidence is below the report's strong threshold.",
		})
	}
	orderSignalsKnown := in.DeliveryKnown || in.TakeoutKnown || in.ReservableKnown || containsAnyFold(place.Website, orderOnlineHints)
	if orderSignalsKnown && order < 4 {
		issues = append(issues, Issue{
			Title:       "Order-online path is weak",
			Description: "Enable delivery or takeout signals and link a clear order path so guests can act directly from the listing or website.",
		})
	}
	if in.MenuKnown && menu < 5 {
		issues = append(issues, Issue{
			Title:       "Menu data underrepresented",
			Description: "Structured menu items and photos help guests decide faster and improve local SEO relevance.",
		})
	}
	if in.PlaceKnown && seo < 5 {
		issues = append(issues, Issue{
			Title:       "Local SEO keywords underused",
			Description: fmt.Sprintf("Add city + cuisine terms (e.g. \"%s in your suburb\") to titles and posts.", firstWord(place.Name)),
		})
	}
	if len(issues) > 3 {
		issues = issues[:3]
	}
	return issues
}

func metricOf(key, label string, score, max int, assessed, complete bool) Metric {
	ratio := 0.0
	if max > 0 {
		ratio = float64(score) / float64(max)
	}
	status, color := metricStatus(ratio)
	if !assessed {
		status, color = "Not assessed", colorUnknown
	} else if !complete {
		status, color = "Partially assessed", colorUnknown
	}
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

func countTrue(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
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

func booleanEvidencePoints(value bool, weight float64) float64 {
	if value {
		return weight
	}
	return 0
}

func recentReviewQuality(reviews []Review) (float64, bool) {
	total := 0.0
	count := 0
	for _, review := range reviews {
		if review.Rating <= 0 || review.Rating > 5 {
			continue
		}
		total += review.Rating
		count++
	}
	if count == 0 {
		return 0, false
	}
	average := total / float64(count)
	return clampFloat((average-2)/3, 0, 1), true
}

func recentReviewFreshness(reviews []Review, now time.Time) (float64, bool) {
	total := 0.0
	count := 0
	for _, review := range reviews {
		ageDays, ok := reviewAgeDays(review, now)
		if !ok {
			continue
		}
		switch {
		case ageDays <= 30:
			total += 1
		case ageDays <= 90:
			total += 0.75
		case ageDays <= 180:
			total += 0.5
		case ageDays <= 365:
			total += 0.25
		}
		count++
	}
	if count == 0 {
		return 0, false
	}
	return total / float64(count), true
}

func reviewAgeDays(review Review, now time.Time) (float64, bool) {
	if published, err := time.Parse(time.RFC3339, strings.TrimSpace(review.PublishTime)); err == nil {
		return math.Max(0, now.Sub(published).Hours()/24), true
	}
	if published, err := time.Parse("2006-01-02", strings.TrimSpace(review.PublishTime)); err == nil {
		return math.Max(0, now.Sub(published).Hours()/24), true
	}
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(review.RelativeTime)))
	if len(fields) < 2 {
		return 0, false
	}
	quantity := 1.0
	if parsed, err := strconv.ParseFloat(fields[0], 64); err == nil {
		quantity = parsed
	} else if fields[0] != "a" && fields[0] != "an" {
		return 0, false
	}
	unit := fields[1]
	switch {
	case strings.HasPrefix(unit, "minute"), strings.HasPrefix(unit, "hour"):
		return 0, true
	case strings.HasPrefix(unit, "day"):
		return quantity, true
	case strings.HasPrefix(unit, "week"):
		return quantity * 7, true
	case strings.HasPrefix(unit, "month"):
		return quantity * 30, true
	case strings.HasPrefix(unit, "year"):
		return quantity * 365, true
	default:
		return 0, false
	}
}

func containsAnyFold(value string, hints []string) bool {
	lower := strings.ToLower(value)
	for _, hint := range hints {
		if strings.Contains(lower, strings.ToLower(hint)) {
			return true
		}
	}
	return false
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
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
	return false
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
