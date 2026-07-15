package scrapejobs

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusWaiting   = "waiting"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"

	DefaultNiche             = "restaurant"
	DefaultMaxRequestsWindow = 500
)

var (
	ErrForbidden       = errors.New("scrape jobs require an internal administrator")
	ErrNotFound        = errors.New("scrape job not found")
	ErrInvalidCity     = errors.New("unsupported scrape city")
	ErrInvalidNiche    = errors.New("unsupported scrape niche")
	ErrActiveJobExists = errors.New("an active scrape job already exists")
	ErrNotFailed       = errors.New("scrape job is not failed")
)

type CreateInput struct {
	City  string
	Niche string
}

type Progress struct {
	CellsTotal          int `json:"cells_total"`
	CellsPending        int `json:"cells_pending"`
	CellsCompleted      int `json:"cells_completed"`
	CellsSubdivided     int `json:"cells_subdivided"`
	CellsFailed         int `json:"cells_failed"`
	CellsSaturated      int `json:"cells_saturated"`
	CandidatesTotal     int `json:"candidates_total"`
	CandidatesPending   int `json:"candidates_pending"`
	CandidatesImported  int `json:"candidates_imported"`
	CandidatesDuplicate int `json:"candidates_duplicate"`
	CandidatesFailed    int `json:"candidates_failed"`
}

type Job struct {
	ID                   uuid.UUID  `json:"id"`
	City                 string     `json:"city"`
	CityKey              string     `json:"city_key"`
	Niche                string     `json:"niche"`
	Status               string     `json:"status"`
	CycleNumber          int        `json:"cycle_number"`
	MaxRequestsPerWindow int        `json:"max_requests_per_window"`
	RequestsUsedWindow   int        `json:"requests_used_window"`
	RequestsUsedTotal    int64      `json:"requests_used_total"`
	WindowStartedAt      *time.Time `json:"window_started_at,omitempty"`
	ResumeAt             *time.Time `json:"resume_at,omitempty"`
	WaitingReason        string     `json:"waiting_reason,omitempty"`
	LastCycleCompletedAt *time.Time `json:"last_cycle_completed_at,omitempty"`
	CurrentCellID        *uuid.UUID `json:"current_cell_id,omitempty"`
	LastError            string     `json:"last_error,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Progress             Progress   `json:"progress"`
}

type TriggerResult struct {
	Created bool `json:"created"`
	Job     Job  `json:"job"`
}
