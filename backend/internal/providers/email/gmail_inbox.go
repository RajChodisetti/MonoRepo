package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type InboxReader interface {
	ProfileHistoryID(ctx context.Context) (string, error)
	ListHistoryMessageIDs(ctx context.Context, startHistoryID string) (messageIDs []string, newHistoryID string, err error)
	ListRecentMessageIDs(ctx context.Context, query string) ([]string, error)
	GetMessage(ctx context.Context, id string) (InboxMessage, error)
}

type InboxMessage struct {
	ID           string
	ThreadID     string
	From         string
	To           string
	DeliveredTo  string
	Subject      string
	InReplyTo    string
	References   string
	RFCMessageID string
	BodyText     string
}

func NewGmailInbox(emailCfg config.EmailConfig, gmailCfg config.GmailMailConfig) (InboxReader, error) {
	provider, err := NewGmail(emailCfg, gmailCfg)
	if err != nil {
		return nil, err
	}
	gmail, ok := provider.(*gmailProvider)
	if !ok {
		return nil, fmt.Errorf("gmail inbox requires the Gmail provider")
	}
	return gmail, nil
}

func (provider *gmailProvider) ProfileHistoryID(ctx context.Context) (string, error) {
	var parsed struct {
		HistoryID uint64 `json:"historyId"`
	}
	if err := provider.gmailGet(ctx, "/users/me/profile", nil, &parsed); err != nil {
		return "", err
	}
	if parsed.HistoryID == 0 {
		return "", fmt.Errorf("gmail profile history id is missing")
	}
	return fmt.Sprintf("%d", parsed.HistoryID), nil
}

func (provider *gmailProvider) ListHistoryMessageIDs(ctx context.Context, startHistoryID string) ([]string, string, error) {
	startHistoryID = strings.TrimSpace(startHistoryID)
	if startHistoryID == "" {
		return nil, "", fmt.Errorf("gmail history start id is required")
	}
	query := url.Values{}
	query.Set("startHistoryId", startHistoryID)
	query.Set("historyTypes", "messageAdded")
	var parsed struct {
		History []struct {
			MessagesAdded []struct {
				Message struct {
					ID string `json:"id"`
				} `json:"message"`
			} `json:"messagesAdded"`
		} `json:"history"`
		HistoryID uint64 `json:"historyId"`
	}
	if err := provider.gmailGet(ctx, "/users/me/history", query, &parsed); err != nil {
		return nil, "", err
	}
	seen := make(map[string]struct{})
	ids := make([]string, 0)
	for _, history := range parsed.History {
		for _, added := range history.MessagesAdded {
			id := strings.TrimSpace(added.Message.ID)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	newHistoryID := startHistoryID
	if parsed.HistoryID > 0 {
		newHistoryID = fmt.Sprintf("%d", parsed.HistoryID)
	}
	return ids, newHistoryID, nil
}

func (provider *gmailProvider) ListRecentMessageIDs(ctx context.Context, queryText string) ([]string, error) {
	query := url.Values{}
	query.Set("maxResults", "50")
	if strings.TrimSpace(queryText) != "" {
		query.Set("q", strings.TrimSpace(queryText))
	}
	var parsed struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := provider.gmailGet(ctx, "/users/me/messages", query, &parsed); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(parsed.Messages))
	for _, message := range parsed.Messages {
		id := strings.TrimSpace(message.ID)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (provider *gmailProvider) GetMessage(ctx context.Context, id string) (InboxMessage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return InboxMessage{}, fmt.Errorf("gmail message id is required")
	}
	query := url.Values{}
	query.Set("format", "full")
	var parsed gmailMessagePayload
	if err := provider.gmailGet(ctx, "/users/me/messages/"+url.PathEscape(id), query, &parsed); err != nil {
		return InboxMessage{}, err
	}
	headers := parsed.headerMap()
	return InboxMessage{
		ID:           strings.TrimSpace(parsed.ID),
		ThreadID:     strings.TrimSpace(parsed.ThreadID),
		From:         headerAddress(headers["from"]),
		To:           headers["to"],
		DeliveredTo:  headers["delivered-to"],
		Subject:      headers["subject"],
		InReplyTo:    headers["in-reply-to"],
		References:   headers["references"],
		RFCMessageID: headers["message-id"],
		BodyText:     parsed.plainText(),
	}, nil
}

type gmailMessagePayload struct {
	ID       string        `json:"id"`
	ThreadID string        `json:"threadId"`
	Payload  gmailMIMEPart `json:"payload"`
}

type gmailMIMEPart struct {
	MimeType string `json:"mimeType"`
	Headers  []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"headers"`
	Body struct {
		Data string `json:"data"`
	} `json:"body"`
	Parts []gmailMIMEPart `json:"parts"`
}

func (payload gmailMessagePayload) headerMap() map[string]string {
	headers := make(map[string]string)
	for _, header := range payload.Payload.Headers {
		name := strings.ToLower(strings.TrimSpace(header.Name))
		if name == "" {
			continue
		}
		headers[name] = strings.TrimSpace(header.Value)
	}
	return headers
}

func (payload gmailMessagePayload) plainText() string {
	if text := payload.Payload.plainText(); text != "" {
		return text
	}
	return strings.TrimSpace(payload.Payload.htmlText())
}

func (part gmailMIMEPart) plainText() string {
	if strings.EqualFold(strings.TrimSpace(part.MimeType), "text/plain") {
		if decoded := decodeGmailBody(part.Body.Data); decoded != "" {
			return decoded
		}
	}
	for _, child := range part.Parts {
		if text := child.plainText(); text != "" {
			return text
		}
	}
	return ""
}

func (part gmailMIMEPart) htmlText() string {
	if strings.EqualFold(strings.TrimSpace(part.MimeType), "text/html") {
		if decoded := decodeGmailBody(part.Body.Data); decoded != "" {
			return decoded
		}
	}
	for _, child := range part.Parts {
		if text := child.htmlText(); text != "" {
			return text
		}
	}
	return ""
}

func decodeGmailBody(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(raw)
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(decoded))
}

func headerAddress(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil {
		return strings.ToLower(value)
	}
	return strings.ToLower(strings.TrimSpace(parsed.Address))
}

func (provider *gmailProvider) gmailGet(ctx context.Context, path string, query url.Values, dest any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	accessToken, err := provider.accessToken(ctx)
	if err != nil {
		return err
	}
	endpoint := provider.apiBase + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("gmail inbox request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Accept", "application/json")
	resp, err := provider.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("gmail inbox HTTPS get: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("gmail inbox read: %w", err)
	}
	var apiErr struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &apiErr)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(apiErr.Error.Message)
		if message == "" {
			message = strings.TrimSpace(apiErr.Error.Status)
		}
		if message == "" {
			message = resp.Status
		}
		return fmt.Errorf("gmail inbox API error (%d): %s", resp.StatusCode, redactEmailAddresses(message))
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("gmail inbox decode: %w", err)
	}
	return nil
}

var _ InboxReader = (*gmailProvider)(nil)
