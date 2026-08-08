package consultations

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	time24Re = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	time12Re = regexp.MustCompile(`(?i)^(\d{1,2})(?::(\d{2}))?\s*(am|pm)$`)
)

type SlotConfig struct {
	Timezone            *time.Location
	BusinessHourStart   int
	BusinessHourEnd     int
	SlotDurationMinutes int
	HorizonDays         int
}

type Slot struct {
	Date      string `json:"date"`
	Time      string `json:"time"`
	ISO       string `json:"iso"`
	Available bool   `json:"available"`
}

func ParseDate(dateStr string, loc *time.Location) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required (YYYY-MM-DD)")
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return t, nil
}

func ParseTime(timeStr string) (hour, minute int, err error) {
	timeStr = strings.TrimSpace(strings.ToLower(timeStr))
	if timeStr == "" {
		return 0, 0, fmt.Errorf("time is required")
	}

	if m := time24Re.FindStringSubmatch(timeStr); len(m) == 3 {
		h, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		if h < 0 || h > 23 || min < 0 || min > 59 {
			return 0, 0, fmt.Errorf("invalid time")
		}
		return h, min, nil
	}

	if m := time12Re.FindStringSubmatch(timeStr); len(m) == 4 {
		h, _ := strconv.Atoi(m[1])
		min := 0
		if m[2] != "" {
			min, _ = strconv.Atoi(m[2])
		}
		ampm := strings.ToLower(m[3])
		if h < 1 || h > 12 || min < 0 || min > 59 {
			return 0, 0, fmt.Errorf("invalid time")
		}
		if ampm == "pm" && h != 12 {
			h += 12
		}
		if ampm == "am" && h == 12 {
			h = 0
		}
		return h, min, nil
	}

	return 0, 0, fmt.Errorf("invalid time format, use HH:MM or e.g. 2:00 PM")
}

func BuildSlotStart(date time.Time, timeStr string, loc *time.Location) (time.Time, error) {
	hour, minute, err := ParseTime(timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, loc), nil
}

func FormatSlotTime(t time.Time) string {
	return t.Format("15:04")
}

func GenerateConfirmationCode() (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "TVI-" + strings.ToUpper(hex.EncodeToString(buf)), nil
}

func GenerateCandidateSlots(cfg SlotConfig, startDate time.Time, numDays int) []time.Time {
	return generateCandidateSlotsAt(cfg, startDate, numDays, time.Now().In(cfg.Timezone))
}

func generateCandidateSlotsAt(cfg SlotConfig, startDate time.Time, numDays int, now time.Time) []time.Time {
	if numDays < 1 {
		numDays = 1
	}
	loc := cfg.Timezone
	now = now.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	localStart := startDate.In(loc)
	startDate = time.Date(localStart.Year(), localStart.Month(), localStart.Day(), 0, 0, 0, 0, loc)

	if startDate.Before(today) {
		startDate = today
	}

	var slots []time.Time
	cursor := startDate
	daysAdded := 0

	for daysAdded < numDays {
		if cursor.Weekday() == time.Saturday || cursor.Weekday() == time.Sunday {
			cursor = cursor.AddDate(0, 0, 1)
			continue
		}

		for _, slot := range GenerateDayCandidateSlots(cfg, cursor) {
			if slot.After(now) {
				slots = append(slots, slot)
			}
		}

		daysAdded++
		cursor = cursor.AddDate(0, 0, 1)
	}

	return slots
}

func GenerateDayCandidateSlots(cfg SlotConfig, date time.Time) []time.Time {
	loc := cfg.Timezone
	localDate := date.In(loc)
	if localDate.Weekday() == time.Saturday || localDate.Weekday() == time.Sunday {
		return nil
	}
	if cfg.SlotDurationMinutes < 1 {
		return nil
	}

	dayStart := time.Date(
		localDate.Year(),
		localDate.Month(),
		localDate.Day(),
		cfg.BusinessHourStart,
		0,
		0,
		0,
		loc,
	)
	dayEnd := time.Date(
		localDate.Year(),
		localDate.Month(),
		localDate.Day(),
		cfg.BusinessHourEnd,
		0,
		0,
		0,
		loc,
	)
	duration := time.Duration(cfg.SlotDurationMinutes) * time.Minute

	var slots []time.Time
	for slot := dayStart; !slot.Add(duration).After(dayEnd); slot = slot.Add(duration) {
		slots = append(slots, slot)
	}
	return slots
}

func GenerateMonthCandidateSlots(cfg SlotConfig, monthStart time.Time) []time.Time {
	loc := cfg.Timezone
	localMonth := monthStart.In(loc)
	cursor := time.Date(localMonth.Year(), localMonth.Month(), 1, 0, 0, 0, 0, loc)
	monthEnd := cursor.AddDate(0, 1, 0)

	var slots []time.Time
	for cursor.Before(monthEnd) {
		slots = append(slots, GenerateDayCandidateSlots(cfg, cursor)...)
		cursor = cursor.AddDate(0, 0, 1)
	}
	return slots
}

func IsCandidateSlot(cfg SlotConfig, candidate time.Time) bool {
	local := candidate.In(cfg.Timezone)
	if local.Second() != 0 || local.Nanosecond() != 0 {
		return false
	}
	for _, slot := range GenerateDayCandidateSlots(cfg, local) {
		if slot.Equal(candidate) {
			return true
		}
	}
	return false
}

func ParseMonth(month string, loc *time.Location) (time.Time, error) {
	month = strings.TrimSpace(month)
	parsed, err := time.ParseInLocation("2006-01", month, loc)
	if err != nil || parsed.Format("2006-01") != month {
		return time.Time{}, fmt.Errorf("invalid month format, use YYYY-MM")
	}
	return parsed, nil
}

func ToSlotDTO(t time.Time, available bool) Slot {
	return Slot{
		Date:      t.Format("2006-01-02"),
		Time:      FormatSlotTime(t),
		ISO:       t.Format(time.RFC3339),
		Available: available,
	}
}
