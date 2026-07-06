package consultations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	calendarprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/calendar"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

var ErrConflict = errors.New("slot already booked")

type Service struct {
	cfg      config.ConsultationConfig
	repo     Repository
	calendar calendarprovider.Provider
	email    emailprovider.Provider
	log      *slog.Logger
	slotCfg  SlotConfig
}

func NewService(
	cfg config.ConsultationConfig,
	repo Repository,
	cal calendarprovider.Provider,
	email emailprovider.Provider,
	log *slog.Logger,
) *Service {
	return &Service{
		cfg:      cfg,
		repo:     repo,
		calendar: cal,
		email:    email,
		log:      log,
		slotCfg: SlotConfig{
			Timezone:            cfg.Timezone,
			BusinessHourStart:   cfg.BusinessHourStart,
			BusinessHourEnd:     cfg.BusinessHourEnd,
			SlotDurationMinutes: cfg.SlotDurationMinutes,
			HorizonDays:         cfg.AvailabilityHorizon,
		},
	}
}

type AvailabilityResult struct {
	Status string `json:"status"`
	Slots  []Slot `json:"slots"`
}

type CheckResult struct {
	Available    bool     `json:"available"`
	Alternatives []string `json:"alternatives,omitempty"`
	Slot         string   `json:"slot,omitempty"`
	Date         string   `json:"date,omitempty"`
	Time         string   `json:"time,omitempty"`
}

type BookRequest struct {
	Date          string `json:"date"`
	Time          string `json:"time"`
	ProspectName  string `json:"prospect_name"`
	ProspectEmail string `json:"prospect_email"`
	ProspectPhone string `json:"prospect_phone"`
	Source        string `json:"source"`
}

type BookSuccess struct {
	Status           string `json:"status"`
	ConfirmationCode string `json:"confirmation_code"`
	ProspectName     string `json:"prospect_name"`
	ProspectEmail    string `json:"prospect_email"`
	Slot             string `json:"slot"`
	BookingDate      string `json:"booking_date"`
	BookingTime      string `json:"booking_time"`
	CalendarLink     string `json:"calendar_link"`
	Message          string `json:"message"`
}

type BookConflict struct {
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	Alternatives []string `json:"alternatives"`
}

func (s *Service) GetAvailability(ctx context.Context, dateStr string, days int) (AvailabilityResult, error) {
	if days < 1 {
		days = s.cfg.DefaultAvailDays
	}
	if days > s.cfg.AvailabilityHorizon {
		days = s.cfg.AvailabilityHorizon
	}

	startDate, err := s.resolveStartDate(dateStr)
	if err != nil {
		return AvailabilityResult{}, err
	}

	free, err := s.listFreeSlots(ctx, startDate, days)
	if err != nil {
		return AvailabilityResult{}, err
	}

	slots := make([]Slot, 0, len(free))
	for _, t := range free {
		slots = append(slots, ToSlotDTO(t, true))
	}
	return AvailabilityResult{Status: "success", Slots: slots}, nil
}

func (s *Service) CheckSlot(ctx context.Context, dateStr, timeStr string) (CheckResult, error) {
	slotStart, slotEnd, err := s.parseSlot(dateStr, timeStr)
	if err != nil {
		return CheckResult{}, err
	}

	available, err := s.isSlotFree(ctx, slotStart, slotEnd)
	if err != nil {
		return CheckResult{}, err
	}
	result := CheckResult{
		Available: available,
		Slot:      slotStart.Format(time.RFC3339),
		Date:      slotStart.Format("2006-01-02"),
		Time:      FormatSlotTime(slotStart),
	}
	if !available {
		alts, err := s.nearestAlternatives(ctx, slotStart, 3)
		if err != nil {
			return CheckResult{}, err
		}
		result.Alternatives = alts
	}
	return result, nil
}

