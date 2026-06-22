package main

import (
	"context"
	"log"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/app"
)

func main() {
	ctx := context.Background()

	worker, err := app.NewWorker(ctx)
	if err != nil {
		log.Fatalf("worker startup failed: %v", err)
	}
	if err := worker.Run(ctx); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}
