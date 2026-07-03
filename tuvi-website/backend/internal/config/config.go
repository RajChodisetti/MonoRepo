package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr string
	APIToken string

	DatabaseURL string

	NotifyEmail string

	EmailFromAddress string
	EmailFromName    string
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	SMTPUseTLS       bool
	EmailDisabled    bool

	GoogleCalendarID          string
	GoogleServiceAccountJSON  string
	GoogleCalendarDisabled    bool

	Timezone            *time.Location
	BusinessHourStart   int
	BusinessHourEnd     int
	SlotDurationMinutes int
	AvailabilityHorizon int
	DefaultAvailDays    int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:                 envOr("HTTP_ADDR", ":8090"),
		APIToken:                 strings.TrimSpace(os.Getenv("TUVI_API_TOKEN")),
		DatabaseURL:              strings.TrimSpace(os.Getenv("DATABASE_URL")),
		NotifyEmail:              envOr("NOTIFY_EMAIL", "contact@tuvisolutions.com"),
		EmailFromAddress:         envOr("EMAIL_FROM_ADDRESS", "noreply@tuvisolutions.com"),
		EmailFromName:            envOr("EMAIL_FROM_NAME", "Tuvi Solutions"),
		SMTPHost:                 strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                 envInt("SMTP_PORT", 587),
		SMTPUsername:             strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:             strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		SMTPUseTLS:               envBool("SMTP_USE_TLS", true),
		EmailDisabled:            envBool("EMAIL_DISABLED", false),
		GoogleCalendarID:         strings.TrimSpace(os.Getenv("GOOGLE_CALENDAR_ID")),
		GoogleServiceAccountJSON: strings.TrimSpace(os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON")),
		GoogleCalendarDisabled:   envBool("GOOGLE_CALENDAR_DISABLED", false),
		BusinessHourStart:        envInt("BUSINESS_HOUR_START", 9),
		BusinessHourEnd:          envInt("BUSINESS_HOUR_END", 17),
		SlotDurationMinutes:      envInt("SLOT_DURATION_MINUTES", 30),
		AvailabilityHorizon:      envInt("AVAILABILITY_HORIZON_DAYS", 14),
		DefaultAvailDays:         envInt("DEFAULT_AVAILABILITY_DAYS", 5),
	}

	loc, err := time.LoadLocation(envOr("TIMEZONE", "Australia/Sydney"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid TIMEZONE: %w", err)
	}
	cfg.Timezone = loc

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.APIToken == "" {
		return Config{}, fmt.Errorf("TUVI_API_TOKEN is required")
	}
	if !cfg.GoogleCalendarDisabled {
		if cfg.GoogleCalendarID == "" || cfg.GoogleServiceAccountJSON == "" {
			return Config{}, fmt.Errorf("GOOGLE_CALENDAR_ID and GOOGLE_SERVICE_ACCOUNT_JSON are required (or set GOOGLE_CALENDAR_DISABLED=true)")
		}
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
