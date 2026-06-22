package httpapi

import (
	"net/http"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

const notDeveloperMessage = "You are not a developer."

func allowHealthEndpoints(cfg config.Config, response http.ResponseWriter) bool {
	if cfg.App.Role == auth.RoleDeveloper {
		return true
	}

	writeError(response, http.StatusForbidden, "forbidden", notDeveloperMessage)
	return false
}
