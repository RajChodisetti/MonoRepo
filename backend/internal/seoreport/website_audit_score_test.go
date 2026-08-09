package seoreport

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestCombineWebsiteAuditJPEGIncludesMobileAndDesktop(t *testing.T) {
	encode := func(width, height int, fill color.Color) []byte {
		t.Helper()
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := range height {
			for x := range width {
				img.Set(x, y, fill)
			}
		}
		var out bytes.Buffer
		if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 100}); err != nil {
			t.Fatalf("encode fixture: %v", err)
		}
		return out.Bytes()
	}

	mobile := encode(2, 3, color.RGBA{R: 255, A: 255})
	desktop := encode(4, 2, color.RGBA{B: 255, A: 255})
	combined, err := combineWebsiteAuditJPEG(mobile, desktop)
	if err != nil {
		t.Fatalf("combineWebsiteAuditJPEG: %v", err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(combined))
	if err != nil {
		t.Fatalf("decode combined image: %v", err)
	}
	if got := decoded.Bounds().Size(); got.X != 22 || got.Y != 3 {
		t.Fatalf("combined size = %v, want (22,3)", got)
	}
}

func TestWebsiteCaptureDataURLsPreserveViewportProvenance(t *testing.T) {
	desktopURL, mobileURL := websiteCaptureDataURLs([]byte("mobile"), nil)
	if desktopURL != "" || mobileURL == "" {
		t.Fatalf("mobile-only capture mislabeled: desktop=%q mobile=%q", desktopURL, mobileURL)
	}

	desktopURL, mobileURL = websiteCaptureDataURLs(nil, []byte("desktop"))
	if desktopURL == "" || mobileURL != "" {
		t.Fatalf("desktop-only capture mislabeled: desktop=%q mobile=%q", desktopURL, mobileURL)
	}
}

func TestParseWebsiteVisionJSONRejectsMalformedScore(t *testing.T) {
	if _, _, valid := parseWebsiteVisionJSON("not json"); valid {
		t.Fatal("malformed vision response must not become a fabricated default score")
	}
	if _, _, valid := parseWebsiteVisionJSON(`{"summary":"score omitted"}`); valid {
		t.Fatal("vision response without a score must not become an observed zero")
	}
	score, summary, valid := parseWebsiteVisionJSON(`{"score":93,"summary":"Strong responsive layout."}`)
	if !valid || score != 93 || summary != "Strong responsive layout." {
		t.Fatalf("valid response parsed as score=%d summary=%q valid=%v", score, summary, valid)
	}
}
