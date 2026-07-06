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
	if numDays < 1 {
		numDays = 1
	}
	loc := cfg.Timezone
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

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

		for hour := cfg.BusinessHourStart; hour < cfg.BusinessHourEnd; hour++ {
			for minute := 0; minute < 60; minute += cfg.SlotDurationMinutes {
				endMinute := minute + cfg.SlotDurationMinutes
				endHour := hour
				if endMinute >= 60 {
					endHour++
					endMinute -= 60
				}
				if endHour > cfg.BusinessHourEnd || (endHour == cfg.BusinessHourEnd && endMinute > 0) {
					continue
				}
				slot := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), hour, minute, 0, 0, loc)
				if slot.After(now) {
					slots = append(slots, slot)
				}
			}
		}

		daysAdded++
		cursor = cursor.AddDate(0, 0, 1)
	}

	return slots
}

func ToSlotDTO(t time.Time, available bool) Slot {
	return Slot{
		Date:      t.Format("2006-01-02"),
		Time:      FormatSlotTime(t),
		ISO:       t.Format(time.RFC3339),
		Available: available,
	}
}

func ContainsTime(times []time.Time, target time.Time) bool {
	for _, t := range times {
		if t.Equal(target) {
			return true
		}
	}
	return false
}
