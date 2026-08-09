package seoreport

import (
	"context"
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
	t.Logf("captured mobile=%d desktop=%d JPEG bytes", len(mobileJPEG), len(desktopJPEG))
}
