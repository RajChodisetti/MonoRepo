package seoreport

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
)

func jpegDataURL(raw []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(raw)
}

func solidJPEG(t *testing.T, width, height int, fill color.Color) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			canvas.Set(x, y, fill)
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, canvas, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return out.Bytes()
}

func TestBuildWebsiteCaptureArtifactsPreservesDistinctViewports(t *testing.T) {
	mobile := solidJPEG(t, 39, 84, color.RGBA{B: 255, A: 255})
	desktop := solidJPEG(t, 128, 80, color.RGBA{R: 255, A: 255})

	artifacts := buildWebsiteCaptureArtifacts(mobile, desktop)
	if artifacts.DesktopDataURL != jpegDataURL(desktop) {
		t.Fatalf("desktop data URL = %q, want desktop capture", artifacts.DesktopDataURL)
	}
	if artifacts.MobileDataURL != jpegDataURL(mobile) {
		t.Fatalf("mobile data URL = %q, want mobile capture", artifacts.MobileDataURL)
	}
	if artifacts.VisionViewport != "desktop_and_mobile" {
		t.Fatalf("vision viewport = %q, want combined provenance", artifacts.VisionViewport)
	}
	if bytes.Equal(artifacts.VisionJPEG, desktop) || bytes.Equal(artifacts.VisionJPEG, mobile) {
		t.Fatal("vision input reused one viewport instead of creating a composite")
	}
	combined, _, err := image.Decode(bytes.NewReader(artifacts.VisionJPEG))
	if err != nil {
		t.Fatalf("decode combined vision input: %v", err)
	}
	bounds := combined.Bounds()
	leftR, _, leftB, _ := combined.At(bounds.Min.X+10, bounds.Min.Y+bounds.Dy()/2).RGBA()
	rightR, _, rightB, _ := combined.At(bounds.Max.X-10, bounds.Min.Y+bounds.Dy()/2).RGBA()
	if leftR <= leftB || rightB <= rightR {
		t.Fatalf("combined provenance lost: left R/B=%d/%d right R/B=%d/%d", leftR, leftB, rightR, rightB)
	}
}

func TestBuildWebsiteCaptureArtifactsDoesNotAliasMobileAsDesktop(t *testing.T) {
	mobile := solidJPEG(t, 39, 84, color.RGBA{B: 255, A: 255})

	artifacts := buildWebsiteCaptureArtifacts(mobile, nil)
	if artifacts.DesktopDataURL != "" {
		t.Fatalf("desktop data URL = %q, want empty when desktop capture failed", artifacts.DesktopDataURL)
	}
	if artifacts.MobileDataURL != jpegDataURL(mobile) {
		t.Fatalf("mobile data URL = %q, want mobile capture", artifacts.MobileDataURL)
	}
	if string(artifacts.VisionJPEG) != string(mobile) {
		t.Fatalf("vision input = %q, want mobile fallback", artifacts.VisionJPEG)
	}
	if artifacts.VisionViewport != "mobile" {
		t.Fatalf("vision viewport=%q, want mobile", artifacts.VisionViewport)
	}
}

func TestBuildWebsiteCaptureArtifactsDoesNotAliasDesktopAsMobile(t *testing.T) {
	desktop := solidJPEG(t, 128, 80, color.RGBA{R: 255, A: 255})

	artifacts := buildWebsiteCaptureArtifacts(nil, desktop)
	if artifacts.DesktopDataURL != jpegDataURL(desktop) {
		t.Fatalf("desktop data URL = %q, want desktop capture", artifacts.DesktopDataURL)
	}
	if artifacts.MobileDataURL != "" {
		t.Fatalf("mobile data URL = %q, want empty when mobile capture failed", artifacts.MobileDataURL)
	}
	if string(artifacts.VisionJPEG) != string(desktop) {
		t.Fatalf("vision input = %q, want desktop capture", artifacts.VisionJPEG)
	}
	if artifacts.VisionViewport != "desktop" {
		t.Fatalf("vision viewport=%q, want desktop", artifacts.VisionViewport)
	}
}

