package main

import (
	"context"
	"log"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/app"
)

func main() {
	ctx := context.Background()

	api, err := app.NewAPI(ctx)
	if err != nil {
		log.Fatalf("api startup failed: %v", err)
	}
	if err := api.Run(ctx); err != nil {
		log.Fatalf("api stopped with error: %v", err)
	}
}
