package consultations

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// GoogleCalendarAddURL builds a Google Calendar "add event" link for the booking
// slot. Opening it marks "Tuvi Consultation" on the user's calendar at that time.
func GoogleCalendarAddURL(start, end time.Time, title, details, location string) string {
	if start.IsZero() {
		return ""
	}
	if end.IsZero() || !end.After(start) {
		end = start.Add(30 * time.Minute)
	}
	if strings.TrimSpace(title) == "" {
		title = "Tuvi Consultation"
	}

	loc := start.Location()
	if loc == nil || loc == time.UTC {
		// Prefer business timezone when available on the start time.
		if sydney, err := time.LoadLocation("Australia/Sydney"); err == nil {
			loc = sydney
			start = start.In(loc)
			end = end.In(loc)
		}
	} else {
		start = start.In(loc)
		end = end.In(loc)
	}

	// Local wall-clock format + ctz keeps the slot where the user booked it.
	dates := fmt.Sprintf("%s/%s", start.Format("20060102T150405"), end.Format("20060102T150405"))

	q := url.Values{}
	q.Set("action", "TEMPLATE")
	q.Set("text", title)
	q.Set("dates", dates)
	q.Set("ctz", loc.String())
	if details = strings.TrimSpace(details); details != "" {
		q.Set("details", details)
	}
	if location = strings.TrimSpace(location); location != "" {
		q.Set("location", location)
	}

	return "https://calendar.google.com/calendar/render?" + q.Encode()
}
