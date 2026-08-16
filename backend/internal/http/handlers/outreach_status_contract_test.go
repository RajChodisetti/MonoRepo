package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/outreach"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type outreachStatusContractRepo struct{}

func (outreachStatusContractRepo) ListEligibleLeads(context.Context, int) ([]outreach.EligibleLead, error) {
	return nil, nil
}

func (outreachStatusContractRepo) CountEligibleLeads(context.Context) (int, error) {
	return 0, nil
}

func (outreachStatusContractRepo) CountSentDeliveriesByPhase(context.Context) (outreach.SentCounts, error) {
	return outreach.SentCounts{Total: 9, Phase1: 5, Phase2: 2, Phase3: 1, Other: 1}, nil
}

func TestOutreachStatusSentCountsRuntimeAndOpenAPIContract(t *testing.T) {
	service := outreach.NewService(
		outreachStatusContractRepo{},
		nil,
		nil,
		nil,
		nil,
		outreach.DemoTokenResolver{},
		nil,
		nil,
		config.EmailConfig{},
		config.OutreachConfig{},
		config.AppURLsConfig{},
		nil,
		nil,
	)
	handler := NewOutreachBulkHandler(
		service,
		func(w http.ResponseWriter, status int, value any) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(value)
		},
		func(w http.ResponseWriter, status int, code string, message string) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
		},
	)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/outreach/bulk-send/status", nil)
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{
		UserID: uuid.New(),
		Role:   auth.RoleInternalAdmin,
	}))
	response := httptest.NewRecorder()

	handler.Status(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		SentCounts outreach.SentCounts `json:"sent_counts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	want := outreach.SentCounts{Total: 9, Phase1: 5, Phase2: 2, Phase3: 1, Other: 1}
	if body.SentCounts != want {
		t.Fatalf("sent_counts = %#v, want %#v", body.SentCounts, want)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() did not return this test file")
	}
	openAPIPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "docs", "openapi", "openapi.yaml")
	openAPI, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contract := string(openAPI)
	for _, required := range []string{
		"sent_counts:\n          $ref: '#/components/schemas/OutreachSentCounts'",
		"required: [total, phase_1, phase_2, phase_3, other]",
		"phase_1:",
		"phase_2:",
		"phase_3:",
		"other:",
	} {
		if !strings.Contains(contract, required) {
			t.Fatalf("OpenAPI outreach status contract missing %q", required)
		}
	}
}
