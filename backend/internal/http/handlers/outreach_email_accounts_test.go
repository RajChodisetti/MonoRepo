package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

func TestOutreachEmailAccountsRequiresAuthentication(t *testing.T) {
	handler := testEmailAccountsHandler()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/outreach/email-accounts", nil)
	response := httptest.NewRecorder()

	handler.List(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestOutreachEmailAccountsRejectsUnknownCredentialFields(t *testing.T) {
	handler := testEmailAccountsHandler()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/outreach/email-accounts", strings.NewReader(`{
		"account_key":"sales",
		"mailbox_email":"sales@example.com",
		"credentials":{"client_id":"id","client_secret":"secret","refresh_token":"refresh"},
		"plaintext_backup":true
	}`))
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{
		UserID: uuid.New(), Role: auth.RoleInternalAdmin,
	}))
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func testEmailAccountsHandler() *OutreachEmailAccountsHandler {
	writeJSON := func(w http.ResponseWriter, status int, value any) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(value)
	}
	writeError := func(w http.ResponseWriter, status int, code, message string) {
		writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
	}
	return NewOutreachEmailAccountsHandler(nil, writeJSON, writeError)
}
