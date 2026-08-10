package seoreport

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"
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
	DeliveryKnown   bool
	TakeoutKnown    bool
	ReservableKnown bool
	Enrichment      Enrichment
	// Website visual audit (from screenshot + AI). When QualityKnown is false,
	// scoreWebsite awards only a small reachability score (never visual points).
	WebsiteReachable        bool
	WebsiteQualityKnown     bool
	WebsiteQualityScore     int // 0–100
	WebsiteReview           string
	WebsiteScreenshot       string
	WebsiteMobileScreenshot string
	WebsitePageEvidence     WebsitePageEvidence
	MenuEvidence            MenuEvidence
	SocialPresence          SocialPresence
	CompetitorScan          CompetitorScan
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
	in.Place = place
	hasHours := in.HasHours || in.Enrichment.HasHours

	menuEvidence := normalizeMenuEvidence(in.MenuEvidence, place.Website)
	socialPresence := scoreSocialPresence(in.SocialPresence)
	in.MenuEvidence = menuEvidence
	in.SocialPresence = socialPresence

	websiteTitle := ""
	if in.WebsiteReachable {
		websiteTitle = in.WebsitePageEvidence.Title
	}
	seo := scoreSEO(place, websiteTitle)
	reviews := scoreReviews(place, in.Reviews)
	website := scoreWebsite(place.Website, in.WebsiteReachable, in.WebsiteQualityKnown, in.WebsiteQualityScore)
	order := scoreOrderOnline(place.Website, in)
	menu := scoreMenu(menuEvidence)
	contact := scoreContact(place.Phone, place.Email)
	listing := scoreListing(place, hasHours, in.PhotoCount, socialPresence)
	competitorScan := in.CompetitorScan
	if strings.TrimSpace(competitorScan.ScoreKind) == "" {
		competitorScan = CompetitorScan{
			Status:    "unavailable",
			RadiusKM:  competitorRadiusMeters / 1000,
			ScoreKind: "google_visibility",
			Notice:    "Nearby same-cuisine visibility data was unavailable for this report.",
		}
	}

	metrics := []Metric{
		metricOf("seo", "SEO keywords", seo, 20,
			"Cuisine, locality, and descriptive listing signals determine how clearly Google can understand the venue.",
			seoEvidence(place, websiteTitle),
			"Use the primary cuisine and suburb naturally in the business description, website title, and local posts."),
		metricOf("reviews", "Recent reviews", reviews, 25,
			fmt.Sprintf("The listing has %d reviews and a %s rating; recent review quality and freshness are also weighted.", intValue(place.UserRatingCount), ratingLabel(place.Rating)),
			reviewEvidence(place, in.Reviews),
			"Ask recent guests for specific, honest feedback and reply consistently to new reviews."),
		metricOf("website", "Website design", website, 20,
			websiteRationale(place.Website, in.WebsiteReachable, in.WebsiteQualityKnown, in.WebsiteQualityScore),
			websiteEvidence(place.Website, in.WebsiteReachable, in.WebsiteScreenshot, in.WebsiteMobileScreenshot, in.WebsitePageEvidence),
			"Keep the homepage fast, mobile-readable, and explicit about cuisine, menu, ordering, and reservations."),
		metricOf("order_online", "Order online", order, 5,
			"Delivery, takeout, reservations, and a clear direct-order path are scored from known listing and website signals.",
			orderEvidence(in),
			"Add a prominent direct order or reservation link and keep Google service attributes accurate."),
		metricOf("menu", "Menu data", menu, 10,
			menuEvidence.Rationale,
			menuMetricEvidence(menuEvidence),
			"Publish a crawlable menu page and add Restaurant/Menu structured data to the official website."),
		metricOf("contact", "Phone & email", contact, 10,
			"Guests need both a public phone number and a discoverable business email.",
			contactEvidence(place.Phone, place.Email),
			"List a monitored business phone and email consistently on Google and the official website."),
		metricOf("listing", "Listing completeness", listing, 10,
			"Address, hours, operating status, cuisine category, photos, and verified social profiles make up this bucket.",
			listingEvidence(place, hasHours, in.PhotoCount, socialPresence),
			"Complete every Google listing field and link maintained social profiles from the official website."),
	}

	rawTotal := seo + reviews + website + order + menu + contact + listing
	overall := strictOverallScore(rawTotal)
	label, color := labelForScore(overall)

	issues := buildIssues(place, in, seo, reviews, website, order, menu, contact)

	return Report{
		RestaurantName: place.Name,
		Address:        place.Address,
		OverallScore:   overall,
		OverallLabel:   label,
		OverallColor:   color,
		Metrics:        metrics,
		Competitors:    competitorScan.Rows,
		CompetitorScan: competitorScan,
		MenuEvidence:   menuEvidence,
		SocialPresence: socialPresence,
		Issues:         issues,
		// Revenue impact needs venue sales and conversion data. Keep this legacy
		// compatibility field neutral instead of publishing a fabricated estimate.
		EstimatedMonthlyLoss:    0,
		FullReportLocked:        true,
		UnlockCTA:               "Unlock the full SEO report by verifying your email.",
		WebsiteScreenshot:       in.WebsiteScreenshot,
		WebsiteMobileScreenshot: in.WebsiteMobileScreenshot,
		WebsiteQualityScore:     in.WebsiteQualityScore,
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

func scoreSEO(place PlaceDetails, websiteTitle string) int {
	points := 0
	cuisines := cuisineLabels(append([]string{place.PrimaryType}, place.Types...))
	blob := strings.ToLower(strings.Join([]string{
		place.Name,
		place.Address,
		place.EditorialSummary,
	}, " "))

	if place.EditorialSummary != "" {
		points += 4
	}
	if len(cuisines) > 0 {
		points += 6
		hits := 0
		for _, cuisine := range cuisines {
			if strings.Contains(blob, strings.ToLower(cuisine)) {
				hits++
			}
		}
		if hits > 0 {
			points += 4
		}
	}
	// A formatted address proves where the venue is, but not that its public
	// title/content targets that locality. Award locality points only when a
	// locality component from the address is corroborated by a separate public
	// signal: the listing name, editorial summary, or captured website title.
	if hasExplicitLocalityEvidence(place.Address, place.Name, place.EditorialSummary, websiteTitle) {
		points += 6
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

	// Stricter volume curve — 200 reviews no longer maxes the bucket.
	volume := math.Min(1, float64(count)/500.0) * 10 // max 10
	stars := (rating / 5.0) * 10                     // max 10

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
		recent = (avg / 5.0) * 4 // max 4
		// Recency bonus from relative time strings.
		fresh := 0
		for _, r := range reviews {
			rt := strings.ToLower(r.RelativeTime)
			if strings.Contains(rt, "hour") || strings.Contains(rt, "day") || strings.Contains(rt, "week") {
				fresh++
			}
		}
		if fresh >= 2 {
			recent += 1
		}
	} else if count > 0 {
		recent = 0.5
	}

	return clamp(int(math.Round(volume+stars+recent)), 0, 25)
}

// scoreWebsite maps homepage visual quality into the 20-point Website bucket.
// Presence alone must NEVER award a full 20 — scoring is driven by screenshot + AI review.
func scoreWebsite(website string, reachable, qualityKnown bool, qualityScore int) int {
	website = strings.TrimSpace(website)
	if website == "" || isLinkAggregatorWebsite(website) || !reachable {
		return 0
	}
	parsed, err := url.Parse(website)
	if err != nil || parsed.Host == "" {
		return 3
	}
	if isSocialWebsite(website) {
		// Social links are not a real restaurant website experience.
		return 2
	}

	if qualityKnown {
		points := int(math.Round(float64(qualityScore) / 100.0 * 20.0))
		return clamp(points, 1, 20)
	}

	// The live homepage was reached, but screenshot quality was not scored.
	return 4
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
	if in.WebsiteReachable && !isLinkAggregatorWebsite(website) {
		lower := strings.ToLower(website)
		if in.WebsitePageEvidence.HasOrderCTA {
			points += 2
		} else {
			for _, hint := range orderOnlineHints {
				if strings.Contains(lower, strings.ToLower(hint)) {
					points += 2
					break
				}
			}
		}
	}
	return clamp(points, 0, 5)
}

func scoreMenu(evidence MenuEvidence) int {
	points := 0
	if evidence.Status != "present" {
		return 0
	}
	points += 6
	if evidence.HasWebsiteLink {
		points += 2
	}
	if evidence.HasStructuredData {
		points += 2
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

func scoreListing(place PlaceDetails, hasHours bool, photoCount int, social SocialPresence) int {
	points := 0
	if place.MapsURI != "" || place.Address != "" {
		points++
	}
	if hasHours {
		points += 2
	}
	status := strings.ToUpper(strings.TrimSpace(place.BusinessStatus))
	if status == "OPERATIONAL" {
		points += 1
	}
	if photoCount >= 8 {
		points += 2
	} else if photoCount > 0 {
		points += 1
	}
	if strings.TrimSpace(place.PrimaryType) != "" || len(cuisineLabels(place.Types)) > 0 {
		points++
	}
	points += clamp(social.Score, 0, 3)
	return clamp(points, 0, 10)
}

func buildIssues(place PlaceDetails, in ScoreInput, seo, reviews, website, order, menu, contact int) []Issue {
	issues := make([]Issue, 0, 4)
	if place.Website == "" {
		issues = append(issues, Issue{
			Title:       "No website on Google",
			Description: "Competitors with a dedicated website capture more direct orders and Google clicks.",
		})
	} else if isLinkAggregatorWebsite(place.Website) {
		issues = append(issues, Issue{
			Title:       "No dedicated restaurant website",
			Description: "The listing points to a Linktree aggregator. Add a dedicated, mobile-friendly site for menus, direct orders, reservations, and search visibility.",
		})
	} else if !in.WebsiteReachable {
		issues = append(issues, Issue{
			Title:       "Website listed but not reachable",
			Description: "The listed homepage could not be verified, so no reachability or visual-quality points were awarded.",
		})
	} else if website < 10 {
		desc := "Your linked site under-delivers on design, clarity, or booking CTAs versus stronger local rivals."
		if strings.TrimSpace(in.WebsiteReview) != "" {
			desc = in.WebsiteReview
		}
		issues = append(issues, Issue{
			Title:       "Website design needs work",
			Description: desc,
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
	if in.SocialPresence.Status != "present" {
		description := "Link maintained restaurant social profiles from the official website so guests can verify activity and updates."
		if in.SocialPresence.Status == "unknown" {
			description = "The website scan could not verify social links. Add clear profile links on the official homepage."
		}
		issues = append(issues, Issue{Title: "Social presence is incomplete", Description: description})
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
			Description: "Publish a crawlable website menu and Menu structured data; generic Google listing photos do not prove menu presence.",
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

func buildCompetitorScan(target placeSnapshot, candidates []placeSnapshot, cuisineType string) CompetitorScan {
	cuisineType = strings.ToLower(strings.TrimSpace(cuisineType))
	if cuisineType == "" {
		cuisineType = competitorCuisineType(target.Details)
	}
	scan := CompetitorScan{
		RadiusKM:     competitorRadiusMeters / 1000,
		Cuisine:      strings.TrimSpace(strings.ReplaceAll(cuisineType, "_", " ")),
		ScoreKind:    "google_visibility",
		CurrentScore: googleVisibilityScore(target),
	}
	type scoredCandidate struct {
		snapshot placeSnapshot
		score    int
	}
	stronger := make([]scoredCandidate, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	currentID := sanitizePlaceID(target.Details.PlaceID)
	for _, candidate := range candidates {
		candidateID := sanitizePlaceID(candidate.Details.PlaceID)
		if candidateID == "" || candidateID == currentID || isClosedBusiness(candidate.Details.BusinessStatus) {
			continue
		}
		if _, exists := seen[candidateID]; exists {
			continue
		}
		seen[candidateID] = struct{}{}
		if cuisineType != "restaurant" && !placeHasType(candidate.Details, cuisineType) {
			continue
		}
		if strings.TrimSpace(candidate.Details.Name) == "" {
			continue
		}
		if target.Details.Latitude != nil && target.Details.Longitude != nil {
			if candidate.Details.Latitude == nil || candidate.Details.Longitude == nil {
				continue
			}
			candidate.DistanceKM = haversineKM(
				*target.Details.Latitude,
				*target.Details.Longitude,
				*candidate.Details.Latitude,
				*candidate.Details.Longitude,
			)
		}
		if candidate.DistanceKM > competitorRadiusMeters/1000+0.001 {
			continue
		}
		scan.SampleSize++
		score := googleVisibilityScore(candidate)
		if score > scan.CurrentScore {
			stronger = append(stronger, scoredCandidate{snapshot: candidate, score: score})
		}
	}

	if scan.SampleSize == 0 {
		scan.Status = "no_data"
		scan.Notice = "No eligible same-cuisine Google Places listings were returned inside 10 km."
		return scan
	}

	sort.SliceStable(stronger, func(i, j int) bool {
		if stronger[i].score != stronger[j].score {
			return stronger[i].score > stronger[j].score
		}
		leftRating := floatValue(stronger[i].snapshot.Details.Rating)
		rightRating := floatValue(stronger[j].snapshot.Details.Rating)
		if leftRating != rightRating {
			return leftRating > rightRating
		}
		leftCount := intValue(stronger[i].snapshot.Details.UserRatingCount)
		rightCount := intValue(stronger[j].snapshot.Details.UserRatingCount)
		if leftCount != rightCount {
			return leftCount > rightCount
		}
		if stronger[i].snapshot.DistanceKM != stronger[j].snapshot.DistanceKM {
			return stronger[i].snapshot.DistanceKM < stronger[j].snapshot.DistanceKM
		}
		return stronger[i].snapshot.Details.Name < stronger[j].snapshot.Details.Name
	})
	for index, candidate := range stronger {
		details := candidate.snapshot.Details
		scan.Rows = append(scan.Rows, CompetitorRow{
			Rank:            ordinalPosition(index + 1),
			PlaceID:         details.PlaceID,
			Name:            details.Name,
			Rating:          ratingLabel(details.Rating),
			UserRatingCount: intValue(details.UserRatingCount),
			Score:           fmt.Sprintf("%d/100", candidate.score),
			VisibilityScore: candidate.score,
			ScoreMax:        100,
			DistanceKM:      math.Round(candidate.snapshot.DistanceKM*10) / 10,
			Reasons:         visibilityReasons(target, candidate.snapshot),
			Attributions:    append([]PlaceAttribution(nil), details.Attributions...),
			ScoreColor:      colorGood,
		})
	}
	scan.Status = "complete"
	scan.CurrentPosition = len(scan.Rows) + 1
	scan.CurrentRestaurantLeading = len(scan.Rows) == 0
	if scan.CurrentRestaurantLeading {
		scan.Notice = "No sampled same-cuisine restaurant had a higher deterministic Google visibility score."
	} else {
		scan.Notice = "Only sampled restaurants with a stronger deterministic Google visibility score are shown; this is not a live Google Search rank."
	}
	return scan
}

func googleVisibilityScore(snapshot placeSnapshot) int {
	rating := floatValue(snapshot.Details.Rating)
	count := intValue(snapshot.Details.UserRatingCount)
	points := int(math.Round(math.Max(0, math.Min(5, rating)) / 5 * 25))
	points += int(math.Round(math.Min(1, float64(maxInt(count, 0))/500.0) * 25))
	if strings.TrimSpace(snapshot.Details.Website) != "" {
		points += 10
	}
	if strings.TrimSpace(snapshot.Details.Phone) != "" {
		points += 10
	}
	if snapshot.HasHours {
		points += 10
	}
	switch {
	case snapshot.PhotoCount >= 8:
		points += 10
	case snapshot.PhotoCount >= 3:
		points += 7
	case snapshot.PhotoCount > 0:
		points += 4
	}
	if (snapshot.DeliveryKnown && snapshot.Delivery) ||
		(snapshot.TakeoutKnown && snapshot.Takeout) ||
		(snapshot.ReservableKnown && snapshot.Reservable) {
		points += 5
	}
	if strings.EqualFold(strings.TrimSpace(snapshot.Details.BusinessStatus), "OPERATIONAL") {
		points += 5
	}
	return clamp(points, 0, 100)
}

func ternaryVisibilityScore(snapshot *placeSnapshot) int {
	if snapshot == nil {
		return 0
	}
	return googleVisibilityScore(*snapshot)
}

func visibilityReasons(target, candidate placeSnapshot) []string {
	reasons := make([]string, 0, 4)
	if floatValue(candidate.Details.Rating) > floatValue(target.Details.Rating)+0.049 {
		reasons = append(reasons, "Higher Google rating")
	}
	if intValue(candidate.Details.UserRatingCount) > intValue(target.Details.UserRatingCount) {
		reasons = append(reasons, "More Google reviews")
	}
	if target.Details.Website == "" && candidate.Details.Website != "" {
		reasons = append(reasons, "Website linked")
	}
	if !target.HasHours && candidate.HasHours {
		reasons = append(reasons, "Hours complete")
	}
	if candidate.PhotoCount > target.PhotoCount {
		reasons = append(reasons, "More listing photos")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Stronger combined listing completeness")
	}
	if len(reasons) > 3 {
		reasons = reasons[:3]
	}
	return reasons
}

func ordinalPosition(position int) string {
	if position%100 >= 11 && position%100 <= 13 {
		return fmt.Sprintf("%dth", position)
	}
	suffix := "th"
	switch position % 10 {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}
	return fmt.Sprintf("%d%s", position, suffix)
}

func metricOf(key, label string, score, max int, rationale string, evidence []string, recommendation string) Metric {
	ratio := 0.0
	if max > 0 {
		ratio = float64(score) / float64(max)
	}
	status, color := metricStatus(ratio)
	return Metric{
		Key:            key,
		Label:          label,
		Score:          score,
		Max:            max,
		Status:         status,
		StatusColor:    color,
		Value:          ratio,
		Rationale:      strings.TrimSpace(rationale),
		Evidence:       evidence,
		Recommendation: strings.TrimSpace(recommendation),
	}
}

func labelForScore(score int) (string, string) {
	if score >= 70 {
		return "Good", colorGood
	}
	if score >= 45 {
		return "Fair", colorFair
	}
	return "Poor", colorPoor
}

func metricStatus(ratio float64) (string, string) {
	if ratio >= 0.65 {
		return "Good", colorGood
	}
	if ratio >= 0.35 {
		return "Fair", colorFair
	}
	return "Poor", colorPoor
}

// strictOverallScore now preserves the explicit 100-point weighting contract.
// The historical second compression made the declared metric maxima impossible.
func strictOverallScore(raw int) int {
	return clamp(raw, 0, 100)
}

func normalizeMenuEvidence(evidence MenuEvidence, website string) MenuEvidence {
	switch evidence.Status {
	case "present", "not_found", "unknown":
		if strings.TrimSpace(evidence.Rationale) == "" {
			evidence.Rationale = placesMenuEvidenceLimitation
		}
		return evidence
	}
	if strings.TrimSpace(website) == "" {
		return noWebsiteMenuEvidence()
	}
	return unknownWebsiteMenuEvidence("The official website was listed but menu evidence was not inspected.")
}

func scoreSocialPresence(presence SocialPresence) SocialPresence {
	presence.Max = 3
	if presence.Status != "present" || len(presence.Profiles) == 0 {
		presence.Score = 0
		if presence.Status == "" {
			presence.Status = "unknown"
		}
		return presence
	}
	verified := socialPresenceFromProfiles(presence.Profiles)
	presence.Profiles = verified.Profiles
	if len(presence.Profiles) == 0 {
		presence.Status = "not_found"
		presence.Score = 0
		presence.Rationale = verified.Rationale
		return presence
	}
	platforms := make(map[string]struct{}, len(presence.Profiles))
	for _, profile := range presence.Profiles {
		platform := strings.ToLower(strings.TrimSpace(profile.Platform))
		if platform != "" {
			platforms[platform] = struct{}{}
		}
	}
	if len(platforms) >= 2 {
		presence.Score = 3
	} else {
		presence.Score = 2
	}
	return presence
}

func seoEvidence(place PlaceDetails, websiteTitle string) []string {
	evidence := []string{
		fmt.Sprintf("Primary category: %s", firstNonEmpty(place.PrimaryType, "not supplied")),
		fmt.Sprintf("Cuisine categories found: %d", len(cuisineLabels(append([]string{place.PrimaryType}, place.Types...)))),
	}
	if strings.TrimSpace(place.EditorialSummary) != "" {
		evidence = append(evidence, "Google editorial summary available")
	} else {
		evidence = append(evidence, "No Google editorial summary")
	}
	if hasExplicitLocalityEvidence(place.Address, place.Name, place.EditorialSummary, websiteTitle) {
		evidence = append(evidence, "Locality corroborated in the listing name, editorial summary, or reachable website title")
	} else if strings.TrimSpace(place.Address) != "" {
		evidence = append(evidence, "Address available; no separate locality-targeting signal verified")
	}
	return evidence
}

func reviewEvidence(place PlaceDetails, reviews []Review) []string {
	evidence := []string{
		fmt.Sprintf("Rating: %s", ratingLabel(place.Rating)),
		fmt.Sprintf("Review count: %d", intValue(place.UserRatingCount)),
		fmt.Sprintf("Recent review sample: %d", len(reviews)),
	}
	fresh := 0
	for _, review := range reviews {
		when := strings.ToLower(review.RelativeTime)
		if strings.Contains(when, "hour") || strings.Contains(when, "day") || strings.Contains(when, "week") {
			fresh++
		}
	}
	evidence = append(evidence, fmt.Sprintf("Fresh reviews in sample: %d", fresh))
	return evidence
}

func websiteRationale(website string, reachable, qualityKnown bool, qualityScore int) string {
	if strings.TrimSpace(website) == "" {
		return "No official website was listed."
	}
	if isLinkAggregatorWebsite(website) {
		return "The listed destination is a Linktree aggregator, not a dedicated restaurant website; no reachability or visual-quality points were awarded."
	}
	if !reachable {
		return "An official website URL is listed, but the live homepage could not be verified as reachable; no reachability or visual-quality points were awarded."
	}
	if isSocialWebsite(website) {
		return "The listing points to a social profile rather than a dedicated restaurant website."
	}
	if qualityKnown {
		return fmt.Sprintf("The captured homepage received a %d/100 visual and usability assessment.", clamp(qualityScore, 0, 100))
	}
	return "A dedicated website is listed, but a complete visual assessment was unavailable."
}

func websiteEvidence(website string, reachable bool, desktopScreenshot, mobileScreenshot string, page WebsitePageEvidence) []string {
	if isLinkAggregatorWebsite(website) {
		return []string{
			"Listed destination: Linktree aggregator",
			"Dedicated restaurant website: not found",
			"Website visual-quality assessment: not applicable",
		}
	}
	evidence := []string{fmt.Sprintf("Official website: %s", presenceLabel(website != ""))}
	evidence = append(evidence,
		fmt.Sprintf("Live homepage reachable: %s", presenceLabel(reachable)),
		fmt.Sprintf("Loaded homepage scheme: %s", firstNonEmpty(strings.ToUpper(page.LoadedScheme), "unknown")),
		fmt.Sprintf("HTML page title: %s", presenceLabel(page.Title != "")),
		fmt.Sprintf("Mobile viewport metadata: %s", presenceLabel(page.HasMetaViewport)),
		fmt.Sprintf("Order/reservation CTA: %s", presenceLabel(page.HasOrderCTA)),
		fmt.Sprintf("Menu CTA: %s", presenceLabel(page.HasMenuCTA)),
		fmt.Sprintf("Contact CTA: %s", presenceLabel(page.HasContactCTA)),
		fmt.Sprintf("Desktop capture: %s", presenceLabel(desktopScreenshot != "")),
		fmt.Sprintf("Mobile capture: %s", presenceLabel(mobileScreenshot != "")),
	)
	return evidence
}

func orderEvidence(in ScoreInput) []string {
	websiteEligible := in.WebsiteReachable && !isLinkAggregatorWebsite(in.Place.Website)
	return []string{
		fmt.Sprintf("Delivery: %s", knownBoolLabel(in.DeliveryKnown, in.Delivery)),
		fmt.Sprintf("Takeout: %s", knownBoolLabel(in.TakeoutKnown, in.Takeout)),
		fmt.Sprintf("Reservations: %s", knownBoolLabel(in.ReservableKnown, in.Reservable)),
		fmt.Sprintf("Direct-order CTA or URL hint on reachable website: %s", presenceLabel(websiteEligible && (in.WebsitePageEvidence.HasOrderCTA || hasOrderOnlineHint(in.Place.Website)))),
	}
}

func menuMetricEvidence(evidence MenuEvidence) []string {
	return []string{
		fmt.Sprintf("Website menu link: %s", presenceLabel(evidence.HasWebsiteLink)),
		fmt.Sprintf("Menu structured data: %s", presenceLabel(evidence.HasStructuredData)),
		"Google Places generic photos are excluded from menu scoring",
	}
}

func contactEvidence(phone, email string) []string {
	return []string{
		fmt.Sprintf("Public phone: %s", presenceLabel(strings.TrimSpace(phone) != "")),
		fmt.Sprintf("Business email: %s", presenceLabel(strings.TrimSpace(email) != "")),
	}
}

func listingEvidence(place PlaceDetails, hasHours bool, photoCount int, social SocialPresence) []string {
	return []string{
		fmt.Sprintf("Address/Maps link: %s", presenceLabel(place.Address != "" || place.MapsURI != "")),
		fmt.Sprintf("Opening hours: %s", presenceLabel(hasHours)),
		fmt.Sprintf("Business status: %s", firstNonEmpty(place.BusinessStatus, "unknown")),
		fmt.Sprintf("Listing photos: %d", photoCount),
		fmt.Sprintf("Cuisine category: %s", firstNonEmpty(place.PrimaryType, "not supplied")),
		fmt.Sprintf("Verified social platforms: %d", distinctSocialPlatforms(social.Profiles)),
	}
}

func distinctSocialPlatforms(profiles []SocialProfile) int {
	platforms := make(map[string]struct{}, len(profiles))
	for _, profile := range profiles {
		if platform := strings.ToLower(strings.TrimSpace(profile.Platform)); platform != "" {
			platforms[platform] = struct{}{}
		}
	}
	return len(platforms)
}

func hasOrderOnlineHint(website string) bool {
	lower := strings.ToLower(website)
	for _, hint := range orderOnlineHints {
		if strings.Contains(lower, strings.ToLower(hint)) {
			return true
		}
	}
	return false
}

func presenceLabel(present bool) string {
	if present {
		return "present"
	}
	return "not found"
}

func knownBoolLabel(known, value bool) string {
	if !known {
		return "unknown"
	}
	if value {
		return "available"
	}
	return "not available"
}

func ratingLabel(rating *float64) string {
	if rating == nil {
		return "—"
	}
	return fmt.Sprintf("%.1f", *rating)
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func floatValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func cuisineLabels(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		lower := strings.ToLower(t)
		if _, skip := genericTypes[lower]; skip {
			continue
		}
		label := strings.ReplaceAll(t, "_", " ")
		out = append(out, strings.TrimSpace(label))
	}
	return out
}

func hasExplicitLocalityEvidence(address string, corroboratingText ...string) bool {
	terms := formattedAddressLocalityTerms(address)
	if len(terms) == 0 {
		return false
	}
	words := make(map[string]struct{})
	for _, text := range corroboratingText {
		for _, word := range normalizedWords(text) {
			words[word] = struct{}{}
		}
	}
	for _, term := range terms {
		if _, ok := words[term]; ok {
			return true
		}
	}
	return false
}

// formattedAddressLocalityTerms deliberately ignores the first comma-separated
// component because it is normally a street address. A single undifferentiated
// address cannot safely identify its locality component and earns no SEO bonus.
func formattedAddressLocalityTerms(address string) []string {
	components := strings.Split(address, ",")
	if len(components) < 2 {
		return nil
	}
	ignored := map[string]struct{}{
		"australia": {}, "canada": {}, "england": {}, "ireland": {},
		"kingdom": {}, "states": {}, "united": {}, "zealand": {},
	}
	seen := make(map[string]struct{})
	terms := make([]string, 0, 4)
	for _, component := range components[1:] {
		for _, word := range normalizedWords(component) {
			if len([]rune(word)) < 4 {
				continue
			}
			if _, skip := ignored[word]; skip {
				continue
			}
			if _, exists := seen[word]; exists {
				continue
			}
			seen[word] = struct{}{}
			terms = append(terms, word)
		}
	}
	return terms
}

func normalizedWords(value string) []string {
	return strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r)
	})
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
