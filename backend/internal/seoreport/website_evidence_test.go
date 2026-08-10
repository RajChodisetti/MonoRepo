package seoreport

import (
	"context"
	"strings"
	"testing"
)

func TestSocialProfileFromURLAcceptsPlatformSpecificProfileShapes(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		platform  string
		handle    string
		canonical string
	}{
		{
			name:      "instagram handle with tracking",
			rawURL:    "https://www.instagram.com/neighborhood.table/?utm_source=restaurant",
			platform:  "Instagram",
			handle:    "neighborhood.table",
			canonical: "https://instagram.com/neighborhood.table",
		},
		{
			name:      "facebook vanity page",
			rawURL:    "facebook.com/NeighborhoodTable",
			platform:  "Facebook",
			handle:    "NeighborhoodTable",
			canonical: "https://facebook.com/NeighborhoodTable",
		},
		{
			name:      "facebook legacy business page",
			rawURL:    "https://www.facebook.com/pages/Neighborhood-Table/123456789",
			platform:  "Facebook",
			handle:    "Neighborhood-Table",
			canonical: "https://facebook.com/pages/Neighborhood-Table/123456789",
		},
		{
			name:      "linkedin company",
			rawURL:    "https://au.linkedin.com/company/neighborhood-table?trk=website",
			platform:  "LinkedIn",
			handle:    "neighborhood-table",
			canonical: "https://linkedin.com/company/neighborhood-table",
		},
		{
			name:      "tiktok handle",
			rawURL:    "https://www.tiktok.com/@neighborhood.table",
			platform:  "TikTok",
			handle:    "neighborhood.table",
			canonical: "https://tiktok.com/@neighborhood.table",
		},
		{
			name:      "youtube handle",
			rawURL:    "https://www.youtube.com/@NeighborhoodTable",
			platform:  "YouTube",
			handle:    "NeighborhoodTable",
			canonical: "https://youtube.com/@NeighborhoodTable",
		},
		{
			name:      "youtube channel",
			rawURL:    "https://youtube.com/channel/UCNeighborhoodTable123",
			platform:  "YouTube",
			handle:    "UCNeighborhoodTable123",
			canonical: "https://youtube.com/channel/UCNeighborhoodTable123",
		},
		{
			name:      "x handle",
			rawURL:    "https://x.com/local_table",
			platform:  "X",
			handle:    "local_table",
			canonical: "https://x.com/local_table",
		},
		{
			name:      "twitter handle",
			rawURL:    "https://mobile.twitter.com/local_table?utm_medium=footer",
			platform:  "X",
			handle:    "local_table",
			canonical: "https://twitter.com/local_table",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, ok := socialProfileFromURL(test.rawURL, "website_link")
			if !ok {
				t.Fatalf("profile URL %q was rejected", test.rawURL)
			}
			if profile.Platform != test.platform || profile.Handle != test.handle ||
				profile.URL != test.canonical || profile.Source != "website_link" {
				t.Fatalf("profile=%#v", profile)
			}
		})
	}
}

