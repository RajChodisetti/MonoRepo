package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Google struct {
	service    *calendar.Service
	calendarID string
}

func NewGoogle(calendarID, credentialsPath string) (*Google, error) {
	ctx := context.Background()
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read service account json: %w", err)
	}
	creds, err := google.CredentialsFromJSON(ctx, data, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("parse service account credentials: %w", err)
	}
	service, err := calendar.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("create calendar service: %w", err)
	}
	return &Google{service: service, calendarID: calendarID}, nil
}

// EnsureWritable picks a calendar the service account can write to.
// If the configured calendar is not shared, it creates (or reuses) a dedicated calendar.
func EnsureWritable(ctx context.Context, calendarID, credentialsPath string) (*Google, string, error) {
	g, err := NewGoogle(calendarID, credentialsPath)
	if err != nil {
		return nil, "", err
	}
	if err := g.pingWrite(ctx); err == nil {
		return g, calendarID, nil
	}

	// Reuse an existing owned calendar if present.
	list, err := g.service.CalendarList.List().Context(ctx).Do()
	if err == nil {
		for _, item := range list.Items {
			if item.AccessRole == "owner" && item.Summary == "Tuvi Consultations" {
				owned := NewFromService(g.service, item.Id)
				if owned.pingWrite(ctx) == nil {
					return owned, item.Id, nil
				}
			}
		}
	}

	created, err := g.service.Calendars.Insert(&calendar.Calendar{
		Summary:     "Tuvi Consultations",
		Description: "Bookings from Tuvi website and voice assistant",
		TimeZone:    "Australia/Sydney",
	}).Context(ctx).Do()
	if err != nil {
		// Last resort: service account primary calendar.
		saID, saErr := serviceAccountEmail(credentialsPath)
		if saErr != nil {
			return nil, "", fmt.Errorf("create calendar: %w", err)
		}
		fallback := NewFromService(g.service, saID)
		if pingErr := fallback.pingWrite(ctx); pingErr != nil {
			return nil, "", fmt.Errorf(
				"calendar %q not writable (share it with the service account) and fallback failed: %w",
				calendarID, pingErr,
			)
		}
		return fallback, saID, nil
	}

	owned := NewFromService(g.service, created.Id)
	if err := owned.pingWrite(ctx); err != nil {
		return nil, "", fmt.Errorf("created calendar not writable: %w", err)
	}
	return owned, created.Id, nil
}

func NewFromService(service *calendar.Service, calendarID string) *Google {
	return &Google{service: service, calendarID: calendarID}
}

func serviceAccountEmail(credentialsPath string) (string, error) {
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return "", err
	}
	var meta struct {
		ClientEmail string `json:"client_email"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", err
	}
	if meta.ClientEmail == "" {
		return "", fmt.Errorf("client_email missing in service account json")
	}
	return meta.ClientEmail, nil
}

func (g *Google) pingWrite(ctx context.Context) error {
	start := time.Now().UTC().Truncate(time.Hour).Add(24 * time.Hour)
	end := start.Add(5 * time.Minute)
	res, err := g.CreateEvent(ctx, CreateEventInput{
		Title:       "Tuvi calendar connectivity check",
		Description: "auto-generated; safe to delete",
		Start:       start,
		End:         end,
	})
	if err != nil {
		return err
	}
	_ = g.DeleteEvent(ctx, res.EventID)
	return nil
}

func (g *Google) FreeBusy(ctx context.Context, start, end time.Time) ([]BusyPeriod, error) {
	req := &calendar.FreeBusyRequest{
		TimeMin: start.UTC().Format(time.RFC3339),
		TimeMax: end.UTC().Format(time.RFC3339),
		Items:   []*calendar.FreeBusyRequestItem{{Id: g.calendarID}},
	}
	resp, err := g.service.Freebusy.Query(req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("freebusy query: %w", err)
	}
	cal, ok := resp.Calendars[g.calendarID]
	if !ok {
		return nil, fmt.Errorf("calendar %q not in freebusy response", g.calendarID)
	}
	periods := make([]BusyPeriod, 0, len(cal.Busy))
	for _, busy := range cal.Busy {
		bs, err := time.Parse(time.RFC3339, busy.Start)
		if err != nil {
			continue
		}
		be, err := time.Parse(time.RFC3339, busy.End)
		if err != nil {
			continue
		}
		periods = append(periods, BusyPeriod{Start: bs, End: be})
	}
	return periods, nil
}

func (g *Google) CreateEvent(ctx context.Context, input CreateEventInput) (CreateEventResult, error) {
	desc := input.Description
	if input.Attendee != "" {
		if desc != "" {
			desc += "\n"
		}
		desc += "Guest email: " + input.Attendee
	}
	event := &calendar.Event{
		Summary:     input.Title,
		Description: desc,
		Start: &calendar.EventDateTime{
			DateTime: input.Start.Format(time.RFC3339),
			TimeZone: input.Start.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: input.End.Format(time.RFC3339),
			TimeZone: input.End.Location().String(),
		},
	}
	created, err := g.service.Events.Insert(g.calendarID, event).Context(ctx).Do()
	if err != nil {
		return CreateEventResult{}, fmt.Errorf("create calendar event: %w", err)
	}
	return CreateEventResult{EventID: created.Id, HTMLLink: created.HtmlLink}, nil
}

func (g *Google) DeleteEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		return nil
	}
	return g.service.Events.Delete(g.calendarID, eventID).Context(ctx).Do()
}
