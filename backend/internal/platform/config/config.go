package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvLocal      = "local"
	EnvTest       = "test"
	EnvStaging    = "staging"
	EnvProduction = "production"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Logging  LoggingConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Email    EmailConfig
	LLM      LLMConfig
	Voice    VoiceConfig
	Storage  StorageConfig
	Token    TokenConfig
	Jobs     JobsConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Version string
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
	URL string
}

type EmailConfig struct {
	Provider       string
	APIKey         string
	FromAddress    string
	DisableSending bool
	RedirectTo     string
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
	Secret string
}

type JobsConfig struct {
	BufferSize int
	RetryDelay time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name:    envString("APP_NAME", "restaurant-platform"),
			Env:     envString("APP_ENV", EnvLocal),
			Version: envString("APP_VERSION", "dev"),
		},
		HTTP: HTTPConfig{
			Addr:               envString("HTTP_ADDR", ":8080"),
			CORSAllowedOrigins: envCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		},
		Logging: LoggingConfig{
			Level:  envString("LOG_LEVEL", "info"),
			Format: envString("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			URL:                 envString("DATABASE_URL", ""),
			MaxConns:            int32(envInt("DATABASE_MAX_CONNS", 5)),
			MinConns:            int32(envInt("DATABASE_MIN_CONNS", 0)),
			MaxConnLifetime:     envDuration("DATABASE_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:     envDuration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:      envDuration("DATABASE_CONNECT_TIMEOUT", 5*time.Second),
			RequireInProduction: true,
		},
		Redis: RedisConfig{
			URL: envString("REDIS_URL", ""),
		},
		Email: EmailConfig{
			Provider:       envString("EMAIL_PROVIDER", "disabled"),
			APIKey:         envString("EMAIL_API_KEY", ""),
			FromAddress:    envString("EMAIL_FROM_ADDRESS", ""),
			DisableSending: envBool("EMAIL_DISABLE_SENDING", true),
			RedirectTo:     envString("EMAIL_REDIRECT_TO", ""),
		},
		LLM: LLMConfig{
			Provider: envString("LLM_PROVIDER", "disabled"),
			APIKey:   envString("LLM_API_KEY", ""),
			Model:    envString("LLM_MODEL", ""),
		},
		Voice: VoiceConfig{
			Provider:      envString("VOICE_PROVIDER", "disabled"),
			WebhookSecret: envString("VOICE_WEBHOOK_SECRET", ""),
		},
		Storage: StorageConfig{
			Provider:        envString("STORAGE_PROVIDER", "disabled"),
			Bucket:          envString("STORAGE_BUCKET", ""),
			Region:          envString("STORAGE_REGION", ""),
			Endpoint:        envString("STORAGE_ENDPOINT", ""),
			AccessKeyID:     envString("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: envString("STORAGE_SECRET_ACCESS_KEY", ""),
		},
		Token: TokenConfig{
			Secret: envString("TOKEN_SECRET", "local-dev-token-secret-change-me-32chars"),
		},
		Jobs: JobsConfig{
			BufferSize: envInt("JOB_BUFFER_SIZE", 32),
			RetryDelay: envDuration("JOB_RETRY_DELAY", 2*time.Second),
		},
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

	if c.App.Env == EnvProduction {
		if strings.TrimSpace(c.Database.URL) == "" && c.Database.RequireInProduction {
			errs = append(errs, fmt.Errorf("DATABASE_URL is required in production"))
		}
		if c.Token.Secret == "local-dev-token-secret-change-me-32chars" {
			errs = append(errs, fmt.Errorf("TOKEN_SECRET must be explicit in production"))
		}
		if c.Email.Provider != "disabled" {
			if strings.TrimSpace(c.Email.APIKey) == "" {
				errs = append(errs, fmt.Errorf("EMAIL_API_KEY is required when EMAIL_PROVIDER is enabled"))
			}
			if strings.TrimSpace(c.Email.FromAddress) == "" {
				errs = append(errs, fmt.Errorf("EMAIL_FROM_ADDRESS is required when EMAIL_PROVIDER is enabled"))
			}
		}
	}

	return errors.Join(errs...)
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func envString(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return fallback
}

func envCSV(key string, fallback []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func envInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}