func TestSocialProfileFromURLRejectsContentActionsAndRedirects(t *testing.T) {
	tests := map[string]string{
		"instagram auth":               "https://instagram.com/oauth/authorize?client_id=123",
		"instagram login":              "https://instagram.com/accounts/login/",
		"instagram post":               "https://instagram.com/p/ABC123/",
		"instagram reel":               "https://instagram.com/reel/ABC123/",
		"instagram profile reel":       "https://instagram.com/local_table/reels/",
		"instagram query redirect":     "https://l.instagram.com/?u=https%3A%2F%2Fexample.com",
		"facebook login":               "https://facebook.com/login.php?next=%2Flocal_table",
		"facebook auth dialog":         "https://facebook.com/dialog/oauth?client_id=123",
		"facebook pages directory":     "https://facebook.com/pages",
		"facebook sharer":              "https://facebook.com/sharer/sharer.php?u=https%3A%2F%2Fexample.com",
		"facebook reel":                "https://facebook.com/reel/123456",
		"facebook page post":           "https://facebook.com/local_table/posts/123456",
		"linkedin login":               "https://linkedin.com/login?fromSignIn=true",
		"linkedin auth":                "https://linkedin.com/oauth/v2/authorization?client_id=123",
		"linkedin sharing":             "https://linkedin.com/sharing/share-offsite/?url=https%3A%2F%2Fexample.com",
		"linkedin feed post":           "https://linkedin.com/feed/update/urn:li:activity:1",
		"linkedin company post":        "https://linkedin.com/company/local-table/posts/",
		"tiktok login":                 "https://tiktok.com/login?redirect_url=%2F%40local_table",
		"tiktok auth":                  "https://tiktok.com/auth/authorize?client_key=123",
		"tiktok embed":                 "https://tiktok.com/embed/v2/123456",
		"tiktok video":                 "https://tiktok.com/@local_table/video/123456",
		"tiktok query redirect":        "https://tiktok.com/link/v2?target=https%3A%2F%2Fexample.com",
		"youtube watch":                "https://youtube.com/watch?v=ABC123",
		"youtube account login":        "https://youtube.com/account?next=%2F%40local_table",
		"youtube embed":                "https://youtube.com/embed/ABC123",
		"youtube short":                "https://youtube.com/shorts/ABC123",
		"youtube redirect":             "https://youtube.com/redirect?q=https%3A%2F%2Fexample.com",
		"youtube shortened video":      "https://youtu.be/ABC123",
		"x intent":                     "https://x.com/intent/post?url=https%3A%2F%2Fexample.com",
		"x auth":                       "https://x.com/i/oauth2/authorize?client_id=123",
		"x status":                     "https://x.com/local_table/status/123456",
		"twitter share":                "https://twitter.com/share?url=https%3A%2F%2Fexample.com",
		"twitter status":               "https://twitter.com/local_table/status/123456",
		"query-only facebook redirect": "https://facebook.com/?next=%2Flocal_table",
		"linktree aggregator":          "https://linktr.ee/local_table",
	}

	for name, rawURL := range tests {
		t.Run(name, func(t *testing.T) {
			if profile, ok := socialProfileFromURL(rawURL, "website_link"); ok {
				t.Fatalf("non-profile URL %q accepted as %#v", rawURL, profile)
			}
		})
	}
}

func TestLinktreeAloneDoesNotCountAsSocialPresence(t *testing.T) {
	_, extracted := extractWebsiteEvidence("https://restaurant.example", websitePageSignals{
		Links: []capturedWebsiteLink{{Href: "https://linktr.ee/local_table", Text: "All our links"}},
	})
	if extracted.Status != "not_found" || extracted.Score != 0 || len(extracted.Profiles) != 0 {
		t.Fatalf("Linktree-only evidence counted as social presence: %#v", extracted)
	}

	defensive := scoreSocialPresence(SocialPresence{
		Status:   "present",
		Profiles: []SocialProfile{{Platform: "Linktree", Handle: "local_table", URL: "https://linktr.ee/local_table"}},
	})
	if defensive.Status != "not_found" || defensive.Score != 0 || len(defensive.Profiles) != 0 {
		t.Fatalf("Linktree-only profile earned social points: %#v", defensive)
	}
}

func TestLinkAggregatorDoesNotHideVerifiedDownstreamProfile(t *testing.T) {
	presence := scoreSocialPresence(SocialPresence{
		Status: "present",
		Profiles: []SocialProfile{
			{Platform: "Linktree", Handle: "local_table", URL: "https://linktr.ee/local_table"},
			{Platform: "Instagram", Handle: "local_table", URL: "https://instagram.com/local_table"},
		},
	})
	if presence.Status != "present" || presence.Score != 2 || len(presence.Profiles) != 1 ||
		presence.Profiles[0].Platform != "Instagram" {
		t.Fatalf("verified downstream profile was not isolated correctly: %#v", presence)
	}
}

