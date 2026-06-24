package httpapi

import (
	"net/http"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

const authorizationHeader = "Authorization"

func RequireAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := bearerToken(r.Header.Get(authorizationHeader))
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Missing or invalid authorization header.")
				return
			}

			claims, err := tokens.ParseToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token.")
				return
			}

			principal, err := claims.Principal()
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid token claims.")
				return
			}

			ctx := auth.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return RequireAnyRole(role)
}

func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required.")
				return
			}
			if _, ok := allowed[principal.Role]; !ok {
				writeError(w, http.StatusForbidden, "forbidden", "You do not have permission to access this resource.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}
