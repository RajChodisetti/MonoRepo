package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

const (
	ProviderResend      = "resend"
	defaultResendAPIURL = "https://api.resend.com"
)

type resendProvider struct {
	cfg     config.EmailConfig
	client  *http.Client
	baseURL string
}

func NewResend(emailCfg config.EmailConfig) (Provider, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(emailCfg.APIBaseURL), "/")
	if baseURL == "" {
		baseURL = defaultResendAPIURL
	}
	return &resendProvider{
		cfg:     emailCfg,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

type resendSendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

type resendSendResponse struct {
	ID string `json:"id"`
}

type resendErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Name       string `json:"name"`
}

func (provider *resendProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}

	to, originalTo := resolveRecipient(provider.cfg, req.To)
	if to == "" {
		return SendResult{}, fmt.Errorf("resend send: recipient is required")
	}

	payload := resendSendRequest{
		From:    formatFromAddress(provider.cfg),
		To:      []string{to},
		Subject: strings.TrimSpace(req.Subject),
		HTML:    req.HTMLBody,
		Text:    req.TextBody,
		ReplyTo: strings.TrimSpace(req.ReplyTo),
	}
	if payload.Subject == "" {
		return SendResult{}, fmt.Errorf("resend send: subject is required")
	}
	if payload.HTML == "" && payload.Text == "" {
		return SendResult{}, fmt.Errorf("resend send: html or text body is required")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{}, fmt.Errorf("resend marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return SendResult{}, fmt.Errorf("resend build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(provider.cfg.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := provider.client.Do(httpReq)
	if err != nil {
		return SendResult{}, fmt.Errorf("resend http send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return SendResult{}, fmt.Errorf("resend read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr resendErrorResponse
		_ = json.Unmarshal(respBody, &apiErr)
		msg := strings.TrimSpace(apiErr.Message)
		if msg == "" {
			msg = strings.TrimSpace(string(respBody))
		}
		if msg == "" {
			msg = resp.Status
		}
		return SendResult{}, fmt.Errorf("resend api error (%d): %s", resp.StatusCode, redactEmailAddresses(msg))
	}

	var parsed resendSendResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return SendResult{}, fmt.Errorf("resend decode response: %w", err)
	}

	messageID := strings.TrimSpace(parsed.ID)
	if messageID == "" {
		messageID = "resend:unavailable"
	}

	result := SendResult{ProviderMessageID: messageID}
	if to != originalTo {
		result.RedirectedTo = to
	}
	return result, nil
}

// NewResendWithClient is used in tests to inject a custom HTTP client and base URL.
func NewResendWithClient(emailCfg config.EmailConfig, client *http.Client, baseURL string) Provider {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultResendAPIURL
	}
	return &resendProvider{cfg: emailCfg, client: client, baseURL: baseURL}
}
