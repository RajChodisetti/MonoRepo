package httpapi_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	httpapi "github.com/rajchodisetti/restaurant-platform/backend/internal/http"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/logger"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/demos"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/store"
)

const testTokenSecret = "local-dev-token-secret-change-me-32chars"

type fakeReadiness struct {
	err error
}

func (readiness fakeReadiness) Ping(context.Context) error {
	if readiness.err != nil {
		return readiness.err
	}
	return nil
}

func testRouter(t *testing.T, readiness fakeReadiness) http.Handler {
	t.Helper()
	return testRouterWithStores(t, readiness, &auth.Mock{}, &restaurants.Mock{}, &restaurants.MembershipMock{}, &demos.Mock{})
}

func testRouterWithUserRepo(t *testing.T, readiness fakeReadiness, users auth.Repository) http.Handler {
	t.Helper()
	return testRouterWithStores(t, readiness, users, &restaurants.Mock{}, &restaurants.MembershipMock{}, &demos.Mock{})
}

func testRouterWithStores(
	t *testing.T,
	readiness fakeReadiness,
	users auth.Repository,
	restaurantsRepo restaurants.Repository,
	memberships restaurants.MembershipRepository,
	demosRepo demos.Repository,
) http.Handler {
	t.Helper()

	cfg := testConfig(t)
	dataStore := store.NewWithRepositories(nil, nil, users, restaurantsRepo, memberships, demosRepo)
	return httpapi.NewRouter(logger.NewWithWriter(cfg.Logging, io.Discard), readiness, dataStore, cfg)
}

func testConfig(t *testing.T) config.Config {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	cfg.App.Env = config.EnvTest
	cfg.Logging.Format = "text"
	cfg.Logging.Level = "error"
	cfg.Token.Secret = testTokenSecret
	cfg.Token.AccessTokenTTL = time.Hour
	return cfg
}

func developerToken(t *testing.T, tokens *auth.TokenManager) string {
	t.Helper()
	if tokens == nil {
		tokens = auth.NewTokenManager(testTokenSecret, time.Hour)
	}
	token, _, err := tokens.IssueToken(uuid.New(), "dev@example.com", auth.RoleDeveloper)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	return token
}
