package seoreport

import (
	"strings"
	"testing"
)

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
