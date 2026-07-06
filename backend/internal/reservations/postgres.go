package reservations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

var SydneyLocation = mustLoadLocation("Australia/Sydney")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.FixedZone("AEST", 10*3600)
	}
	return loc
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (repo *Postgres) RestaurantExists(ctx context.Context, restaurantID uuid.UUID) (bool, error) {
	if repo.pool == nil {
		return false, fmt.Errorf("database pool is not configured")
	}

	const query = `SELECT 1 FROM restaurants WHERE id = $1`
	var exists int
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("restaurant exists: %w", err)
	}
	return true, nil
}

func (repo *Postgres) GetOpeningHours(ctx context.Context, restaurantID uuid.UUID) (map[string]string, error) {
	if repo.pool == nil {
		return nil, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT COALESCE(opening_hours, '{}'::jsonb)
		FROM restaurant_profiles
		WHERE restaurant_id = $1`

	var raw []byte
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return defaultHours(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get opening hours: %w", err)
	}

	hours := map[string]string{}
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &hours); err != nil {
			return defaultHours(), nil
		}
	}
	if len(hours) == 0 {
		return defaultHours(), nil
	}
	return hours, nil
}

func defaultHours() map[string]string {
	return map[string]string{
		"Monday":    "12:00 PM – 10:00 PM",
		"Tuesday":   "12:00 PM – 10:00 PM",
		"Wednesday": "12:00 PM – 10:00 PM",
		"Thursday":  "12:00 PM – 10:00 PM",
		"Friday":    "12:00 PM – 11:00 PM",
		"Saturday":  "12:00 PM – 11:00 PM",
		"Sunday":    "12:00 PM – 9:00 PM",
	}
}

func (repo *Postgres) CountBySlot(ctx context.Context, restaurantID uuid.UUID, date time.Time, slotTime string) (int, error) {
	if repo.pool == nil {
		return 0, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT COUNT(*)
		FROM reservations
		WHERE restaurant_id = $1
		  AND reservation_date = $2
		  AND reservation_time = $3::time
		  AND status IN ('pending', 'confirmed')`

	var count int
	err := repo.pool.QueryRow(ctx, query, restaurantID, date.Format("2006-01-02"), slotTime).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by slot: %w", err)
	}
	return count, nil
}

