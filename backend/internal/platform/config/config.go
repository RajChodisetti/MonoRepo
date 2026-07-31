package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvLocal      = "local"
	EnvTest       = "test"
	EnvStaging    = "staging"
	EnvProduction = "production"

	providerDisabled  = "disabled"
	localDevToken     = "local-dev-token-secret-change-me-32chars"
	localDevTuviToken = "local-dev-tuvi-api-token-change-me"
)

type Config struct {
	App           AppConfig
	HTTP          HTTPConfig
	Logging       LoggingConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Email         EmailConfig
	ZohoMail      ZohoMailConfig
	Outreach      OutreachConfig
	AppURLs       AppURLsConfig
	LLM           LLMConfig
	Places        PlacesConfig
	Voice         VoiceConfig
	Storage       StorageConfig
	Token         TokenConfig
	Demo          DemoConfig
	Jobs          JobsConfig
	Consultations ConsultationConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Version string
}

type AppURLsConfig struct {
	PublicBaseURL       string
	PublicWebURL        string
	PublicMarketingURL  string
	PresentationSiteURL string
}

type HTTPConfig struct {
	Addr               string
	CORSAllowedOrigins []string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type DatabaseConfig struct {
	URL                 string
	MaxConns            int32
	MinConns            int32
	MaxConnLifetime     time.Duration
	MaxConnIdleTime     time.Duration
	ConnectTimeout      time.Duration
	RequireInProduction bool
}

type RedisConfig struct {
	URL                 string
	RequireInProduction bool
}

type EmailConfig struct {
	Provider            string
	APIKey              string
	APIBaseURL          string
	FromAddress         string
	FromName            string
	DisableSending      bool
	RedirectTo          string
	OpenTrackingEnabled bool
	RequireHTTPSLinks   bool
	AllowedLinkHosts    []string
}

type ZohoMailConfig struct {
	AccountKey   string
	AccountID    string
	FromEmail    string
	Region       string
	APIBaseURL   string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type GmailMailConfig struct {
	AccountKey   string
	MailboxEmail string
	FromEmail    string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type OutreachConfig struct {
	BulkMax                     int
	EmailsPerAccount            int
	SendWindow                  time.Duration
	SendJitterMin               time.Duration
	SendJitterMax               time.Duration
	AccountCooldown             time.Duration
	ZohoAccounts                []ZohoMailConfig
	ZohoAccountsJSON            string
	GoogleWorkspaceAccounts     []GmailMailConfig
	GoogleWorkspaceAccountsJSON string
}

type LLMConfig struct {
	Provider string
	APIKey   string
	Model    string
}

type PlacesConfig struct {
	APIKey     string
	BaseURL    string
	RegionCode string
}

type VoiceConfig struct {
	Provider      string
	WebhookSecret string
}

type StorageConfig struct {
	Provider        string
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

type TokenConfig struct {
	Secret         string
	AccessTokenTTL time.Duration
}

type DemoConfig struct {
	TokenTTL time.Duration
}

type JobsConfig struct {
	BufferSize int
	RetryDelay time.Duration
}

type ConsultationConfig struct {
	APIToken string

	NotifyEmail string

	Timezone            *time.Location
	BusinessHourStart   int
	BusinessHourEnd     int
	SlotDurationMinutes int
	DefaultAvailDays    int
	AvailabilityHorizon int

	GoogleCalendarID         string
	GoogleServiceAccountJSON string
	GoogleCalendarDisabled   bool
}

func Load() (Config, error) {
	loadEnvFiles()

	parser := &envParser{}
	cfg := Config{
		App: AppConfig{
			Name:    parser.string("APP_NAME", "restaurant-platform"),
			Env:     parser.string("APP_ENV", EnvLocal),
			Version: parser.string("APP_VERSION", "dev"),
		},
		AppURLs: AppURLsConfig{
			PublicBaseURL:       parser.string("PUBLIC_BASE_URL", "http://localhost:8080"),
			PublicWebURL:        parser.string("PUBLIC_WEB_URL", "http://localhost:3000"),
			PublicMarketingURL:  parser.string("PUBLIC_MARKETING_URL", "http://localhost:3001"),
			PresentationSiteURL: parser.string("PRESENTATION_SITE_URL", "http://localhost:5500"),
		},
		HTTP: HTTPConfig{
			Addr:               parser.listenAddr(),
			CORSAllowedOrigins: parser.csv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:3001", "http://127.0.0.1:3001"}),
		},
		Logging: LoggingConfig{
			Level:  parser.string("LOG_LEVEL", "info"),
			Format: parser.string("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			URL:                 parser.string("DATABASE_URL", ""),
			MaxConns:            int32(parser.int("DATABASE_MAX_CONNS", 5)),
			MinConns:            int32(parser.int("DATABASE_MIN_CONNS", 0)),
			MaxConnLifetime:     parser.duration("DATABASE_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:     parser.duration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:      parser.duration("DATABASE_CONNECT_TIMEOUT", 5*time.Second),
			RequireInProduction: true,
		},
		Redis: RedisConfig{
			URL:                 parser.string("REDIS_URL", ""),
			RequireInProduction: true,
		},
		Email: EmailConfig{
			Provider:            parser.string("EMAIL_PROVIDER", providerDisabled),
			APIKey:              parser.string("EMAIL_API_KEY", ""),
			APIBaseURL:          parser.string("EMAIL_API_BASE_URL", "https://api.resend.com"),
			FromAddress:         parser.string("EMAIL_FROM_ADDRESS", ""),
			FromName:            parser.string("EMAIL_FROM_NAME", "Tuvi Solutions"),
			DisableSending:      parser.bool("EMAIL_DISABLE_SENDING", true),
			RedirectTo:          parser.string("EMAIL_REDIRECT_TO", ""),
			OpenTrackingEnabled: parser.bool("EMAIL_OPEN_TRACKING_ENABLED", true),
		},
		ZohoMail: ZohoMailConfig{
			AccountID:    parser.string("ZOHO_ACCOUNT_ID", ""),
			FromEmail:    parser.string("ZOHO_FROM_EMAIL", ""),
			Region:       parser.string("ZOHO_REGION", "com"),
			APIBaseURL:   parser.string("ZOHO_API_BASE_URL", "https://mail.zoho.com/api/accounts"),
			ClientID:     parser.string("ZOHO_CLIENT_ID", ""),
			ClientSecret: parser.string("ZOHO_CLIENT_SECRET", ""),
			RefreshToken: parser.string("ZOHO_REFRESH_TOKEN", ""),
		},
		Outreach: loadOutreachConfig(parser),
		LLM: LLMConfig{
			Provider: parser.string("LLM_PROVIDER", providerDisabled),
			APIKey:   parser.string("LLM_API_KEY", ""),
			Model:    parser.string("LLM_MODEL", ""),
		},
		Places: loadPlacesConfig(parser),
		Voice: VoiceConfig{
			Provider:      parser.string("VOICE_PROVIDER", providerDisabled),
			WebhookSecret: parser.string("VOICE_WEBHOOK_SECRET", ""),
		},
		Storage: StorageConfig{
			Provider:        parser.string("STORAGE_PROVIDER", providerDisabled),
			Bucket:          parser.string("STORAGE_BUCKET", ""),
			Region:          parser.string("STORAGE_REGION", ""),
			Endpoint:        parser.string("STORAGE_ENDPOINT", ""),
			AccessKeyID:     parser.string("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: parser.string("STORAGE_SECRET_ACCESS_KEY", ""),
		},
		Token: TokenConfig{
			Secret:         parser.string("TOKEN_SECRET", localDevToken),
			AccessTokenTTL: parser.duration("JWT_ACCESS_TOKEN_TTL", 24*time.Hour),
		},
		Demo: DemoConfig{
			TokenTTL: parser.duration("DEMO_TOKEN_TTL", 30*24*time.Hour),
		},
		Jobs: JobsConfig{
			BufferSize: parser.int("JOB_BUFFER_SIZE", 32),
			RetryDelay: parser.duration("JOB_RETRY_DELAY", 2*time.Second),
		},
		Consultations: ConsultationConfig{
			APIToken:                 parser.string("TUVI_API_TOKEN", localDevTuviToken),
			NotifyEmail:              parser.string("CONSULTATION_NOTIFY_EMAIL", "contact@tuvisolutions.com"),
			Timezone:                 parser.location("CONSULTATION_TIMEZONE", "Australia/Sydney"),
			BusinessHourStart:        parser.int("CONSULTATION_BUSINESS_HOUR_START", 9),
			BusinessHourEnd:          parser.int("CONSULTATION_BUSINESS_HOUR_END", 17),
			SlotDurationMinutes:      parser.int("CONSULTATION_SLOT_DURATION_MINUTES", 30),
			DefaultAvailDays:         parser.int("CONSULTATION_DEFAULT_AVAILABILITY_DAYS", 5),
			AvailabilityHorizon:      parser.int("CONSULTATION_AVAILABILITY_HORIZON_DAYS", 14),
			GoogleCalendarID:         parser.string("CONSULTATION_GOOGLE_CALENDAR_ID", ""),
			GoogleServiceAccountJSON: parser.string("CONSULTATION_GOOGLE_SERVICE_ACCOUNT_JSON", ""),
			GoogleCalendarDisabled:   parser.bool("CONSULTATION_GOOGLE_CALENDAR_DISABLED", true),
		},
	}
	cfg.Email.RequireHTTPSLinks = cfg.requiresExplicitSecrets()
	if cfg.App.Env == EnvProduction {
		cfg.Email.AllowedLinkHosts = []string{
			"api.tuvisolutions.com",
			"demo.tuvisolutions.com",
			"tuvisolutions.com",
		}
	}

	if err := parser.join(); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	if !oneOf(c.App.Env, EnvLocal, EnvTest, EnvStaging, EnvProduction) {
		errs = append(errs, fmt.Errorf("APP_ENV must be one of local, test, staging, production"))
	}
	if strings.TrimSpace(c.App.Name) == "" {
		errs = append(errs, fmt.Errorf("APP_NAME is required"))
	}
	if c.Token.AccessTokenTTL <= 0 {
		errs = append(errs, fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be positive"))
	}
	if strings.TrimSpace(c.HTTP.Addr) == "" {
		errs = append(errs, fmt.Errorf("HTTP_ADDR is required"))
	}
	if c.Database.MaxConns < 1 {
		errs = append(errs, fmt.Errorf("DATABASE_MAX_CONNS must be at least 1"))
	}
	if c.Database.MinConns < 0 {
		errs = append(errs, fmt.Errorf("DATABASE_MIN_CONNS cannot be negative"))
	}
	if c.Database.MinConns > c.Database.MaxConns {
		errs = append(errs, fmt.Errorf("DATABASE_MIN_CONNS cannot exceed DATABASE_MAX_CONNS"))
	}
	if c.Database.MaxConnLifetime <= 0 {
		errs = append(errs, fmt.Errorf("DATABASE_MAX_CONN_LIFETIME must be positive"))
	}
	if c.Database.MaxConnIdleTime <= 0 {
		errs = append(errs, fmt.Errorf("DATABASE_MAX_CONN_IDLE_TIME must be positive"))
	}
	if c.Database.ConnectTimeout <= 0 {
		errs = append(errs, fmt.Errorf("DATABASE_CONNECT_TIMEOUT must be positive"))
	}
	if len(c.Token.Secret) < 32 {
		errs = append(errs, fmt.Errorf("TOKEN_SECRET must be at least 32 characters"))
	}
	if c.Demo.TokenTTL <= 0 {
		errs = append(errs, fmt.Errorf("DEMO_TOKEN_TTL must be positive"))
	}
	if c.Jobs.BufferSize < 1 {
		errs = append(errs, fmt.Errorf("JOB_BUFFER_SIZE must be at least 1"))
	}
	if c.Jobs.RetryDelay <= 0 {
		errs = append(errs, fmt.Errorf("JOB_RETRY_DELAY must be positive"))
	}
	if c.Outreach.BulkMax < 1 || c.Outreach.BulkMax > 150 {
		errs = append(errs, fmt.Errorf("OUTREACH_BULK_MAX must be between 1 and 150"))
	}
	if c.Outreach.EmailsPerAccount < 1 || c.Outreach.EmailsPerAccount > 40 {
		errs = append(errs, fmt.Errorf("OUTREACH_EMAILS_PER_ACCOUNT must be between 1 and 40"))
	}
	if c.Outreach.SendWindow < 8*time.Hour {
		errs = append(errs, fmt.Errorf("OUTREACH_SEND_WINDOW must be at least 8h"))
	}
	if c.Outreach.SendJitterMin < 2*time.Minute {
		errs = append(errs, fmt.Errorf("OUTREACH_SEND_JITTER_MIN must be at least 2m"))
	}
	if c.Outreach.SendJitterMax < c.Outreach.SendJitterMin {
		errs = append(errs, fmt.Errorf("OUTREACH_SEND_JITTER_MAX must be greater than or equal to OUTREACH_SEND_JITTER_MIN"))
	}
	if c.Outreach.EmailsPerAccount > 0 && c.Outreach.SendWindow/time.Duration(c.Outreach.EmailsPerAccount) <= c.Outreach.SendJitterMax {
		errs = append(errs, fmt.Errorf("OUTREACH_SEND_WINDOW divided by OUTREACH_EMAILS_PER_ACCOUNT must be greater than OUTREACH_SEND_JITTER_MAX"))
	}
	if c.Outreach.AccountCooldown < 24*time.Hour {
		errs = append(errs, fmt.Errorf("OUTREACH_EMAIL_COOLDOWN must be at least 24h"))
	}
	if c.App.Env == EnvProduction && strings.TrimSpace(c.Email.RedirectTo) != "" {
		errs = append(errs, fmt.Errorf("EMAIL_REDIRECT_TO must be empty in production"))
	}
	if strings.ContainsAny(c.Email.FromName, "\r\n") {
		errs = append(errs, fmt.Errorf("EMAIL_FROM_NAME must not contain newlines"))
	}
	if redirect := strings.TrimSpace(c.Email.RedirectTo); redirect != "" {
		if _, err := canonicalOutreachMailbox(redirect); err != nil {
			errs = append(errs, fmt.Errorf("EMAIL_REDIRECT_TO must be a single valid email address"))
		}
	}
	if strings.TrimSpace(c.Consultations.APIToken) == "" {
		errs = append(errs, fmt.Errorf("TUVI_API_TOKEN is required"))
	}
	if strings.TrimSpace(c.Consultations.NotifyEmail) == "" {
		errs = append(errs, fmt.Errorf("CONSULTATION_NOTIFY_EMAIL is required"))
	}
	if c.Consultations.Timezone == nil {
		errs = append(errs, fmt.Errorf("CONSULTATION_TIMEZONE is required"))
	}
	if c.Consultations.BusinessHourStart < 0 || c.Consultations.BusinessHourStart > 23 {
		errs = append(errs, fmt.Errorf("CONSULTATION_BUSINESS_HOUR_START must be between 0 and 23"))
	}
	if c.Consultations.BusinessHourEnd < 1 || c.Consultations.BusinessHourEnd > 24 {
		errs = append(errs, fmt.Errorf("CONSULTATION_BUSINESS_HOUR_END must be between 1 and 24"))
	}
	if c.Consultations.BusinessHourStart >= c.Consultations.BusinessHourEnd {
		errs = append(errs, fmt.Errorf("CONSULTATION_BUSINESS_HOUR_START must be before CONSULTATION_BUSINESS_HOUR_END"))
	}
	if c.Consultations.SlotDurationMinutes < 1 || c.Consultations.SlotDurationMinutes > 240 {
		errs = append(errs, fmt.Errorf("CONSULTATION_SLOT_DURATION_MINUTES must be between 1 and 240"))
	}
	if c.Consultations.DefaultAvailDays < 1 {
		errs = append(errs, fmt.Errorf("CONSULTATION_DEFAULT_AVAILABILITY_DAYS must be at least 1"))
	}
	if c.Consultations.AvailabilityHorizon < c.Consultations.DefaultAvailDays {
		errs = append(errs, fmt.Errorf("CONSULTATION_AVAILABILITY_HORIZON_DAYS must be greater than or equal to CONSULTATION_DEFAULT_AVAILABILITY_DAYS"))
	}
	if !c.Consultations.GoogleCalendarDisabled {
		if strings.TrimSpace(c.Consultations.GoogleCalendarID) == "" {
			errs = append(errs, fmt.Errorf("CONSULTATION_GOOGLE_CALENDAR_ID is required when consultation Google Calendar is enabled"))
		}
		if strings.TrimSpace(c.Consultations.GoogleServiceAccountJSON) == "" {
			errs = append(errs, fmt.Errorf("CONSULTATION_GOOGLE_SERVICE_ACCOUNT_JSON is required when consultation Google Calendar is enabled"))
		}
	}
	if !oneOf(c.Logging.Level, "debug", "info", "warn", "error") {
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error"))
	}
	if !oneOf(c.Logging.Format, "json", "text") {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be json or text"))
	}

	errs = append(errs, c.validateProviders()...)
	errs = append(errs, c.validateEmailURLs()...)
	errs = append(errs, c.validateDeployedSecrets()...)

	return errors.Join(errs...)
}

func (c Config) validateEmailURLs() []error {
	if c.Email.DisableSending {
		return nil
	}

	values := []struct {
		name  string
		value string
	}{
		{name: "PUBLIC_BASE_URL", value: c.AppURLs.PublicBaseURL},
		{name: "PUBLIC_WEB_URL", value: c.AppURLs.PublicWebURL},
		{name: "PUBLIC_MARKETING_URL", value: c.AppURLs.PublicMarketingURL},
		{name: "PRESENTATION_SITE_URL", value: c.AppURLs.PresentationSiteURL},
	}

	var errs []error
	for _, item := range values {
		parsed, err := url.Parse(strings.TrimSpace(item.value))
		if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" || parsed.User != nil {
			errs = append(errs, fmt.Errorf("%s must be an absolute public URL", item.name))
			continue
		}
		if !c.Email.RequireHTTPSLinks {
			continue
		}
		if !strings.EqualFold(parsed.Scheme, "https") {
			errs = append(errs, fmt.Errorf("%s must use https in %s", item.name, c.App.Env))
			continue
		}
		if isLocalHostname(parsed.Hostname()) {
			errs = append(errs, fmt.Errorf("%s must not use a loopback or local hostname in %s", item.name, c.App.Env))
		}
	}
	if c.App.Env == EnvProduction {
		expectedHosts := []struct {
			name string
			url  string
			host string
		}{
			{name: "PUBLIC_BASE_URL", url: c.AppURLs.PublicBaseURL, host: "api.tuvisolutions.com"},
			{name: "PUBLIC_WEB_URL", url: c.AppURLs.PublicWebURL, host: "demo.tuvisolutions.com"},
			{name: "PUBLIC_MARKETING_URL", url: c.AppURLs.PublicMarketingURL, host: "tuvisolutions.com"},
			{name: "PRESENTATION_SITE_URL", url: c.AppURLs.PresentationSiteURL, host: "tuvisolutions.com"},
		}
		for _, expected := range expectedHosts {
			parsed, err := url.Parse(strings.TrimSpace(expected.url))
			if err != nil {
				continue
			}
			if !strings.EqualFold(parsed.Hostname(), expected.host) {
				errs = append(errs, fmt.Errorf("%s must use %s in production", expected.name, expected.host))
			}
			if parsed.Port() != "" {
				errs = append(errs, fmt.Errorf("%s must not use a custom port in production", expected.name))
			}
		}
		productionPaths := []struct {
			name string
			url  string
			path string
		}{
			{name: "PUBLIC_BASE_URL", url: c.AppURLs.PublicBaseURL, path: ""},
			{name: "PUBLIC_WEB_URL", url: c.AppURLs.PublicWebURL, path: ""},
			{name: "PUBLIC_MARKETING_URL", url: c.AppURLs.PublicMarketingURL, path: ""},
			{name: "PRESENTATION_SITE_URL", url: c.AppURLs.PresentationSiteURL, path: "/services/restaurants"},
		}
		for _, expected := range productionPaths {
			parsed, err := url.Parse(strings.TrimSpace(expected.url))
			if err != nil {
				continue
			}
			path := strings.TrimSuffix(parsed.EscapedPath(), "/")
			if path != expected.path || parsed.RawQuery != "" || parsed.Fragment != "" {
				errs = append(errs, fmt.Errorf("%s must use the canonical production path %q without query or fragment", expected.name, expected.path))
			}
		}
	}
	return errs
}

func isLocalHostname(host string) bool {
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}

func (c Config) validateProviders() []error {
	var errs []error

	if providerEnabled(c.Email.Provider) {
		if strings.TrimSpace(c.Email.FromAddress) == "" {
			errs = append(errs, fmt.Errorf("EMAIL_FROM_ADDRESS is required when EMAIL_PROVIDER is enabled"))
		}
		switch strings.ToLower(strings.TrimSpace(c.Email.Provider)) {
		case "smtp":
			errs = append(errs, fmt.Errorf("EMAIL_PROVIDER=smtp is no longer supported; use EMAIL_PROVIDER=resend with EMAIL_API_KEY"))
		case "resend", "http", "https":
			if strings.TrimSpace(c.Email.APIKey) == "" {
				errs = append(errs, fmt.Errorf("EMAIL_API_KEY is required when EMAIL_PROVIDER is %s", c.Email.Provider))
			}
			if strings.TrimSpace(c.Email.APIBaseURL) == "" {
				errs = append(errs, fmt.Errorf("EMAIL_API_BASE_URL is required when EMAIL_PROVIDER is enabled"))
			}
		case "zoho":
			if strings.TrimSpace(c.ZohoMail.AccountID) == "" {
				errs = append(errs, fmt.Errorf("ZOHO_ACCOUNT_ID is required when EMAIL_PROVIDER is zoho"))
			}
			if strings.TrimSpace(c.ZohoMail.ClientID) == "" || strings.TrimSpace(c.ZohoMail.ClientSecret) == "" || strings.TrimSpace(c.ZohoMail.RefreshToken) == "" {
				errs = append(errs, fmt.Errorf("ZOHO_CLIENT_ID, ZOHO_CLIENT_SECRET, and ZOHO_REFRESH_TOKEN are required when EMAIL_PROVIDER is zoho"))
			}
			if strings.TrimSpace(c.ZohoMail.FromEmail) == "" && strings.TrimSpace(c.Email.FromAddress) == "" {
				errs = append(errs, fmt.Errorf("ZOHO_FROM_EMAIL or EMAIL_FROM_ADDRESS is required when EMAIL_PROVIDER is zoho"))
			}
		default:
			if strings.TrimSpace(c.Email.APIKey) == "" {
				errs = append(errs, fmt.Errorf("EMAIL_API_KEY is required when EMAIL_PROVIDER is enabled"))
			}
		}
	}

	if providerEnabled(c.LLM.Provider) {
		if strings.TrimSpace(c.LLM.APIKey) == "" {
			errs = append(errs, fmt.Errorf("LLM_API_KEY is required when LLM_PROVIDER is enabled"))
		}
		if strings.TrimSpace(c.LLM.Model) == "" {
			errs = append(errs, fmt.Errorf("LLM_MODEL is required when LLM_PROVIDER is enabled"))
		}
	}

	if providerEnabled(c.Voice.Provider) {
		if strings.TrimSpace(c.Voice.WebhookSecret) == "" {
			errs = append(errs, fmt.Errorf("VOICE_WEBHOOK_SECRET is required when VOICE_PROVIDER is enabled"))
		}
	}

	if providerEnabled(c.Storage.Provider) {
		if strings.TrimSpace(c.Storage.Bucket) == "" {
			errs = append(errs, fmt.Errorf("STORAGE_BUCKET is required when STORAGE_PROVIDER is enabled"))
		}
		if strings.TrimSpace(c.Storage.AccessKeyID) == "" {
			errs = append(errs, fmt.Errorf("STORAGE_ACCESS_KEY_ID is required when STORAGE_PROVIDER is enabled"))
		}
		if strings.TrimSpace(c.Storage.SecretAccessKey) == "" {
			errs = append(errs, fmt.Errorf("STORAGE_SECRET_ACCESS_KEY is required when STORAGE_PROVIDER is enabled"))
		}
		if strings.TrimSpace(c.Storage.Region) == "" && strings.TrimSpace(c.Storage.Endpoint) == "" {
			errs = append(errs, fmt.Errorf("STORAGE_REGION or STORAGE_ENDPOINT is required when STORAGE_PROVIDER is enabled"))
		}
	}

	return errs
}

func (c Config) validateDeployedSecrets() []error {
	if !c.requiresExplicitSecrets() {
		return nil
	}

	var errs []error

	if strings.TrimSpace(c.Database.URL) == "" && c.Database.RequireInProduction {
		errs = append(errs, fmt.Errorf("DATABASE_URL is required in %s", c.App.Env))
	}
	if c.Token.Secret == localDevToken {
		errs = append(errs, fmt.Errorf("TOKEN_SECRET must be explicit in %s", c.App.Env))
	}
	if c.Consultations.APIToken == localDevTuviToken {
		errs = append(errs, fmt.Errorf("TUVI_API_TOKEN must be explicit in %s", c.App.Env))
	}
	if strings.TrimSpace(c.Redis.URL) == "" && c.Redis.RequireInProduction {
		errs = append(errs, fmt.Errorf("REDIS_URL is required in %s", c.App.Env))
	}

	return errs
}

func (c Config) requiresExplicitSecrets() bool {
	return c.App.Env == EnvProduction || c.App.Env == EnvStaging
}

func providerEnabled(provider string) bool {
	provider = strings.TrimSpace(provider)
	return provider != "" && provider != providerDisabled
}

func loadPlacesConfig(parser *envParser) PlacesConfig {
	apiKey := strings.TrimSpace(parser.string("GOOGLE_PLACES_API_KEY", ""))
	if apiKey == "" {
		apiKey = strings.TrimSpace(parser.string("PLACES_API", ""))
	}
	return PlacesConfig{
		APIKey:     apiKey,
		BaseURL:    strings.TrimRight(parser.string("PLACES_API_BASE_URL", "https://places.googleapis.com/v1"), "/"),
		RegionCode: strings.ToUpper(parser.string("PLACES_REGION_CODE", "AU")),
	}
}

func loadEnvFiles() {
	for _, path := range []string{".env", "backend/.env", "../.env"} {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		_ = godotenv.Load(path)
	}
}

func normalizeListenAddr(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ":8080"
	}
	if strings.HasPrefix(value, ":") || strings.Contains(value, ":") {
		return value
	}
	return ":" + value
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
