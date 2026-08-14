package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/scrapejobs"
)

type resumeHandlerRepository struct {
	scrapejobs.Repository
	calls int
}

func (repo *resumeHandlerRepository) RetryFailed(_ context.Context, id uuid.UUID) (scrapejobs.Job, error) {
	repo.calls++
	return scrapejobs.Job{ID: id, Status: scrapejobs.StatusQueued}, nil
}

func TestScrapeJobResumeRequiresAuthAndValidID(t *testing.T) {
	t.Parallel()

	repo := &resumeHandlerRepository{}
	handler := newResumeTestHandler(repo)
	admin := auth.Principal{UserID: uuid.New(), Role: auth.RoleInternalAdmin}

	tests := []struct {
		name       string
		id         string
		principal  *auth.Principal
		serve      func(http.ResponseWriter, *http.Request)
		wantStatus int
	}{
		{
			name:       "resume requires authentication",
			id:         uuid.NewString(),
			serve:      handler.Resume,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "resume rejects invalid job id",
			id:         "not-a-uuid",
			principal:  &admin,
			serve:      handler.Resume,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "legacy retry rejects invalid job id",
			id:         "not-a-uuid",
			principal:  &admin,
			serve:      handler.Retry,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := requestWithPathValue(
				http.MethodPost,
				"/api/v1/scrape-jobs/"+test.id+"/resume",
				"id",
				test.id,
			)
			if test.principal != nil {
				request = request.WithContext(auth.WithPrincipal(request.Context(), *test.principal))
			}
			response := httptest.NewRecorder()
			test.serve(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}

	if repo.calls != 0 {
		t.Fatalf("repository called %d times for rejected requests", repo.calls)
	}
}

func TestScrapeJobResumeRequeuesFailedJob(t *testing.T) {
	t.Parallel()

	repo := &resumeHandlerRepository{}
	handler := newResumeTestHandler(repo)
	id := uuid.New()
	request := requestWithPathValue(
		http.MethodPost,
		"/api/v1/scrape-jobs/"+id.String()+"/resume",
		"id",
		id.String(),
	)
	request = request.WithContext(auth.WithPrincipal(request.Context(), auth.Principal{
		UserID: uuid.New(),
		Role:   auth.RoleInternalAdmin,
	}))
	response := httptest.NewRecorder()

	handler.Resume(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusAccepted, response.Body.String())
	}
	if repo.calls != 1 {
		t.Fatalf("repository calls = %d, want 1", repo.calls)
	}
}

func newResumeTestHandler(repo scrapejobs.Repository) *ScrapeJobHandler {
	return NewScrapeJobHandler(
		scrapejobs.NewService(repo),
		func(w http.ResponseWriter, status int, value any) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(value)
		},
		func(w http.ResponseWriter, status int, code string, message string) {
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
		},
	)
}
