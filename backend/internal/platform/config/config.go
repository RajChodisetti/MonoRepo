package config

import (
	"errors"
	"fmt"
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

	providerDisabled = "disabled"
	localDevToken    = "local-dev-token-secret-change-me-32chars"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Logging  LoggingConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Email    EmailConfig
	SMTP     SMTPConfig
	AppURLs  AppURLsConfig
	LLM      LLMConfig
	Voice    VoiceConfig
	Storage  StorageConfig
	Token    TokenConfig
	Demo     DemoConfig
	Jobs     JobsConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Version string
}

type AppURLsConfig struct {
	PublicBaseURL string
	PublicWebURL  string
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
	Provider          string
	APIKey            string
	FromAddress       string
	FromName          string
	DisableSending    bool
	RedirectTo        string
	OpenTrackingEnabled bool
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

type LLMConfig struct {
	Provider string
	APIKey   string
	Model    string
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
	TokenSecret string
	TokenTTL    time.Duration
}

type JobsConfig struct {
	BufferSize int
	RetryDelay time.Duration
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
			PublicBaseURL: parser.string("PUBLIC_BASE_URL", "http://localhost:8080"),
			PublicWebURL:  parser.string("PUBLIC_WEB_URL", "http://localhost:3000"),
		},
		HTTP: HTTPConfig{
			Addr:               parser.listenAddr(),
			CORSAllowedOrigins: parser.csv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
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
			FromAddress:         parser.string("EMAIL_FROM_ADDRESS", ""),
			FromName:            parser.string("EMAIL_FROM_NAME", "Tuvi Solutions"),
			DisableSending:      parser.bool("EMAIL_DISABLE_SENDING", true),
			RedirectTo:          parser.string("EMAIL_REDIRECT_TO", ""),
			OpenTrackingEnabled: parser.bool("EMAIL_OPEN_TRACKING_ENABLED", true),
		},
		SMTP: SMTPConfig{
			Host:     parser.string("SMTP_HOST", "smtp.gmail.com"),
			Port:     parser.int("SMTP_PORT", 587),
			Username: parser.string("SMTP_USERNAME", ""),
			Password: parser.string("SMTP_PASSWORD", ""),
			UseTLS:   parser.bool("SMTP_USE_TLS", true),
		},
		LLM: LLMConfig{
			Provider: parser.string("LLM_PROVIDER", providerDisabled),
			APIKey:   parser.string("LLM_API_KEY", ""),
			Model:    parser.string("LLM_MODEL", ""),
		},
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
			TokenSecret: parser.string("DEMO_TOKEN_SECRET", localDevToken),
			TokenTTL:    parser.duration("DEMO_TOKEN_TTL", 30*24*time.Hour),
		},
		Jobs: JobsConfig{
			BufferSize: parser.int("JOB_BUFFER_SIZE", 32),
			RetryDelay: parser.duration("JOB_RETRY_DELAY", 2*time.Second),
		},
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
	if len(c.Demo.TokenSecret) < 32 {
		errs = append(errs, fmt.Errorf("DEMO_TOKEN_SECRET must be at least 32 characters"))
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
	if !oneOf(c.Logging.Level, "debug", "info", "warn", "error") {
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error"))
	}
	if !oneOf(c.Logging.Format, "json", "text") {
		errs = append(errs, fmt.Errorf("LOG_FORMAT must be json or text"))
	}

	errs = append(errs, c.validateProviders()...)
	errs = append(errs, c.validateDeployedSecrets()...)

	return errors.Join(errs...)
}

func (c Config) validateProviders() []error {
	var errs []error

	if providerEnabled(c.Email.Provider) {
		if strings.TrimSpace(c.Email.FromAddress) == "" {
			errs = append(errs, fmt.Errorf("EMAIL_FROM_ADDRESS is required when EMAIL_PROVIDER is enabled"))
		}
		switch strings.ToLower(strings.TrimSpace(c.Email.Provider)) {
		case "smtp":
			if strings.TrimSpace(c.SMTP.Host) == "" {
				errs = append(errs, fmt.Errorf("SMTP_HOST is required when EMAIL_PROVIDER is smtp"))
			}
			if c.SMTP.Port < 1 || c.SMTP.Port > 65535 {
				errs = append(errs, fmt.Errorf("SMTP_PORT must be between 1 and 65535"))
			}
			if strings.TrimSpace(c.SMTP.Username) == "" {
				errs = append(errs, fmt.Errorf("SMTP_USERNAME is required when EMAIL_PROVIDER is smtp"))
			}
			if strings.TrimSpace(c.SMTP.Password) == "" {
				errs = append(errs, fmt.Errorf("SMTP_PASSWORD is required when EMAIL_PROVIDER is smtp"))
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
