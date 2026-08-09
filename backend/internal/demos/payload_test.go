package demos

import "testing"

func TestMapPublicPayloadIgnoresLegacyMediaManifest(t *testing.T) {
	payload := MapPublicPayload([]byte(`{
		"restaurant_name":"Test Cafe",
		"media_manifest":[{"url":"https://legacy.example/scraped.jpg"}]
	}`))
	if len(payload.MediaManifest) != 0 || len(payload.Media) != 0 {
		t.Fatalf("payload = %#v, want no legacy media", payload)
	}
}
