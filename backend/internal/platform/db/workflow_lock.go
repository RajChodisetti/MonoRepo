package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LockRestaurantWorkflow serializes cross-table lead, review, campaign, and
// delivery mutations for one restaurant. Every caller must acquire it before
// taking restaurant/profile/demo/campaign/delivery row locks.
func LockRestaurantWorkflow(ctx context.Context, tx pgx.Tx, restaurantID uuid.UUID) error {
	if restaurantID == uuid.Nil {
		return fmt.Errorf("restaurant workflow lock requires a restaurant id")
	}
	key := "lead-workflow:" + restaurantID.String()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, key); err != nil {
		return fmt.Errorf("lock restaurant workflow: %w", err)
	}
	return nil
}
