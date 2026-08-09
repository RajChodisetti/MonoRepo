package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompleteVisionUsesLowDetailAndBoundedOutput(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/chat/completions" {
			t.Fatalf("path = %q, want /chat/completions", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got := int(payload["max_tokens"].(float64)); got != 220 {
			t.Fatalf("max_tokens = %d, want 220", got)
		}
		messages := payload["messages"].([]any)
		user := messages[1].(map[string]any)
		content := user["content"].([]any)
		image := content[1].(map[string]any)["image_url"].(map[string]any)
		if got := image["detail"]; got != "low" {
			t.Fatalf("image detail = %v, want low", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"score\":42}"}}]}`))
	}))
	defer server.Close()

	client := &openAIClient{
		apiKey:  "test-key",
		model:   "gpt-4.1-nano",
		baseURL: server.URL,
		http:    server.Client(),
	}
	got, err := client.CompleteVision(context.Background(), "review", []byte("jpeg"), "image/jpeg")
	if err != nil {
		t.Fatalf("CompleteVision: %v", err)
	}
	if got != `{"score":42}` {
		t.Fatalf("response = %q", got)
	}
}