func (s *Service) Book(ctx context.Context, req BookRequest) (BookSuccess, *BookConflict, error) {
	req.ProspectName = strings.TrimSpace(req.ProspectName)
	req.ProspectEmail = strings.TrimSpace(req.ProspectEmail)
	req.ProspectPhone = strings.TrimSpace(req.ProspectPhone)
	req.Source = strings.TrimSpace(req.Source)
	if req.Source == "" {
		req.Source = SourceVoice
	}
	if req.Source != SourceVoice && req.Source != SourceWeb {
		return BookSuccess{}, nil, fmt.Errorf("source must be voice or web")
	}
	if req.ProspectName == "" {
		return BookSuccess{}, nil, fmt.Errorf("prospect_name is required")
	}
	if req.ProspectEmail == "" {
		return BookSuccess{}, nil, fmt.Errorf("prospect_email is required")
	}

	slotStart, slotEnd, err := s.parseSlot(req.Date, req.Time)
	if err != nil {
		return BookSuccess{}, nil, err
	}

	free, err := s.isSlotFree(ctx, slotStart, slotEnd)
	if err != nil {
		return BookSuccess{}, nil, err
	}
	if !free {
		alts, err := s.nearestAlternatives(ctx, slotStart, 3)
		if err != nil {
			return BookSuccess{}, nil, err
		}
		return BookSuccess{}, &BookConflict{
			Status:       "conflict",
			Message:      "That slot is already booked.",
			Alternatives: alts,
		}, nil
	}

	code, err := GenerateConfirmationCode()
	if err != nil {
		return BookSuccess{}, nil, err
	}

	id := uuid.New()
	desc := fmt.Sprintf(
		"Consultation booking\nName: %s\nEmail: %s\nPhone: %s\nConfirmation: %s\nSource: %s",
		req.ProspectName, req.ProspectEmail, req.ProspectPhone, code, req.Source,
	)
	event, err := s.calendar.CreateEvent(ctx, calendarprovider.CreateEventInput{
		Title:       fmt.Sprintf("Tuvi Consultation - %s", req.ProspectName),
		Description: desc,
		Start:       slotStart,
		End:         slotEnd,
		Attendee:    req.ProspectEmail,
	})
	if err != nil {
		return BookSuccess{}, nil, fmt.Errorf("calendar booking failed: %w", err)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		_ = s.calendar.DeleteEvent(ctx, event.EventID)
		return BookSuccess{}, nil, err
	}
	defer tx.Rollback(ctx)

	err = s.repo.Insert(ctx, tx, InsertInput{
		ID:               id,
		ConfirmationCode: code,
		SlotStart:        slotStart,
		SlotEnd:          slotEnd,
		ProspectName:     req.ProspectName,
		ProspectEmail:    req.ProspectEmail,
		ProspectPhone:    req.ProspectPhone,
		GoogleEventID:    event.EventID,
		Source:           req.Source,
	})
	if err != nil {
		_ = s.calendar.DeleteEvent(ctx, event.EventID)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			alts, altErr := s.nearestAlternatives(ctx, slotStart, 3)
			if altErr != nil {
				return BookSuccess{}, nil, altErr
			}
			return BookSuccess{}, &BookConflict{
				Status:       "conflict",
				Message:      "That slot is already booked.",
				Alternatives: alts,
			}, nil
		}
		return BookSuccess{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		_ = s.calendar.DeleteEvent(ctx, event.EventID)
		return BookSuccess{}, nil, err
	}

	go s.sendBookingEmail(context.Background(), req, code, slotStart, event.HTMLLink)

	return BookSuccess{
		Status:           "success",
		ConfirmationCode: code,
		ProspectName:     req.ProspectName,
		ProspectEmail:    req.ProspectEmail,
		Slot:             slotStart.Format(time.RFC3339),
		BookingDate:      slotStart.Format("2006-01-02"),
		BookingTime:      FormatSlotTime(slotStart),
		CalendarLink:     event.HTMLLink,
		Message:          fmt.Sprintf("Your consultation is booked. Confirmation number %s.", code),
	}, nil, nil
}

func (s *Service) resolveStartDate(dateStr string) (time.Time, error) {
	loc := s.cfg.Timezone
	if strings.TrimSpace(dateStr) == "" {
		now := time.Now().In(loc)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	}
	return ParseDate(dateStr, loc)
}

