package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

var ErrHandlerNotRegistered = errors.New("job handler is not registered")

type Job struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	EnqueuedAt     time.Time       `json:"enqueued_at"`
}

type Handler func(context.Context, Job) error

type Queue interface {
	Enqueue(ctx context.Context, job Job) (Job, error)
	Jobs() <-chan Job
}

type InMemoryQueue struct {
	jobs chan Job
}

func NewInMemoryQueue(bufferSize int) *InMemoryQueue {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &InMemoryQueue{jobs: make(chan Job, bufferSize)}
}

func (q *InMemoryQueue) Enqueue(ctx context.Context, job Job) (Job, error) {
	if job.Type == "" {
		return Job{}, fmt.Errorf("job type is required")
	}
	if len(job.Payload) == 0 {
		job.Payload = json.RawMessage(`{}`)
	}
	if job.ID == "" {
		job.ID = newJobID()
	}
	if job.MaxAttempts < 1 {
		job.MaxAttempts = 1
	}
	if job.EnqueuedAt.IsZero() {
		job.EnqueuedAt = time.Now().UTC()
	}

	select {
	case q.jobs <- job:
		return job, nil
	case <-ctx.Done():
		return Job{}, ctx.Err()
	}
}

func (q *InMemoryQueue) Jobs() <-chan Job {
	return q.jobs
}

type Worker struct {
	queue      Queue
	handlers   map[string]Handler
	log        *slog.Logger
	retryDelay time.Duration
}

func NewWorker(queue Queue, log *slog.Logger, retryDelay time.Duration) *Worker {
	if retryDelay <= 0 {
		retryDelay = time.Second
	}
	return &Worker{
		queue:      queue,
		handlers:   map[string]Handler{},
		log:        log,
		retryDelay: retryDelay,
	}
}

func (w *Worker) Register(jobType string, handler Handler) error {
	if jobType == "" {
		return fmt.Errorf("job type is required")
	}
	if handler == nil {
		return fmt.Errorf("job handler is required")
	}
	w.handlers[jobType] = handler
	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job := <-w.queue.Jobs():
			w.process(ctx, job)
		}
	}
}

func (w *Worker) process(ctx context.Context, job Job) {
	handler, ok := w.handlers[job.Type]
	if !ok {
		w.log.ErrorContext(ctx, "job_handler_missing", "job_id", job.ID, "job_type", job.Type)
		return
	}

	if err := handler(ctx, job); err != nil {
		job.Attempts++
		if job.Attempts < job.MaxAttempts {
			w.log.WarnContext(ctx, "job_retry_scheduled",
				"job_id", job.ID,
				"job_type", job.Type,
				"attempts", job.Attempts,
				"max_attempts", job.MaxAttempts,
				"error", err,
			)
			time.AfterFunc(w.retryDelay, func() {
				_, enqueueErr := w.queue.Enqueue(context.Background(), job)
				if enqueueErr != nil {
					w.log.ErrorContext(ctx, "job_retry_enqueue_failed", "job_id", job.ID, "error", enqueueErr)
				}
			})
			return
		}

		w.log.ErrorContext(ctx, "job_failed",
			"job_id", job.ID,
			"job_type", job.Type,
			"attempts", job.Attempts,
			"max_attempts", job.MaxAttempts,
			"error", err,
		)
		return
	}

	w.log.InfoContext(ctx, "job_completed", "job_id", job.ID, "job_type", job.Type)
}

const SampleJobType = "sample.log"

type SamplePayload struct {
	Message string `json:"message"`
}

func NewSampleJob(message string) (Job, error) {
	payload, err := json.Marshal(SamplePayload{Message: message})
	if err != nil {
		return Job{}, err
	}
	return Job{
		Type:           SampleJobType,
		Payload:        payload,
		IdempotencyKey: "sample:" + message,
		MaxAttempts:    3,
	}, nil
}

func SampleHandler(log *slog.Logger) Handler {
	return func(ctx context.Context, job Job) error {
		var payload SamplePayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}
		log.InfoContext(ctx, "sample_job_processed", "job_id", job.ID, "message", payload.Message)
		return nil
	}
}

func newJobID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("job-%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