func TestAuditWebsiteClassifiesLinktreeAsListedButNonDedicated(t *testing.T) {
	audit := AuditWebsite(context.Background(), "https://linktr.ee/local_table", nil)
	if !audit.Listed || audit.Reachable || audit.Source != "aggregator" || audit.QualityScore != 0 {
		t.Fatalf("Linktree classification=%#v", audit)
	}
	if audit.ViewportCoverage != "none" || audit.Screenshot != "" || audit.MobileScreenshot != "" {
		t.Fatalf("Linktree received website capture evidence: %#v", audit)
	}
	if audit.SocialPresence.Status != "not_found" || audit.SocialPresence.Score != 0 ||
		len(audit.SocialPresence.Profiles) != 0 {
		t.Fatalf("Linktree received social proof: %#v", audit.SocialPresence)
	}
	if audit.MenuEvidence.Status != "not_found" || !strings.Contains(audit.Review, "not a dedicated restaurant website") {
		t.Fatalf("Linktree review/menu classification=%#v", audit)
	}
}

func TestExtractWebsiteEvidenceUsesOnlyVerifiedMenuAndCanonicalSocialLinks(t *testing.T) {
	menu, social := extractWebsiteEvidence("https://restaurant.example", websitePageSignals{
		Links: []capturedWebsiteLink{
			{Href: "/menu", Text: "Our menu"},
			{Href: "https://www.instagram.com/real_restaurant/?utm_source=site", Text: "Instagram"},
			{Href: "https://instagram.com.evil.example/fake", Text: "Not Instagram"},
		},
		JSONLD: []string{`{"@type":"Restaurant","hasMenu":{"@type":"Menu","url":"/dinner-menu"}}`},
	})
	if menu.Status != "present" || !menu.HasWebsiteLink || !menu.HasStructuredData {
		t.Fatalf("menu evidence=%#v", menu)
	}
	if menu.MenuURL != "https://restaurant.example/menu" || !strings.Contains(menu.Rationale, placesMenuEvidenceLimitation) {
		t.Fatalf("menu URL/rationale=%#v", menu)
	}
	if social.Status != "present" || len(social.Profiles) != 1 || social.Profiles[0].Platform != "Instagram" {
		t.Fatalf("social evidence=%#v", social)
	}
	if social.Profiles[0].URL != "https://instagram.com/real_restaurant" || social.Profiles[0].Handle != "real_restaurant" {
		t.Fatalf("social profile not canonical: %#v", social.Profiles[0])
	}
}

func TestExtractWebsiteEvidenceRecognizesStandaloneMenuJSONLD(t *testing.T) {
	menu, _ := extractWebsiteEvidence("https://restaurant.example", websitePageSignals{
		JSONLD: []string{`{"@context":"https://schema.org","@type":["CreativeWork","Menu"],"url":"/menu.json"}`},
	})
	if menu.Status != "present" || !menu.HasStructuredData || menu.MenuURL != "https://restaurant.example/menu.json" {
		t.Fatalf("standalone Menu JSON-LD not recognized: %#v", menu)
	}
}

func TestExtractWebsiteEvidenceFailsClosedWithoutWebsiteProof(t *testing.T) {
	menu, social := extractWebsiteEvidence("https://restaurant.example", websitePageSignals{
		Links:  []capturedWebsiteLink{{Href: "/gallery", Text: "Food photos"}},
		JSONLD: []string{`{"@type":"Restaurant","image":["dish.jpg","menu-photo.jpg"]}`},
	})
	if menu.Status != "not_found" || menu.HasWebsiteLink || menu.HasStructuredData || menu.MenuURL != "" {
		t.Fatalf("generic images were treated as menu proof: %#v", menu)
	}
	if social.Status != "not_found" || social.Score != 0 || social.Max != 3 {
		t.Fatalf("unexpected social evidence: %#v", social)
	}
}

func TestExtractWebsiteEvidenceCountsDirectOfficialMenuURL(t *testing.T) {
	menu, _ := extractWebsiteEvidence("https://restaurant.example/food-and-drink/menu", websitePageSignals{})
	if menu.Status != "present" || !menu.HasWebsiteLink || menu.MenuURL != "https://restaurant.example/food-and-drink/menu" {
		t.Fatalf("direct official menu URL not recognized: %#v", menu)
	}
}