func (s *Service) parseSlot(dateStr, timeStr string) (time.Time, time.Time, error) {
	date, err := ParseDate(dateStr, s.cfg.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	slotStart, err := BuildSlotStart(date, timeStr, s.cfg.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	now := time.Now().In(s.cfg.Timezone)
	if slotStart.Before(now) {
		return time.Time{}, time.Time{}, fmt.Errorf("cannot book a slot in the past")
	}
	if slotStart.Weekday() == time.Saturday || slotStart.Weekday() == time.Sunday {
		return time.Time{}, time.Time{}, fmt.Errorf("consultations are only available on weekdays")
	}
	hour := slotStart.Hour()
	if hour < s.cfg.BusinessHourStart || hour >= s.cfg.BusinessHourEnd {
		return time.Time{}, time.Time{}, fmt.Errorf("time must be between %02d:00 and %02d:00", s.cfg.BusinessHourStart, s.cfg.BusinessHourEnd)
	}
	slotEnd := slotStart.Add(time.Duration(s.cfg.SlotDurationMinutes) * time.Minute)
	return slotStart, slotEnd, nil
}

func (s *Service) listFreeSlots(ctx context.Context, startDate time.Time, days int) ([]time.Time, error) {
	candidates := GenerateCandidateSlots(s.slotCfg, startDate, days)
	if len(candidates) == 0 {
		return nil, nil
	}

	rangeStart := candidates[0]
	rangeEnd := candidates[len(candidates)-1].Add(time.Duration(s.cfg.SlotDurationMinutes) * time.Minute)

	booked, err := s.repo.BookedSlotStarts(ctx, rangeStart, rangeEnd.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	busy, err := s.calendar.FreeBusy(ctx, rangeStart, rangeEnd.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	var free []time.Time
	duration := time.Duration(s.cfg.SlotDurationMinutes) * time.Minute
	for _, slot := range candidates {
		if ContainsTime(booked, slot) {
			continue
		}
		slotEnd := slot.Add(duration)
		if overlapsBusy(slot, slotEnd, busy) {
			continue
		}
		free = append(free, slot)
	}
	return free, nil
}

func (s *Service) isSlotFree(ctx context.Context, slotStart, slotEnd time.Time) (bool, error) {
	booked, err := s.repo.IsSlotBooked(ctx, slotStart)
	if err != nil {
		return false, err
	}
	if booked {
		return false, nil
	}
	busy, err := s.calendar.FreeBusy(ctx, slotStart, slotEnd)
	if err != nil {
		return false, err
	}
	return !overlapsBusy(slotStart, slotEnd, busy), nil
}

func (s *Service) nearestAlternatives(ctx context.Context, around time.Time, limit int) ([]string, error) {
	loc := s.cfg.Timezone
	start := time.Date(around.Year(), around.Month(), around.Day(), 0, 0, 0, 0, loc)
	free, err := s.listFreeSlots(ctx, start, s.cfg.DefaultAvailDays+3)
	if err != nil {
		return nil, err
	}
	var alts []string
	for _, slot := range free {
		if slot.Equal(around) {
			continue
		}
		alts = append(alts, slot.Format(time.RFC3339))
		if len(alts) >= limit {
			break
		}
	}
	if len(alts) == 0 {
		free, err = s.listFreeSlots(ctx, start, s.cfg.AvailabilityHorizon)
		if err != nil {
			return nil, err
		}
		for _, slot := range free {
			if slot.Equal(around) {
				continue
			}
			alts = append(alts, slot.Format(time.RFC3339))
			if len(alts) >= limit {
				break
			}
		}
	}
	return alts, nil
}

func overlapsBusy(start, end time.Time, busy []calendarprovider.BusyPeriod) bool {
	for _, b := range busy {
		if start.Before(b.End) && end.After(b.Start) {
			return true
		}
	}
	return false
}

func (s *Service) sendBookingEmail(ctx context.Context, req BookRequest, code string, slotStart time.Time, calendarLink string) {
	if s.email == nil {
		return
	}

	when := slotStart.In(s.cfg.Timezone).Format("Monday, 2 January 2006 at 3:04 PM MST")
	calendarText := ""
	if calendarLink != "" {
		calendarText = fmt.Sprintf("\nOpen in Google Calendar: %s\n", calendarLink)
	}

	subject := fmt.Sprintf("New consultation booked - %s", code)
	text := fmt.Sprintf(`A new consultation has been booked.

Confirmation: %s
Name: %s
Email: %s
Phone: %s
When: %s
Source: %s%s`, code, req.ProspectName, req.ProspectEmail, req.ProspectPhone, when, req.Source, calendarText)

	html := fmt.Sprintf(`<p>A new consultation has been booked.</p>
<ul>
<li><strong>Confirmation:</strong> %s</li>
<li><strong>Name:</strong> %s</li>
<li><strong>Email:</strong> %s</li>
<li><strong>Phone:</strong> %s</li>
<li><strong>When:</strong> %s</li>
<li><strong>Source:</strong> %s</li>
</ul>`, code, req.ProspectName, req.ProspectEmail, req.ProspectPhone, when, req.Source)

	if _, err := s.email.Send(ctx, emailprovider.SendRequest{
		To:       s.cfg.NotifyEmail,
		Subject:  subject,
		TextBody: text,
		HTMLBody: html,
		ReplyTo:  req.ProspectEmail,
		Metadata: map[string]string{"confirmation_code": code, "source": req.Source},
	}); err != nil && s.log != nil {
		s.log.ErrorContext(ctx, "consultation_notification_email_failed", "error", err, "confirmation_code", code)
	}

	if req.ProspectEmail == "" {
		return
	}
	prospectSubject := fmt.Sprintf("Your Tuvi consultation is confirmed - %s", code)
	prospectText := fmt.Sprintf(
		"Hi %s,\n\nYour consultation is confirmed for %s.\nConfirmation: %s%s\nWe look forward to speaking with you.\n\nTuvi Solutions\n",
		req.ProspectName, when, code, calendarText,
	)
	prospectHTML := fmt.Sprintf(
		`<p>Hi %s,</p><p>Your consultation is confirmed for <strong>%s</strong>.</p><p>Confirmation: <strong>%s</strong></p><p>We look forward to speaking with you.</p><p>Tuvi Solutions</p>`,
		req.ProspectName, when, code,
	)
	if _, err := s.email.Send(ctx, emailprovider.SendRequest{
		To:       req.ProspectEmail,
		Subject:  prospectSubject,
		TextBody: prospectText,
		HTMLBody: prospectHTML,
		Metadata: map[string]string{"confirmation_code": code, "source": req.Source},
	}); err != nil && s.log != nil {
		s.log.ErrorContext(ctx, "consultation_confirmation_email_failed", "error", err, "confirmation_code", code)
	}
}
