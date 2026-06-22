package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestInMemoryQueueAppliesDefaults(t *testing.T) {
	queue := NewInMemoryQueue(1)

	job, err := queue.Enqueue(context.Background(), Job{Type: "test"})
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if job.ID == "" {
		t.Fatal("job ID is empty")
	}
	if job.MaxAttempts != 1 {
		t.Fatalf("MaxAttempts = %d, want 1", job.MaxAttempts)
	}
	if job.EnqueuedAt.IsZero() {
		t.Fatal("EnqueuedAt is zero")
	}
}

func TestWorkerProcessesRegisteredHandler(t *testing.T) {
	queue := NewInMemoryQueue(2)
	worker := NewWorker(queue, testLogger(), time.Millisecond)
	processed := make(chan Job, 1)

	if err := worker.Register("test", func(ctx context.Context, job Job) error {
		processed <- job
		return nil
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	enqueued, err := queue.Enqueue(ctx, Job{Type: "test"})
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	select {
	case got := <-processed:
		if got.ID != enqueued.ID {
			t.Fatalf("processed job ID = %q, want %q", got.ID, enqueued.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not process job")
	}
}

func TestWorkerRetriesFailedJob(t *testing.T) {
	queue := NewInMemoryQueue(4)
	worker := NewWorker(queue, testLogger(), time.Millisecond)
	var attempts atomic.Int32
	done := make(chan struct{})

	if err := worker.Register("retry", func(ctx context.Context, job Job) error {
		if attempts.Add(1) == 1 {
			return errors.New("temporary failure")
		}
		close(done)
		return nil
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = worker.Run(ctx)
	}()

	if _, err := queue.Enqueue(ctx, Job{Type: "retry", MaxAttempts: 2}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not retry job")
	}

	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
