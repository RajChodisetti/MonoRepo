package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type zohoProvider struct {
	cfg    config.ZohoMailConfig
	email  config.EmailConfig
	client *http.Client
}

func NewZoho(emailCfg config.EmailConfig, zohoCfg config.ZohoMailConfig) (Provider, error) {
	return &zohoProvider{
		cfg:   zohoCfg,
		email: emailCfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (provider *zohoProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}

	to, originalTo := resolveRecipient(provider.email, req.To)
	if to == "" {
		return SendResult{}, fmt.Errorf("zoho send: recipient is required")
	}

	accessToken, err := provider.refreshAccessToken(ctx)
	if err != nil {
		return SendResult{}, err
	}

	content := strings.TrimSpace(req.HTMLBody)
	mailFormat := "html"
	if content == "" {
		content = strings.TrimSpace(req.TextBody)
		mailFormat = "plaintext"
	}
	if content == "" {
		return SendResult{}, fmt.Errorf("zoho send: html or text body is required")
	}

	fromAddress := strings.TrimSpace(provider.cfg.FromEmail)
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(provider.email.FromAddress)
	}

	payload := map[string]string{
		"fromAddress": fromAddress,
		"toAddress":   to,
		"subject":     strings.TrimSpace(req.Subject),
		"content":     content,
		"mailFormat":  mailFormat,
		"action":      "sendMessage",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendResult{}, fmt.Errorf("zoho marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/%s/messages", strings.TrimRight(provider.cfg.APIBaseURL, "/"), provider.cfg.AccountID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return SendResult{}, fmt.Errorf("zoho build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := provider.client.Do(httpReq)
	if err != nil {
		return SendResult{}, fmt.Errorf("zoho http send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return SendResult{}, fmt.Errorf("zoho read response: %w", err)
	}

	var parsed struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"description"`
		} `json:"status"`
		Data struct {
			MessageID string `json:"messageId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return SendResult{}, fmt.Errorf("zoho decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 || (parsed.Status.Code != 0 && parsed.Status.Code != 200 && parsed.Status.Code != 201) {
		msg := strings.TrimSpace(parsed.Status.Message)
		if msg == "" {
			msg = strings.TrimSpace(string(respBody))
		}
		if msg == "" {
			msg = resp.Status
		}
		return SendResult{}, fmt.Errorf("zoho api error (%d): %s", resp.StatusCode, msg)
	}

	messageID := strings.TrimSpace(parsed.Data.MessageID)
	if messageID == "" {
		messageID = "zoho:" + originalTo
	}

	result := SendResult{ProviderMessageID: messageID}
	if to != originalTo {
		result.RedirectedTo = to
	}
	return result, nil
}

func (provider *zohoProvider) refreshAccessToken(ctx context.Context) (string, error) {
	region := strings.TrimSpace(provider.cfg.Region)
	if region == "" {
		region = "com"
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", strings.TrimSpace(provider.cfg.ClientID))
	form.Set("client_secret", strings.TrimSpace(provider.cfg.ClientSecret))
	form.Set("refresh_token", strings.TrimSpace(provider.cfg.RefreshToken))

	tokenURL := fmt.Sprintf("https://accounts.zoho.%s/oauth/v2/token", region)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("zoho token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := provider.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("zoho token http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("zoho token read: %w", err)
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("zoho token decode: %w", err)
	}
	if parsed.AccessToken == "" {
		msg := strings.TrimSpace(parsed.Error)
		if msg == "" {
			msg = strings.TrimSpace(string(respBody))
		}
		return "", fmt.Errorf("zoho token refresh failed: %s", msg)
	}
	return parsed.AccessToken, nil
}