func TestExtractWebsiteEvidenceRejectsAmbiguousOrNonNavigableMenuLinks(t *testing.T) {
	tests := []struct {
		name string
		link capturedWebsiteLink
	}{
		{name: "bare menu points to about", link: capturedWebsiteLink{Href: "/about", Text: "Menu"}},
		{name: "menu nested under about", link: capturedWebsiteLink{Href: "/about/menu", Text: "View menu"}},
		{name: "same-page fragment", link: capturedWebsiteLink{Href: "#menu", Text: "View our menu"}},
		{name: "navigation toggle path", link: capturedWebsiteLink{Href: "/menu-toggle", Text: "Menu"}},
		{name: "javascript control", link: capturedWebsiteLink{Href: "javascript:void(0)", Text: "Food menu"}},
		{name: "mailto action", link: capturedWebsiteLink{Href: "mailto:hello@restaurant.example", Text: "Food menu"}},
		{name: "telephone action", link: capturedWebsiteLink{Href: "tel:+61390000000", Text: "Food menu"}},
		{name: "host and query substring only", link: capturedWebsiteLink{Href: "https://menu.example.com/about?menu=true", Text: "Learn more"}},
		{name: "homepage query toggle", link: capturedWebsiteLink{Href: "/?menu=true", Text: "View our menu"}},
		{name: "contact mislabeled as menu", link: capturedWebsiteLink{Href: "/contact", Text: "View our food menu"}},
		{name: "menu document nested under contact", link: capturedWebsiteLink{Href: "/contact/menu.pdf", Text: "Menu"}},
		{name: "ambiguous bare menu on dining page", link: capturedWebsiteLink{Href: "/dining", Text: "Menu"}},
		{name: "generic mobile navigation", link: capturedWebsiteLink{Href: "/navigation", Text: "Open menu"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signals := websitePageSignals{
				LoadedURL: "https://restaurant.example/",
				Links:     []capturedWebsiteLink{test.link},
			}
			menu, _ := extractWebsiteEvidence("https://restaurant.example/", signals)
			if menu.Status != "not_found" || menu.HasWebsiteLink || menu.MenuURL != "" || scoreMenu(menu) != 0 {
				t.Fatalf("ambiguous link produced menu evidence/points: %#v score=%d", menu, scoreMenu(menu))
			}
			page := extractWebsitePageEvidence("https://restaurant.example/", signals)
			if page.HasMenuCTA {
				t.Fatalf("ambiguous link produced deterministic menu CTA: %#v", page)
			}
		})
	}
}

func TestExtractWebsiteEvidenceAcceptsGenuineMenuTargets(t *testing.T) {
	tests := []capturedWebsiteLink{
		{Href: "/menu", Text: "Menu"},
		{Href: "/food-menu", Text: "See food menu"},
		{Href: "/dining/menu.pdf", Text: "Download"},
		{Href: "/assets/summer-menu.pdf", Text: "Dinner"},
		{Href: "/downloads/dinner.pdf", Text: "Menu"},
		{Href: "/dining", Text: "View our food menu"},
		{Href: "https://orders.example.com/restaurant/table", Text: "Order online"},
	}
	for _, link := range tests {
		t.Run(link.Href+" "+link.Text, func(t *testing.T) {
			menu, _ := extractWebsiteEvidence("https://restaurant.example/", websitePageSignals{
				LoadedURL: "https://restaurant.example/",
				Links:     []capturedWebsiteLink{link},
			})
			if menu.Status != "present" || !menu.HasWebsiteLink || menu.MenuURL == "" || scoreMenu(menu) != 8 {
				t.Fatalf("genuine target was not accepted: %#v score=%d", menu, scoreMenu(menu))
			}
		})
	}
}

