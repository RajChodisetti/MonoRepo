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
	if cfg.Email.Provider != "disabled" {
		t.Fatalf("Email.Provider = %q, want disabled", cfg.Email.Provider)
	}
}

func TestLoadProductionRequiresDatabaseAndExplicitToken(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("TOKEN_SECRET", "local-dev-token-secret-change-me-32chars")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}

	msg := err.Error()
	for _, want := range []string{"DATABASE_URL", "TOKEN_SECRET"} {
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
		t.Fatalf("Validate() error = %q, want DATABASE_MIN_CONNS", err)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"APP_NAME",
		"APP_ENV",
		"APP_VERSION",
		"HTTP_ADDR",
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
		"JOB_BUFFER_SIZE",
		"JOB_RETRY_DELAY",
	} {
		t.Setenv(key, "")
	}
}
