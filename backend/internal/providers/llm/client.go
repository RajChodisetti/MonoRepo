package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

// Client is a minimal chat-completion provider.
type Client interface {
	Enabled() bool
	Complete(ctx context.Context, prompt string) (string, error)
	// CompleteVision sends a text prompt plus one image (JPEG/PNG bytes) to a vision-capable model.
	CompleteVision(ctx context.Context, prompt string, image []byte, mimeType string) (string, error)
}

type disabledClient struct{}

func (disabledClient) Enabled() bool { return false }

func (disabledClient) Complete(context.Context, string) (string, error) {
	return "", fmt.Errorf("llm disabled")
}

func (disabledClient) CompleteVision(context.Context, string, []byte, string) (string, error) {
	return "", fmt.Errorf("llm disabled")
}

// NewFromConfig returns an OpenAI-compatible client when LLM_PROVIDER is enabled.
func NewFromConfig(cfg config.LLMConfig) Client {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" || provider == "disabled" {
		return disabledClient{}
	}
	apiKey := strings.TrimSpace(cfg.APIKey)
	model := strings.TrimSpace(cfg.Model)
	if apiKey == "" || model == "" {
		return disabledClient{}
	}
	baseURL := "https://api.openai.com/v1"
	if provider == "openai" || provider == "http" || provider == "https" {
		// default OpenAI-compatible base
	}
	return &openAIClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 55 * time.Second},
	}
}

type openAIClient struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

func (c *openAIClient) Enabled() bool { return c != nil && c.apiKey != "" && c.model != "" }

func (c *openAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("llm disabled")
	}
	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a concise local SEO advisor for restaurants."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.4,
		"max_tokens":  220,
	}
	return c.chat(ctx, body)
}

func (c *openAIClient) CompleteVision(ctx context.Context, prompt string, image []byte, mimeType string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("llm disabled")
	}
	if len(image) == 0 {
		return "", fmt.Errorf("empty image")
	}
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	encoded := base64.StdEncoding.EncodeToString(image)
	dataURL := "data:" + mimeType + ";base64," + encoded
	body := map[string]any{
		"model": c.model,
		"messages": []any{
			map[string]any{
				"role":    "system",
				"content": "You are a harsh restaurant website UX and conversion auditor. Be strict.",
			},
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": prompt},
					map[string]any{
						"type": "image_url",
						"image_url": map[string]any{
							"url": dataURL,
						},
					},
				},
			},
		},
		"temperature": 0.2,
		"max_tokens":  450,
	}
	return c.chat(ctx, body)
}

func (c *openAIClient) chat(ctx context.Context, body map[string]any) (string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := string(raw)
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return "", fmt.Errorf("llm complete failed (%d): %s", resp.StatusCode, msg)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