func TestExtractWebsiteEvidenceRejectsFalseOrEmptyMenuJSONLD(t *testing.T) {
	tests := map[string]string{
		"false":                   `{"@type":"Restaurant","hasMenu":false}`,
		"empty string":            `{"@type":"Restaurant","hasMenu":""}`,
		"empty object":            `{"@type":"Restaurant","hasMenu":{}}`,
		"empty array":             `{"@type":"Restaurant","hasMenu":[]}`,
		"null":                    `{"@type":"Restaurant","hasMenu":null}`,
		"empty menu":              `{"@type":"Restaurant","menu":{"name":""}}`,
		"fragment only":           `{"@type":"Restaurant","hasMenu":"#menu"}`,
		"empty menu section type": `{"@type":"Restaurant","hasMenu":{"@type":"MenuSection"}}`,
		"empty nested sections":   `{"@type":"Restaurant","hasMenu":{"hasMenuSection":[{"@type":"MenuSection"}]}}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			menu, _ := extractWebsiteEvidence("https://restaurant.example/", websitePageSignals{JSONLD: []string{raw}})
			if menu.Status != "not_found" || menu.HasStructuredData || menu.MenuURL != "" || scoreMenu(menu) != 0 {
				t.Fatalf("false/empty JSON-LD produced menu evidence/points: %#v score=%d", menu, scoreMenu(menu))
			}
		})
	}
}

func TestExtractWebsiteEvidenceAcceptsSubstantiveMenuJSONLD(t *testing.T) {
	tests := map[string]string{
		"valid URL":         `{"@type":"Restaurant","hasMenu":"/dinner-menu"}`,
		"actual Menu type":  `{"@context":"https://schema.org","@type":"Menu","name":"Dinner"}`,
		"embedded sections": `{"@type":"Restaurant","hasMenu":{"name":"Dinner","hasMenuSection":[{"@type":"MenuSection","hasMenuItem":[{"@type":"MenuItem","name":"Green curry"}]}]}}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			menu, _ := extractWebsiteEvidence("https://restaurant.example/", websitePageSignals{JSONLD: []string{raw}})
			if menu.Status != "present" || !menu.HasStructuredData || scoreMenu(menu) < 8 {
				t.Fatalf("substantive JSON-LD was not accepted: %#v score=%d", menu, scoreMenu(menu))
			}
		})
	}
}

func TestExtractPublicWebsiteContactsAcceptsOnlyValidContactSchemes(t *testing.T) {
	email, phone := extractPublicWebsiteContacts(websitePageSignals{Links: []capturedWebsiteLink{
		{Href: "https://restaurant.example/contact", Text: "hello@ignored.example"},
		{Href: "mailto:Name%20%3Cfake@example.com%3E", Text: "invalid display address"},
		{Href: "mailto:hello%2Bbookings@restaurant.example?subject=Table", Text: "Email"},
		{Href: "tel:call-us", Text: "invalid phone"},
		{Href: "tel:%2B61%203%209000%200000", Text: "Call"},
	}})
	if email != "hello+bookings@restaurant.example" {
		t.Fatalf("public email=%q", email)
	}
	if phone != "+61 3 9000 0000" {
		t.Fatalf("public phone=%q", phone)
	}
}

func TestExtractPublicWebsiteContactsFailsClosed(t *testing.T) {
	email, phone := extractPublicWebsiteContacts(websitePageSignals{Links: []capturedWebsiteLink{
		{Href: "mailto:two@example.com,other@example.com"},
		{Href: "mailto:no-public-domain@localhost"},
		{Href: "tel:+12"},
		{Href: "tel:+61-CALL-NOW"},
	}})
	if email != "" || phone != "" {
		t.Fatalf("invalid contacts escaped validation: email=%q phone=%q", email, phone)
	}
}

func TestExtractWebsitePageEvidenceUsesDeterministicHTMLSignals(t *testing.T) {
	evidence := extractWebsitePageEvidence("http://restaurant.example", websitePageSignals{
		Title:           "Restaurant | Thai food",
		LoadedURL:       "http://restaurant.example/home",
		HasMetaViewport: true,
		Links: []capturedWebsiteLink{
			{Href: "/menu", Text: "View menu"},
			{Href: "https://orders.example/restaurant", Text: "Order online"},
			{Href: "/contact", Text: "Contact us"},
		},
	})
	if evidence.Title == "" || evidence.LoadedScheme != "http" || !evidence.HasMetaViewport ||
		!evidence.HasMenuCTA || !evidence.HasOrderCTA || !evidence.HasContactCTA {
		t.Fatalf("deterministic page evidence=%#v", evidence)
	}
}