func TestBuildWebsiteCaptureArtifactsHandlesNoCaptures(t *testing.T) {
	artifacts := buildWebsiteCaptureArtifacts(nil, nil)
	if artifacts.DesktopDataURL != "" || artifacts.MobileDataURL != "" || len(artifacts.VisionJPEG) != 0 || artifacts.VisionViewport != "none" {
		t.Fatalf("empty captures produced artifacts: %#v", artifacts)
	}
}

func TestWebsiteVisionPromptStatesViewportCoverageHonestly(t *testing.T) {
	combined := websiteVisionPrompt("desktop_and_mobile")
	if !strings.Contains(combined, "desktop is on the LEFT") || !strings.Contains(combined, "mobile is on the RIGHT") {
		t.Fatalf("combined prompt lacks viewport positions: %q", combined)
	}
	mobile := websiteVisionPrompt("mobile")
	if !strings.Contains(mobile, "Only the mobile viewport") || !strings.Contains(mobile, "do not claim desktop") {
		t.Fatalf("mobile-only prompt overclaims coverage: %q", mobile)
	}
	desktop := websiteVisionPrompt("desktop")
	if !strings.Contains(desktop, "Only the desktop viewport") || !strings.Contains(desktop, "do not claim mobile") {
		t.Fatalf("desktop-only prompt overclaims coverage: %q", desktop)
	}
}

func TestNormalizeListedWebsiteURLPreservesExplicitHTTPScheme(t *testing.T) {
	listedHTTP := "http://restaurant.example/menu"
	if got := normalizeListedWebsiteURL(listedHTTP); got != listedHTTP {
		t.Fatalf("explicit HTTP listing rewritten to %q", got)
	}
	if got := normalizeListedWebsiteURL("restaurant.example"); got != "https://restaurant.example" {
		t.Fatalf("scheme-less URL normalized to %q", got)
	}
}

func TestWebsiteQualityMaximumRemainsReachable(t *testing.T) {
	if got := clampStrictWebsiteQuality(100); got != 100 {
		t.Fatalf("quality clamp=%d, want reachable maximum 100", got)
	}
}

func TestParseWebsiteVisionJSONRequiresCompleteContract(t *testing.T) {
	valid := `{"score":42,"summary":"Readable, but the menu path needs stronger emphasis.","strengths":["Readable"],"weaknesses":["Menu path"]}`
	score, summary, ok := parseWebsiteVisionJSON(valid)
	if !ok || score != 42 || summary != "Readable, but the menu path needs stronger emphasis." {
		t.Fatalf("valid response parsed as score=%d summary=%q ok=%v", score, summary, ok)
	}

	invalid := map[string]string{
		"empty":            "",
		"malformed":        `{"score":42`,
		"markdown fence":   "```json\n" + valid + "\n```",
		"missing score":    `{"summary":"Review","strengths":[],"weaknesses":[]}`,
		"empty summary":    `{"score":42,"summary":" ","strengths":[],"weaknesses":[]}`,
		"score too high":   `{"score":101,"summary":"Review","strengths":[],"weaknesses":[]}`,
		"wrong score type": `{"score":"42","summary":"Review","strengths":[],"weaknesses":[]}`,
		"null arrays":      `{"score":42,"summary":"Review","strengths":null,"weaknesses":[]}`,
		"unknown field":    `{"score":42,"summary":"Review","strengths":[],"weaknesses":[],"confidence":1}`,
	}
	for name, raw := range invalid {
		t.Run(name, func(t *testing.T) {
			score, summary, ok := parseWebsiteVisionJSON(raw)
			if ok || score != 0 || summary != "" {
				t.Fatalf("invalid response parsed as score=%d summary=%q ok=%v", score, summary, ok)
			}
			audit := websiteAuditFromVisionResponse("https://restaurant.example", raw)
			if audit.Source != "fallback" || audit.QualityScore != 0 {
				t.Fatalf("invalid response audit=%#v, want zero-score fallback", audit)
			}
		})
	}
}
