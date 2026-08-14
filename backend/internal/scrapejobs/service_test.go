package scrapejobs

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
)

type resumeTestRepository struct {
	Repository
	resumedID uuid.UUID
	job       Job
	err       error
}

func (repo *resumeTestRepository) RetryFailed(_ context.Context, id uuid.UUID) (Job, error) {
	repo.resumedID = id
	return repo.job, repo.err
}

func TestResumeFailedRequiresInternalAdmin(t *testing.T) {
	t.Parallel()

	repo := &resumeTestRepository{}
	service := NewService(repo)
	jobID := uuid.New()

	_, err := service.ResumeFailed(
		context.Background(),
		auth.Principal{UserID: uuid.New(), Role: auth.RoleDeveloper},
		jobID,
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("ResumeFailed() error = %v, want ErrForbidden", err)
	}
	if repo.resumedID != uuid.Nil {
		t.Fatalf("repository resumed id = %s before authorization", repo.resumedID)
	}
}

func TestResumeFailedUsesDurableRepositoryPath(t *testing.T) {
	t.Parallel()

	jobID := uuid.New()
	want := Job{ID: jobID, Status: StatusQueued}
	repo := &resumeTestRepository{job: want}
	service := NewService(repo)
	principal := auth.Principal{UserID: uuid.New(), Role: auth.RoleInternalAdmin}

	got, err := service.ResumeFailed(context.Background(), principal, jobID)
	if err != nil {
		t.Fatalf("ResumeFailed() error = %v", err)
	}
	if got.ID != want.ID || got.Status != StatusQueued {
		t.Fatalf("ResumeFailed() job = %#v, want %#v", got, want)
	}
	if repo.resumedID != jobID {
		t.Fatalf("repository resumed id = %s, want %s", repo.resumedID, jobID)
	}

	if _, err := service.RetryFailed(context.Background(), principal, jobID); err != nil {
		t.Fatalf("legacy RetryFailed() error = %v", err)
	}
}
