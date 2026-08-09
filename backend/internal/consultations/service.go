package consultations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	emailprovider "github.com/rajchodisetti/restaurant-platform/backend/internal/providers/email"
)

var (
	ErrConflict                 = errors.New("slot already booked")
	ErrInvalidCalendar          = errors.New("invalid consultation calendar")
	ErrCalendarRevisionConflict = errors.New("consultation calendar revision conflict")
)

type Service struct {
	cfg     config.ConsultationConfig
	repo    Repository
	email   emailprovider.Provider
	log     *slog.Logger
	slotCfg SlotConfig
}

func NewService(
	cfg config.ConsultationConfig,
	repo Repository,
	email emailprovider.Provider,
	log *slog.Logger,
) *Service {
	return &Service{
		cfg:   cfg,
		repo:  repo,
		email: email,
		log:   log,
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
	ProspectPhone    string `json:"prospect_phone"`
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

func (s *Service) GetCalendar(ctx context.Context, month string) (CalendarResult, error) {
	monthStart, err := ParseMonth(month, s.cfg.Timezone)
	if err != nil {
		return CalendarResult{}, invalidCalendarError("%v", err)
	}
	return s.buildCalendar(ctx, monthStart, time.Now().In(s.cfg.Timezone))
}

func (s *Service) UpdateCalendar(
	ctx context.Context,
	month string,
	updates []CalendarSlotUpdate,
	expectedRevision int64,
	updatedBy uuid.UUID,
) (CalendarResult, error) {
	monthStart, err := ParseMonth(month, s.cfg.Timezone)
	if err != nil {
		return CalendarResult{}, invalidCalendarError("%v", err)
	}

	if expectedRevision < 0 {
		return CalendarResult{}, invalidCalendarError("expected_revision must be zero or greater")
	}

	now := time.Now().In(s.cfg.Timezone)
	allCandidates := make(map[int64]time.Time)
	expectedFuture := make(map[int64]time.Time)
	for _, slot := range GenerateMonthCandidateSlots(s.slotCfg, monthStart) {
		allCandidates[slotKey(slot)] = slot
		if slot.After(now) {
			expectedFuture[slotKey(slot)] = slot
		}
	}

	seen := make(map[int64]struct{}, len(updates))
	inputs := make([]SlotOverrideInput, 0, len(updates))
	for _, update := range updates {
		rawISO := strings.TrimSpace(update.ISO)
		parsed, err := time.Parse(time.RFC3339, rawISO)
		if err != nil {
			return CalendarResult{}, invalidCalendarError("slot %q must be RFC3339", rawISO)
		}
		key := slotKey(parsed)
		expectedSlot, ok := allCandidates[key]
		if !ok {
			return CalendarResult{}, invalidCalendarError(
				"slot %q is not an on-grid candidate in %s",
				rawISO,
				monthStart.Format("2006-01"),
			)
		}
		if rawISO != expectedSlot.Format(time.RFC3339) {
			return CalendarResult{}, invalidCalendarError(
				"slot %q must use the configured %s timezone offset",
				rawISO,
				s.cfg.Timezone.String(),
			)
		}
		if _, duplicate := seen[key]; duplicate {
			return CalendarResult{}, invalidCalendarError("slot %q is duplicated", rawISO)
		}
		seen[key] = struct{}{}
		// A slot may cross from future to past while an admin has the month
		// open. It is still validated as a canonical candidate, then ignored.
		if !expectedSlot.After(now) {
			continue
		}
		inputs = append(inputs, SlotOverrideInput{
			SlotStart:   expectedSlot,
			IsAvailable: update.IsAvailable,
		})
	}

	for key, expectedSlot := range expectedFuture {
		if _, ok := seen[key]; !ok {
			return CalendarResult{}, invalidCalendarError(
				"slots must contain every future on-grid candidate for %s; missing %q",
				monthStart.Format("2006-01"),
				expectedSlot.Format(time.RFC3339),
			)
		}
	}

	monthEnd := monthStart.AddDate(0, 1, 0)
	if _, err := s.repo.ReplaceMonthSlotOverrides(
		ctx,
		monthStart,
		monthEnd,
		inputs,
		expectedRevision,
		updatedBy,
	); err != nil {
		return CalendarResult{}, err
	}
	return s.buildCalendar(ctx, monthStart, time.Now().In(s.cfg.Timezone))
}

func (s *Service) buildCalendar(
	ctx context.Context,
	monthStart time.Time,
	now time.Time,
) (CalendarResult, error) {
	monthEnd := monthStart.AddDate(0, 1, 0)
	var (
		overrides []SlotOverride
		booked    []BookedInterval
		revision  int64
	)
	// The revision fence prevents returning an old override snapshot with a new
	// revision, which could otherwise let a client overwrite a save it never saw.
	for attempt := 0; attempt < 3; attempt++ {
		before, err := s.repo.CalendarRevision(ctx, monthStart)
		if err != nil {
			return CalendarResult{}, err
		}
		overrides, err = s.repo.SlotOverrides(ctx, monthStart, monthEnd)
		if err != nil {
			return CalendarResult{}, err
		}
		booked, err = s.repo.BookedIntervals(ctx, monthStart, monthEnd)
		if err != nil {
			return CalendarResult{}, err
		}
		after, err := s.repo.CalendarRevision(ctx, monthStart)
		if err != nil {
			return CalendarResult{}, err
		}
		if before == after {
			revision = after
			break
		}
		if attempt == 2 {
			return CalendarResult{}, ErrCalendarRevisionConflict
		}
	}

	overrideBySlot := make(map[int64]bool, len(overrides))
	for _, override := range overrides {
		overrideBySlot[slotKey(override.SlotStart)] = override.IsAvailable
	}
	type calendarSlotAt struct {
		start time.Time
		slot  CalendarSlot
	}
	entries := make([]calendarSlotAt, 0)
	onGrid := make(map[int64]struct{})
	for _, candidate := range GenerateMonthCandidateSlots(s.slotCfg, monthStart) {
		key := slotKey(candidate)
		onGrid[key] = struct{}{}
		isAvailable := true
		if override, ok := overrideBySlot[key]; ok {
			isAvailable = override
		}
		candidateEnd := candidate.Add(time.Duration(s.cfg.SlotDurationMinutes) * time.Minute)
		isBooked := false
		for _, interval := range booked {
			if intervalsOverlap(candidate, candidateEnd, interval.SlotStart, interval.SlotEnd) {
				isBooked = true
				break
			}
		}
		past := !candidate.After(now)
		entries = append(entries, calendarSlotAt{
			start: candidate,
			slot: CalendarSlot{
				Date:               candidate.Format("2006-01-02"),
				Time:               FormatSlotTime(candidate),
				ISO:                candidate.Format(time.RFC3339),
				IsAvailable:        isAvailable,
				Booked:             isBooked,
				Past:               past,
				EffectiveAvailable: isAvailable && !isBooked && !past,
				OnGrid:             true,
			},
		})
	}

	// Historical clients could submit an in-hours time that was not aligned to
	// the configured grid. Preserve visibility of those confirmed calls.
	for _, interval := range booked {
		bookedSlot := interval.SlotStart
		key := slotKey(bookedSlot)
		if _, ok := onGrid[key]; ok {
			continue
		}
		local := bookedSlot.In(s.cfg.Timezone)
		if local.Before(monthStart) || !local.Before(monthEnd) {
			continue
		}
		entries = append(entries, calendarSlotAt{
			start: local,
			slot: CalendarSlot{
				Date:               local.Format("2006-01-02"),
				Time:               FormatSlotTime(local),
				ISO:                local.Format(time.RFC3339),
				IsAvailable:        false,
				Booked:             true,
				Past:               !local.After(now),
				EffectiveAvailable: false,
				OnGrid:             false,
			},
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].start.Before(entries[j].start)
	})
	slots := make([]CalendarSlot, 0, len(entries))
	for _, entry := range entries {
		slots = append(slots, entry.slot)
	}

	return CalendarResult{
		Month:               monthStart.Format("2006-01"),
		Revision:            revision,
		BookedCallCount:     len(booked),
		Timezone:            s.cfg.Timezone.String(),
		SlotDurationMinutes: s.cfg.SlotDurationMinutes,
		BusinessHourStart:   fmt.Sprintf("%02d:00", s.cfg.BusinessHourStart),
		BusinessHourEnd:     fmt.Sprintf("%02d:00", s.cfg.BusinessHourEnd),
		Slots:               slots,
	}, nil
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
	slotStart, _, err := s.parseSlot(dateStr, timeStr)
	if err != nil {
		return CheckResult{}, err
	}

	available, err := s.isSlotFree(ctx, slotStart)
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

	code, err := GenerateConfirmationCode()
	if err != nil {
		return BookSuccess{}, nil, err
	}

	id := uuid.New()
	inserted, err := s.repo.InsertIfAvailable(ctx, InsertInput{
		ID:               id,
		ConfirmationCode: code,
		SlotStart:        slotStart,
		SlotEnd:          slotEnd,
		ProspectName:     req.ProspectName,
		ProspectEmail:    req.ProspectEmail,
		ProspectPhone:    req.ProspectPhone,
		GoogleEventID:    "",
		Source:           req.Source,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, ErrConflict) || (errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23P01")) {
			return s.bookingConflict(ctx, slotStart, "That slot is already booked.")
		}
		return BookSuccess{}, nil, err
	}
	if !inserted {
		return s.bookingConflict(ctx, slotStart, "That slot is not available.")
	}

	go s.sendBookingEmail(context.Background(), req, code, slotStart, "")

	return BookSuccess{
		Status:           "success",
		ConfirmationCode: code,
		ProspectName:     req.ProspectName,
		ProspectEmail:    req.ProspectEmail,
		ProspectPhone:    req.ProspectPhone,
		Slot:             slotStart.Format(time.RFC3339),
		BookingDate:      slotStart.Format("2006-01-02"),
		BookingTime:      FormatSlotTime(slotStart),
		CalendarLink:     "",
		Message:          fmt.Sprintf("Your consultation is booked. Confirmation number %s.", code),
	}, nil, nil
}

func (s *Service) bookingConflict(
	ctx context.Context,
	slotStart time.Time,
	message string,
) (BookSuccess, *BookConflict, error) {
	alternatives, err := s.nearestAlternatives(ctx, slotStart, 3)
	if err != nil {
		return BookSuccess{}, nil, err
	}
	return BookSuccess{}, &BookConflict{
		Status:       "conflict",
		Message:      message,
		Alternatives: alternatives,
	}, nil
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
	if !slotStart.After(now) {
		return time.Time{}, time.Time{}, fmt.Errorf("cannot book a slot in the past")
	}
	if slotStart.Weekday() == time.Saturday || slotStart.Weekday() == time.Sunday {
		return time.Time{}, time.Time{}, fmt.Errorf("consultations are only available on weekdays")
	}
	hour := slotStart.Hour()
	if hour < s.cfg.BusinessHourStart || hour >= s.cfg.BusinessHourEnd {
		return time.Time{}, time.Time{}, fmt.Errorf("time must be between %02d:00 and %02d:00", s.cfg.BusinessHourStart, s.cfg.BusinessHourEnd)
	}
	if !IsCandidateSlot(s.slotCfg, slotStart) {
		return time.Time{}, time.Time{}, fmt.Errorf(
			"time must align to a %d-minute consultation slot",
			s.cfg.SlotDurationMinutes,
		)
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

	booked, err := s.repo.BookedIntervals(ctx, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	overrides, err := s.repo.SlotOverrides(ctx, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	overrideBySlot := make(map[int64]bool, len(overrides))
	for _, override := range overrides {
		overrideBySlot[slotKey(override.SlotStart)] = override.IsAvailable
	}

	var free []time.Time
	for _, slot := range candidates {
		key := slotKey(slot)
		if available, overridden := overrideBySlot[key]; overridden && !available {
			continue
		}
		slotEnd := slot.Add(time.Duration(s.cfg.SlotDurationMinutes) * time.Minute)
		isBooked := false
		for _, interval := range booked {
			if intervalsOverlap(slot, slotEnd, interval.SlotStart, interval.SlotEnd) {
				isBooked = true
				break
			}
		}
		if isBooked {
			continue
		}
		free = append(free, slot)
	}
	return free, nil
}

func (s *Service) isSlotFree(ctx context.Context, slotStart time.Time) (bool, error) {
	enabled, err := s.repo.IsSlotEnabled(ctx, slotStart)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}
	slotEnd := slotStart.Add(time.Duration(s.cfg.SlotDurationMinutes) * time.Minute)
	booked, err := s.repo.HasConfirmedOverlap(ctx, slotStart, slotEnd)
	if err != nil {
		return false, err
	}
	return !booked, nil
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

func invalidCalendarError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidCalendar, fmt.Sprintf(format, args...))
}

func slotKey(slot time.Time) int64 {
	return slot.UTC().Unix()
}

func (s *Service) sendBookingEmail(ctx context.Context, req BookRequest, code string, slotStart time.Time, calendarLink string) {
	if s.email == nil {
		if s.log != nil {
			s.log.ErrorContext(ctx, "consultation_email_provider_nil", "confirmation_code", code)
		}
		return
	}

	local := slotStart.In(s.cfg.Timezone)
	when := local.Format("Monday, 2 January 2006 at 3:04 PM MST")
	dayLabel := local.Format("Monday 2 January")
	timeLabel := local.Format("3:04 pm")
	meta := map[string]string{"confirmation_code": code, "source": req.Source}

	// Prospect confirmation first — this is the user-facing booking email.
	if email := strings.TrimSpace(req.ProspectEmail); email != "" {
		prospect := buildProspectConfirmationEmail(
			req.ProspectName,
			email,
			req.ProspectPhone,
			code,
			when,
			dayLabel,
			timeLabel,
			calendarLink,
		)
		result, err := s.email.Send(ctx, emailprovider.SendRequest{
			To:       email,
			Subject:  prospect.Subject,
			TextBody: prospect.TextBody,
			HTMLBody: prospect.HTMLBody,
			Metadata: meta,
		})
		if err != nil {
			if s.log != nil {
				s.log.ErrorContext(ctx, "consultation_confirmation_email_failed",
					"error", err, "confirmation_code", code, "to", email)
			}
		} else if s.log != nil {
			if result.Skipped {
				s.log.WarnContext(ctx, "consultation_confirmation_email_skipped",
					"confirmation_code", code, "to", email)
			} else {
				s.log.InfoContext(ctx, "consultation_confirmation_email_sent",
					"confirmation_code", code, "to", email, "redirected_to", result.RedirectedTo)
			}
		}
	} else if s.log != nil {
		s.log.WarnContext(ctx, "consultation_confirmation_email_skipped_no_address",
			"confirmation_code", code)
	}

	notify := buildInternalBookingNotifyEmail(
		req.ProspectName,
		req.ProspectEmail,
		req.ProspectPhone,
		code,
		when,
		req.Source,
		calendarLink,
	)
	if _, err := s.email.Send(ctx, emailprovider.SendRequest{
		To:       s.cfg.NotifyEmail,
		Subject:  notify.Subject,
		TextBody: notify.TextBody,
		HTMLBody: notify.HTMLBody,
		ReplyTo:  strings.TrimSpace(req.ProspectEmail),
		Metadata: meta,
	}); err != nil && s.log != nil {
		s.log.ErrorContext(ctx, "consultation_notification_email_failed",
			"error", err, "confirmation_code", code)
	}
}
