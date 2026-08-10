package seoreport

import (
	"encoding/base64"
	"testing"
)

func jpegDataURL(raw []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(raw)
}

func TestBuildWebsiteCaptureArtifactsPreservesDistinctViewports(t *testing.T) {
	mobile := []byte("mobile-jpeg")
	desktop := []byte("desktop-jpeg")

	artifacts := buildWebsiteCaptureArtifacts(mobile, desktop)
	if artifacts.DesktopDataURL != jpegDataURL(desktop) {
		t.Fatalf("desktop data URL = %q, want desktop capture", artifacts.DesktopDataURL)
	}
	if artifacts.MobileDataURL != jpegDataURL(mobile) {
		t.Fatalf("mobile data URL = %q, want mobile capture", artifacts.MobileDataURL)
	}
	if string(artifacts.VisionJPEG) != string(desktop) {
		t.Fatalf("vision input = %q, want desktop capture", artifacts.VisionJPEG)
	}
}

func TestBuildWebsiteCaptureArtifactsDoesNotAliasMobileAsDesktop(t *testing.T) {
	mobile := []byte("mobile-only-jpeg")

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
}

func TestBuildWebsiteCaptureArtifactsDoesNotAliasDesktopAsMobile(t *testing.T) {
	desktop := []byte("desktop-only-jpeg")

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
}

func TestBuildWebsiteCaptureArtifactsHandlesNoCaptures(t *testing.T) {
	artifacts := buildWebsiteCaptureArtifacts(nil, nil)
	if artifacts.DesktopDataURL != "" || artifacts.MobileDataURL != "" || len(artifacts.VisionJPEG) != 0 {
		t.Fatalf("empty captures produced artifacts: %#v", artifacts)
	}
}

func TestWebsiteQualityMaximumRemainsReachable(t *testing.T) {
	if got := clampStrictWebsiteQuality(100); got != 100 {
		t.Fatalf("quality clamp=%d, want reachable maximum 100", got)
	}
}
