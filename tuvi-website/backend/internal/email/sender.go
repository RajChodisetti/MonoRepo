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
)

const (
	ProviderZoho   = "zoho"
	ProviderResend = "resend"
)

type Config struct {
	Provider    string
	FromAddress string
	FromName    string
	Disabled    bool

	ZohoAccountID    string
	ZohoFromEmail    string
	ZohoRegion       string
	ZohoAPIBaseURL   string
	ZohoClientID     string
	ZohoClientSecret string
	ZohoRefreshToken string

	ResendAPIKey     string
	ResendAPIBaseURL string
}

type Message struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
	ReplyTo  string
}

type Sender struct {
	cfg    Config
	client *http.Client
}

func NewSender(cfg Config) *Sender {
	if strings.TrimSpace(cfg.ZohoAPIBaseURL) == "" {
		cfg.ZohoAPIBaseURL = "https://mail.zoho.com/api/accounts"
	}
	if strings.TrimSpace(cfg.ResendAPIBaseURL) == "" {
		cfg.ResendAPIBaseURL = "https://api.resend.com"
	}
	if strings.TrimSpace(cfg.Provider) == "" {
		cfg.Provider = ProviderZoho
	}
	return &Sender{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Sender) Send(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.cfg.Disabled {
		return nil
	}

	provider := strings.ToLower(strings.TrimSpace(s.cfg.Provider))
	switch provider {
	case ProviderZoho, "http", "https":
		return s.sendZoho(ctx, msg)
	case ProviderResend:
		return s.sendResend(ctx, msg)
	case "smtp":
		return fmt.Errorf("smtp is no longer supported; set EMAIL_PROVIDER=zoho")
	default:
		return fmt.Errorf("unsupported email provider %q", s.cfg.Provider)
	}
}

func (s *Sender) sendZoho(ctx context.Context, msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}

	accessToken, err := s.refreshZohoToken(ctx)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(msg.HTMLBody)
	mailFormat := "html"
	if content == "" {
		content = strings.TrimSpace(msg.TextBody)
		mailFormat = "plaintext"
	}
	if content == "" {
		return fmt.Errorf("html or text body is required")
	}

	fromAddress := strings.TrimSpace(s.cfg.ZohoFromEmail)
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(s.cfg.FromAddress)
	}

	payload := map[string]string{
		"fromAddress": fromAddress,
		"toAddress":   to,
		"subject":     strings.TrimSpace(msg.Subject),
		"content":     content,
		"mailFormat":  mailFormat,
		"action":      "sendMessage",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("zoho marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/%s/messages", strings.TrimRight(s.cfg.ZohoAPIBaseURL, "/"), s.cfg.ZohoAccountID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("zoho build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("zoho http send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("zoho read response: %w", err)
	}

	var parsed struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"description"`
		} `json:"status"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("zoho decode response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || (parsed.Status.Code != 0 && parsed.Status.Code != 200 && parsed.Status.Code != 201) {
		msgText := strings.TrimSpace(parsed.Status.Message)
		if msgText == "" {
			msgText = strings.TrimSpace(string(respBody))
		}
		if msgText == "" {
			msgText = resp.Status
		}
		return fmt.Errorf("zoho api error (%d): %s", resp.StatusCode, msgText)
	}
	return nil
}

func (s *Sender) refreshZohoToken(ctx context.Context) (string, error) {
	region := strings.TrimSpace(s.cfg.ZohoRegion)
	if region == "" {
		region = "com"
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", strings.TrimSpace(s.cfg.ZohoClientID))
	form.Set("client_secret", strings.TrimSpace(s.cfg.ZohoClientSecret))
	form.Set("refresh_token", strings.TrimSpace(s.cfg.ZohoRefreshToken))

	tokenURL := fmt.Sprintf("https://accounts.zoho.%s/oauth/v2/token", region)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("zoho token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(httpReq)
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

func (s *Sender) sendResend(ctx context.Context, msg Message) error {
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}
	if strings.TrimSpace(s.cfg.ResendAPIKey) == "" {
		return fmt.Errorf("RESEND_API_KEY is required when EMAIL_PROVIDER=resend")
	}

	from := formatFromAddress(s.cfg.FromAddress, s.cfg.FromName)
	payload := map[string]any{
		"from":    from,
		"to":      []string{to},
		"subject": strings.TrimSpace(msg.Subject),
	}
	if htmlBody := strings.TrimSpace(msg.HTMLBody); htmlBody != "" {
		payload["html"] = htmlBody
	}
	if textBody := strings.TrimSpace(msg.TextBody); textBody != "" {
		payload["text"] = textBody
	}
	if replyTo := strings.TrimSpace(msg.ReplyTo); replyTo != "" {
		payload["reply_to"] = replyTo
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("resend marshal request: %w", err)
	}

	apiURL := strings.TrimRight(s.cfg.ResendAPIBaseURL, "/") + "/emails"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(s.cfg.ResendAPIKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("resend http send: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("resend read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend api error (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func formatFromAddress(fromAddress, fromName string) string {
	from := strings.TrimSpace(fromAddress)
	fromName = strings.TrimSpace(fromName)
	if fromName != "" {
		return fromName + " <" + from + ">"
	}
	return from
}