func (repo *Postgres) GetByClientRequestID(ctx context.Context, restaurantID uuid.UUID, clientRequestID string) (Reservation, error) {
	if repo.pool == nil {
		return Reservation{}, fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT id, restaurant_id, guest_name, guest_phone, guest_email, party_size,
		       reservation_date, reservation_time::text, status, source, notes,
		       COALESCE(client_request_id, ''), created_at, updated_at
		FROM reservations
		WHERE restaurant_id = $1 AND client_request_id = $2
		LIMIT 1`

	return scanReservation(repo.pool.QueryRow(ctx, query, restaurantID, clientRequestID))
}

func (repo *Postgres) Create(ctx context.Context, restaurantID uuid.UUID, input CreateInput) (Reservation, error) {
	if repo.pool == nil {
		return Reservation{}, fmt.Errorf("database pool is not configured")
	}

	source := input.Source
	if source == "" {
		source = SourceVoiceAgent
	}

	const query = `
		INSERT INTO reservations (
			restaurant_id, guest_name, guest_phone, guest_email, party_size,
			reservation_date, reservation_time, status, source, notes, client_request_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7::time, $8, $9, $10, NULLIF($11, ''))
		RETURNING id, restaurant_id, guest_name, guest_phone, guest_email, party_size,
		          reservation_date, reservation_time::text, status, source, notes,
		          COALESCE(client_request_id, ''), created_at, updated_at`

	return scanReservation(repo.pool.QueryRow(ctx, query,
		restaurantID,
		input.GuestName,
		input.GuestPhone,
		input.GuestEmail,
		input.PartySize,
		input.ReservationDate.Format("2006-01-02"),
		input.ReservationTime,
		StatusPending,
		source,
		input.Notes,
		input.ClientRequestID,
	))
}

func scanReservation(scanner interface {
	Scan(dest ...any) error
}) (Reservation, error) {
	var record Reservation
	err := scanner.Scan(
		&record.ID,
		&record.RestaurantID,
		&record.GuestName,
		&record.GuestPhone,
		&record.GuestEmail,
		&record.PartySize,
		&record.ReservationDate,
		&record.ReservationTime,
		&record.Status,
		&record.Source,
		&record.Notes,
		&record.ClientRequestID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, repository.ErrNotFound
	}
	if err != nil {
		return Reservation{}, fmt.Errorf("scan reservation: %w", err)
	}
	return record, nil
}

func ComputeAvailability(
	ctx context.Context,
	repo *Postgres,
	restaurantID uuid.UUID,
	date time.Time,
	partySize int,
) (AvailabilityResult, error) {
	hours, err := repo.GetOpeningHours(ctx, restaurantID)
	if err != nil {
		return AvailabilityResult{}, err
	}

	dayName := date.In(SydneyLocation).Weekday().String()
	hoursLine := hours[dayName]
	if hoursLine == "" {
		hoursLine = hours[strings.ToLower(dayName)]
	}
	if hoursLine == "" {
		for _, v := range hours {
			if v != "" {
				hoursLine = v
				break
			}
		}
	}

	openTime, closeTime, closed := parseHoursRange(hoursLine, date)
	result := AvailabilityResult{
		AvailableSlots: []string{},
		Timezone:       "Australia/Sydney",
		Date:           date.Format("2006-01-02"),
		PartySize:      partySize,
	}

	if closed {
		return result, nil
	}

	now := time.Now().In(SydneyLocation)
	for slot := openTime; slot.Before(closeTime); slot = slot.Add(SlotIntervalMinutes * time.Minute) {
		if date.Format("2006-01-02") == now.Format("2006-01-02") && slot.Before(now) {
			continue
		}

		slotTime := slot.Format("15:04:05")
		count, err := repo.CountBySlot(ctx, restaurantID, date, slotTime)
		if err != nil {
			return AvailabilityResult{}, err
		}
		if count >= DefaultMaxTablesPerSlot {
			continue
		}

		result.AvailableSlots = append(result.AvailableSlots, slot.Format(time.RFC3339))
	}

	sort.Strings(result.AvailableSlots)
	return result, nil
}

func parseHoursRange(hoursLine string, date time.Time) (time.Time, time.Time, bool) {
	line := strings.TrimSpace(strings.ToLower(hoursLine))
	if line == "" || strings.Contains(line, "closed") {
		return time.Time{}, time.Time{}, true
	}

	// Use first segment when hours list multiple windows (e.g. "12–3 pm, 5:30–9 pm").
	if comma := strings.Index(hoursLine, ","); comma >= 0 {
		hoursLine = strings.TrimSpace(hoursLine[:comma])
	}

	// Default lunch/dinner window if unparseable
	defaultOpen := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, SydneyLocation)
	defaultClose := time.Date(date.Year(), date.Month(), date.Day(), 22, 0, 0, 0, SydneyLocation)

	parts := strings.Split(hoursLine, "–")
	if len(parts) < 2 {
		parts = strings.Split(hoursLine, "-")
	}
	if len(parts) < 2 {
		return defaultOpen, defaultClose, false
	}

	open := parseClock(strings.TrimSpace(parts[0]), date)
	close := parseClock(strings.TrimSpace(parts[len(parts)-1]), date)
	if open.IsZero() || close.IsZero() {
		return defaultOpen, defaultClose, false
	}
	return open, close, false
}

func parseClock(value string, date time.Time) time.Time {
	value = strings.TrimSpace(value)
	layouts := []string{"3:04 PM", "3:04PM", "15:04", "3 PM"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, SydneyLocation); err == nil {
			return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, SydneyLocation)
		}
	}
	return time.Time{}
}

func SlotAllowed(slots []string, slotISO string) bool {
	for _, s := range slots {
		if s == slotISO {
			return true
		}
	}
	return false
}

func ParseReservationSlot(slotISO string) (time.Time, string, error) {
	t, err := time.Parse(time.RFC3339, slotISO)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid slot format: %w", err)
	}
	t = t.In(SydneyLocation)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, SydneyLocation), t.Format("15:04:05"), nil
}
