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

	EmailProvider    string
	EmailFromAddress string
	EmailFromName    string
	EmailDisabled    bool

	ZohoAccountID    string
	ZohoFromEmail    string
	ZohoRegion       string
	ZohoAPIBaseURL   string
	ZohoClientID     string
	ZohoClientSecret string
	ZohoRefreshToken string

	ResendAPIKey     string
	ResendAPIBaseURL string

	GoogleCalendarID         string
	GoogleServiceAccountJSON string
	GoogleCalendarDisabled   bool

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
		EmailProvider:            envOr("EMAIL_PROVIDER", "zoho"),
		EmailFromAddress:         envOr("EMAIL_FROM_ADDRESS", "noreply@tuvisolutions.com"),
		EmailFromName:            envOr("EMAIL_FROM_NAME", "Tuvi Solutions"),
		EmailDisabled:            envBool("EMAIL_DISABLED", false),
		ZohoAccountID:            strings.TrimSpace(os.Getenv("ZOHO_ACCOUNT_ID")),
		ZohoFromEmail:            strings.TrimSpace(os.Getenv("ZOHO_FROM_EMAIL")),
		ZohoRegion:               envOr("ZOHO_REGION", "com"),
		ZohoAPIBaseURL:           envOr("ZOHO_API_BASE_URL", "https://mail.zoho.com/api/accounts"),
		ZohoClientID:             strings.TrimSpace(os.Getenv("ZOHO_CLIENT_ID")),
		ZohoClientSecret:         strings.TrimSpace(os.Getenv("ZOHO_CLIENT_SECRET")),
		ZohoRefreshToken:         strings.TrimSpace(os.Getenv("ZOHO_REFRESH_TOKEN")),
		ResendAPIKey:             strings.TrimSpace(os.Getenv("RESEND_API_KEY")),
		ResendAPIBaseURL:         envOr("RESEND_API_BASE_URL", "https://api.resend.com"),
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
	if !cfg.EmailDisabled {
		switch strings.ToLower(strings.TrimSpace(cfg.EmailProvider)) {
		case "smtp":
			return Config{}, fmt.Errorf("EMAIL_PROVIDER=smtp is no longer supported; use zoho or resend")
		case "zoho", "http", "https":
			if cfg.ZohoAccountID == "" || cfg.ZohoClientID == "" || cfg.ZohoClientSecret == "" || cfg.ZohoRefreshToken == "" {
				return Config{}, fmt.Errorf("ZOHO_ACCOUNT_ID, ZOHO_CLIENT_ID, ZOHO_CLIENT_SECRET, and ZOHO_REFRESH_TOKEN are required when EMAIL_PROVIDER=zoho")
			}
			if cfg.ZohoFromEmail == "" && cfg.EmailFromAddress == "" {
				return Config{}, fmt.Errorf("ZOHO_FROM_EMAIL or EMAIL_FROM_ADDRESS is required when EMAIL_PROVIDER=zoho")
			}
		case "resend":
			if cfg.ResendAPIKey == "" {
				return Config{}, fmt.Errorf("RESEND_API_KEY is required when EMAIL_PROVIDER=resend")
			}
		default:
			return Config{}, fmt.Errorf("unsupported EMAIL_PROVIDER %q (use zoho or resend)", cfg.EmailProvider)
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
