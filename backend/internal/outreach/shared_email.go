package outreach

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/auth"
	"github.com/rajchodisetti/restaurant-platform/backend/internal/restaurants"
)

type SharedEmailRestaurant struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
}

type SharedEmailGroup struct {
	Email              string                  `json:"email"`
	RestaurantCount    int                     `json:"restaurant_count"`
	BlockedForOutreach bool                    `json:"blocked_for_outreach"`
	Restaurants        []SharedEmailRestaurant `json:"restaurants"`
}

type SharedEmailGroupList struct {
	Groups []SharedEmailGroup `json:"groups"`
	Total  int                `json:"total"`
}

const sharedEmailGroupsQuery = `
	WITH email_groups AS MATERIALIZED (
	  SELECT lower(trim(email)) AS email, count(*)::int AS restaurant_count
	  FROM restaurants
	  WHERE lower(trim(email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
	  GROUP BY lower(trim(email))
	  HAVING count(*) > 1
	), paged_groups AS (
	  SELECT email, restaurant_count
	  FROM email_groups
	  ORDER BY restaurant_count DESC, email
	  LIMIT $1 OFFSET $2
	)
	SELECT groups.email,
	       groups.restaurant_count,
	       restaurant.id,
	       restaurant.name,
	       restaurant.status
	FROM paged_groups groups
	JOIN restaurants restaurant
	  ON lower(trim(restaurant.email)) = groups.email
	ORDER BY groups.restaurant_count DESC, groups.email, restaurant.name, restaurant.id`

const sharedEmailGroupCountQuery = `
	SELECT count(*)
	FROM (
	  SELECT lower(trim(email))
	  FROM restaurants
	  WHERE lower(trim(email)) ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$'
	  GROUP BY lower(trim(email))
	  HAVING count(*) > 1
	) email_groups`

func (service *Service) ListSharedEmailGroups(
	ctx context.Context,
	principal auth.Principal,
	limit int,
	offset int,
) (SharedEmailGroupList, error) {
	if !auth.IsInternalAdmin(principal.Role) {
		return SharedEmailGroupList{}, restaurants.ErrForbidden
	}
	if service.pool == nil {
		return SharedEmailGroupList{}, fmt.Errorf("database pool is not configured")
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	result := SharedEmailGroupList{Groups: []SharedEmailGroup{}}
	if err := service.pool.QueryRow(ctx, sharedEmailGroupCountQuery).Scan(&result.Total); err != nil {
		return SharedEmailGroupList{}, fmt.Errorf("count shared restaurant emails: %w", err)
	}
	rows, err := service.pool.Query(ctx, sharedEmailGroupsQuery, limit, offset)
	if err != nil {
		return SharedEmailGroupList{}, fmt.Errorf("list shared restaurant emails: %w", err)
	}
	defer rows.Close()

	groupIndex := map[string]int{}
	for rows.Next() {
		var (
			email           string
			restaurantCount int
			restaurant      SharedEmailRestaurant
		)
		if err := rows.Scan(
			&email,
			&restaurantCount,
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Status,
		); err != nil {
			return SharedEmailGroupList{}, fmt.Errorf("scan shared restaurant email: %w", err)
		}
		index, exists := groupIndex[email]
		if !exists {
			index = len(result.Groups)
			groupIndex[email] = index
			result.Groups = append(result.Groups, SharedEmailGroup{
				Email:              email,
				RestaurantCount:    restaurantCount,
				BlockedForOutreach: restaurantCount > 3,
				Restaurants:        []SharedEmailRestaurant{},
			})
		}
		result.Groups[index].Restaurants = append(result.Groups[index].Restaurants, restaurant)
	}
	if err := rows.Err(); err != nil {
		return SharedEmailGroupList{}, fmt.Errorf("iterate shared restaurant emails: %w", err)
	}
	return result, nil
}
