package seoreport

import (
	"bytes"
	"context"
	"image"
	"os"
	"testing"
	"time"
)

// TestBrowserCaptureSmoke is opt-in because it launches the packaged Chromium
// and reaches a public URL. Deployment preflight runs the compiled test binary
// inside the exact API image and capability boundary used in production.
func TestBrowserCaptureSmoke(t *testing.T) {
	target := os.Getenv("TUVI_BROWSER_SMOKE_URL")
	if target == "" {
		t.Skip("set TUVI_BROWSER_SMOKE_URL to run the packaged Chromium smoke")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	mobileJPEG, desktopJPEG, err := captureWebsiteJPEGPair(ctx, target)
	if err != nil {
		t.Fatalf("capture website JPEG pair: %v", err)
	}
	if len(mobileJPEG) < 100 || len(desktopJPEG) < 100 {
		t.Fatalf("capture website JPEG pair returned mobile=%d desktop=%d bytes", len(mobileJPEG), len(desktopJPEG))
	}
	mobileConfig, _, err := image.DecodeConfig(bytes.NewReader(mobileJPEG))
	if err != nil {
		t.Fatalf("decode mobile JPEG config: %v", err)
	}
	desktopConfig, _, err := image.DecodeConfig(bytes.NewReader(desktopJPEG))
	if err != nil {
		t.Fatalf("decode desktop JPEG config: %v", err)
	}
	if mobileConfig.Width != 780 || mobileConfig.Height != 1688 {
		t.Fatalf("mobile dimensions = %dx%d, want 780x1688", mobileConfig.Width, mobileConfig.Height)
	}
	if desktopConfig.Width != 1280 || desktopConfig.Height != 800 {
		t.Fatalf("desktop dimensions = %dx%d, want 1280x800", desktopConfig.Width, desktopConfig.Height)
	}
	t.Logf("captured mobile=%d desktop=%d JPEG bytes", len(mobileJPEG), len(desktopJPEG))
}
