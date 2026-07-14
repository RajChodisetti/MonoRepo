package config

import (
	"strings"
	"testing"
)

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Env != EnvLocal {
		t.Fatalf("App.Env = %q, want %q", cfg.App.Env, EnvLocal)
	}
	if cfg.HTTP.Addr == "" {
		t.Fatal("HTTP.Addr is empty")
	}
	if cfg.Token.Secret == "" {
		t.Fatal("Token.Secret is empty")
	}
	if cfg.Email.Provider != providerDisabled {
		t.Fatalf("Email.Provider = %q, want disabled", cfg.Email.Provider)
	}
}

func TestLoadProductionRequiresDatabaseAndExplicitToken(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("TOKEN_SECRET", localDevToken)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{"DATABASE_URL", "TOKEN_SECRET", "TUVI_API_TOKEN", "REDIS_URL"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Load() error = %q, want it to contain %q", msg, want)
		}
	}
}

func TestLoadStagingRequiresExplicitSecrets(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", EnvStaging)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("TOKEN_SECRET", localDevToken)

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{"DATABASE_URL", "TOKEN_SECRET", "TUVI_API_TOKEN", "REDIS_URL"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Load() error = %q, want it to contain %q", msg, want)
		}
	}
}

func TestHTTPLoadPortFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("HTTP_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":9090")
	}
}

func TestHTTPLoadAddrFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_ADDR", ":3001")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTP.Addr != ":3001" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":3001")
	}
}

func TestLoadRejectsMalformedInteger(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_MAX_CONNS", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "DATABASE_MAX_CONNS must be a valid integer") {
		t.Fatalf("Load() error = %q, want DATABASE_MAX_CONNS parse error", err.Error())
	}
}

func TestLoadRejectsMalformedDuration(t *testing.T) {
	clearEnv(t)
	t.Setenv("JOB_RETRY_DELAY", "soon")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "JOB_RETRY_DELAY must be a valid duration") {
		t.Fatalf("Load() error = %q, want JOB_RETRY_DELAY parse error", err.Error())
	}
}

func TestLoadRejectsUnsafeOutreachLimits(t *testing.T) {
	clearEnv(t)
	t.Setenv("OUTREACH_BULK_MAX", "151")
	t.Setenv("OUTREACH_EMAILS_PER_ACCOUNT", "51")
	t.Setenv("OUTREACH_SEND_INTERVAL", "500ms")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want outreach validation errors")
	}
	for _, want := range []string{"OUTREACH_BULK_MAX", "OUTREACH_EMAILS_PER_ACCOUNT", "OUTREACH_SEND_INTERVAL"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Load() error = %q, want %s validation error", err.Error(), want)
		}
	}
}

func TestLoadRejectsMalformedBoolean(t *testing.T) {
	clearEnv(t)
	t.Setenv("EMAIL_DISABLE_SENDING", "maybe")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "EMAIL_DISABLE_SENDING must be true or false") {
		t.Fatalf("Load() error = %q, want EMAIL_DISABLE_SENDING parse error", err.Error())
	}
}

func TestLoadRejectsMalformedHTTPPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("HTTP_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "HTTP_PORT must be a valid TCP port") {
		t.Fatalf("Load() error = %q, want HTTP_PORT parse error", err.Error())
	}
}

func TestLoadRejectsEnabledEmailWithoutCredentials(t *testing.T) {
	clearEnv(t)
	t.Setenv("EMAIL_PROVIDER", "resend")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{"EMAIL_API_KEY", "EMAIL_FROM_ADDRESS"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Load() error = %q, want it to contain %q", msg, want)
		}
	}
}

func TestLoadRejectsEnabledLLMWithoutCredentials(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_PROVIDER", "openai")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{"LLM_API_KEY", "LLM_MODEL"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Load() error = %q, want it to contain %q", msg, want)
		}
	}
}

func TestLoadRejectsEnabledVoiceWithoutWebhookSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("VOICE_PROVIDER", "twilio")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "VOICE_WEBHOOK_SECRET") {
		t.Fatalf("Load() error = %q, want VOICE_WEBHOOK_SECRET validation error", err.Error())
	}
}

func TestLoadRejectsEnabledStorageWithoutCredentials(t *testing.T) {
	clearEnv(t)
	t.Setenv("STORAGE_PROVIDER", "s3")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{
		"STORAGE_BUCKET",
		"STORAGE_ACCESS_KEY_ID",
		"STORAGE_SECRET_ACCESS_KEY",
		"STORAGE_REGION or STORAGE_ENDPOINT",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("Load() error = %q, want it to contain %q", msg, want)
		}
	}
}

func TestValidateRejectsInvalidDatabasePoolSettings(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Database.MinConns = 10
	cfg.Database.MaxConns = 5

	err = cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "DATABASE_MIN_CONNS") {
		t.Fatalf("Validate() error = %q, want DATABASE_MIN_CONNS", err.Error())
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_VERSION",
		"HTTP_ADDR",
		"HTTP_PORT",
		"PORT",
		"CORS_ALLOWED_ORIGINS",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"DATABASE_URL",
		"DATABASE_MAX_CONNS",
		"DATABASE_MIN_CONNS",
		"DATABASE_MAX_CONN_LIFETIME",
		"DATABASE_MAX_CONN_IDLE_TIME",
		"DATABASE_CONNECT_TIMEOUT",
		"REDIS_URL",
		"EMAIL_PROVIDER",
		"EMAIL_API_KEY",
		"EMAIL_FROM_ADDRESS",
		"EMAIL_DISABLE_SENDING",
		"EMAIL_REDIRECT_TO",
		"OUTREACH_BULK_MAX",
		"OUTREACH_EMAILS_PER_ACCOUNT",
		"OUTREACH_SEND_INTERVAL",
		"OUTREACH_ZOHO_ACCOUNTS_JSON",
		"LLM_PROVIDER",
		"LLM_API_KEY",
		"LLM_MODEL",
		"VOICE_PROVIDER",
		"VOICE_WEBHOOK_SECRET",
		"STORAGE_PROVIDER",
		"STORAGE_BUCKET",
		"STORAGE_REGION",
		"STORAGE_ENDPOINT",
		"STORAGE_ACCESS_KEY_ID",
		"STORAGE_SECRET_ACCESS_KEY",
		"TOKEN_SECRET",
		"JWT_ACCESS_TOKEN_TTL",
		"JOB_BUFFER_SIZE",
		"JOB_RETRY_DELAY",
		"TUVI_API_TOKEN",
		"CONSULTATION_NOTIFY_EMAIL",
		"CONSULTATION_TIMEZONE",
		"CONSULTATION_BUSINESS_HOUR_START",
		"CONSULTATION_BUSINESS_HOUR_END",
		"CONSULTATION_SLOT_DURATION_MINUTES",
		"CONSULTATION_DEFAULT_AVAILABILITY_DAYS",
		"CONSULTATION_AVAILABILITY_HORIZON_DAYS",
		"CONSULTATION_GOOGLE_CALENDAR_ID",
		"CONSULTATION_GOOGLE_SERVICE_ACCOUNT_JSON",
		"CONSULTATION_GOOGLE_CALENDAR_DISABLED",
	} {
		t.Setenv(key, "")
	}
}
