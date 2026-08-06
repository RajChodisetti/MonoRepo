package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

type contactMailer interface {
	Send(ctx context.Context, req emailprovider.SendRequest) (emailprovider.SendResult, error)
}

// ContactPublicHandler accepts website demo/contact form submissions.
type ContactPublicHandler struct {
	mailer     contactMailer
	notifyTo   string
	writeJSON  func(http.ResponseWriter, int, any)
	writeError func(http.ResponseWriter, int, string, string)
}

func NewContactPublicHandler(
	mailer contactMailer,
	notifyTo string,
	writeJSON func(http.ResponseWriter, int, any),
	writeError func(http.ResponseWriter, int, string, string),
) *ContactPublicHandler {
	to := strings.TrimSpace(notifyTo)
	if to == "" {
		to = "contact@tuvisolutions.com"
	}
	return &ContactPublicHandler{
		mailer:     mailer,
		notifyTo:   to,
		writeJSON:  writeJSON,
		writeError: writeError,
	}
}

type contactRequestBody struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	RestaurantName string `json:"restaurantName"`
	City           string `json:"city"`
	Message        string `json:"message"`
	Source         string `json:"source"`
}

// Submit handles POST /api/public/v1/contact
func (handler *ContactPublicHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var body contactRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be JSON.")
		return
	}

	name := strings.TrimSpace(body.Name)
	emailAddr := strings.TrimSpace(body.Email)
	phone := strings.TrimSpace(body.Phone)
	restaurant := strings.TrimSpace(body.RestaurantName)
	city := strings.TrimSpace(body.City)
	message := strings.TrimSpace(body.Message)
	source := strings.TrimSpace(body.Source)
	if source == "" {
		source = "website-demo"
	}

	if name == "" || utf8.RuneCountInString(name) > 120 {
		handler.writeError(w, http.StatusBadRequest, "invalid_name", "Name is required.")
		return
	}
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		handler.writeError(w, http.StatusBadRequest, "invalid_email", "A valid email is required.")
		return
	}
	if restaurant == "" || utf8.RuneCountInString(restaurant) > 160 {
		handler.writeError(w, http.StatusBadRequest, "invalid_restaurant", "Restaurant name is required.")
		return
	}
	if utf8.RuneCountInString(phone) > 40 {
		handler.writeError(w, http.StatusBadRequest, "invalid_phone", "Phone is too long.")
		return
	}
	if utf8.RuneCountInString(city) > 80 {
		handler.writeError(w, http.StatusBadRequest, "invalid_city", "City is too long.")
		return
	}
	if utf8.RuneCountInString(message) > 2000 {
		handler.writeError(w, http.StatusBadRequest, "invalid_message", "Message is too long.")
		return
	}
	if handler.mailer == nil {
		handler.writeError(w, http.StatusServiceUnavailable, "mailer_unavailable", "Contact email is not configured.")
		return
	}

	subject := fmt.Sprintf("Tuvi demo request — %s", restaurant)
	text := fmt.Sprintf(
		"New Tuvi website demo / contact request\n\nName: %s\nEmail: %s\nPhone: %s\nRestaurant: %s\nCity: %s\nSource: %s\n\nMessage:\n%s\n",
		name, emailAddr, emptyDash(phone), restaurant, emptyDash(city), source, emptyDash(message),
	)
	htmlBody := fmt.Sprintf(
		`<div style="font-family:system-ui,sans-serif;line-height:1.5;color:#14241c">
  <h2 style="margin:0 0 12px">New Tuvi demo / contact request</h2>
  <p style="margin:0 0 8px"><strong>Name:</strong> %s</p>
  <p style="margin:0 0 8px"><strong>Email:</strong> %s</p>
  <p style="margin:0 0 8px"><strong>Phone:</strong> %s</p>
  <p style="margin:0 0 8px"><strong>Restaurant:</strong> %s</p>
  <p style="margin:0 0 8px"><strong>City:</strong> %s</p>
  <p style="margin:0 0 8px"><strong>Source:</strong> %s</p>
  <p style="margin:16px 0 0"><strong>Message</strong></p>
  <p style="white-space:pre-wrap;margin:4px 0 0">%s</p>
</div>`,
		html.EscapeString(name),
		html.EscapeString(emailAddr),
		html.EscapeString(emptyDash(phone)),
		html.EscapeString(restaurant),
		html.EscapeString(emptyDash(city)),
		html.EscapeString(source),
		html.EscapeString(emptyDash(message)),
	)

	_, err := handler.mailer.Send(r.Context(), emailprovider.SendRequest{
		To:       handler.notifyTo,
		Subject:  subject,
		TextBody: text,
		HTMLBody: htmlBody,
		ReplyTo:  emailAddr,
		Metadata: map[string]string{
			"kind":       "website_contact",
			"source":     source,
			"restaurant": restaurant,
		},
	})
	if err != nil {
		handler.writeError(w, http.StatusBadGateway, "email_failed", "Could not send your message. Please try again.")
		return
	}

	handler.writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Thanks — we received your request and will reply soon.",
	})
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
