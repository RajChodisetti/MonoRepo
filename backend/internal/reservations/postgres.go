package reservations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	repository "github.com/rajchodisetti/restaurant-platform/backend/internal/platform/errors"
)

const DefaultTimezone = "Australia/Sydney"

var SydneyLocation = mustLoadLocation(DefaultTimezone)

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

func (repo *Postgres) GetRestaurantTimezone(ctx context.Context, restaurantID uuid.UUID) (string, error) {
	if repo.pool == nil {
		return "", fmt.Errorf("database pool is not configured")
	}

	const query = `
		SELECT COALESCE(city, ''), COALESCE(state, ''), COALESCE(country, '')
		FROM restaurant_profiles
		WHERE restaurant_id = $1`

	var city, state, country string
	err := repo.pool.QueryRow(ctx, query, restaurantID).Scan(&city, &state, &country)
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultTimezone, nil
	}
	if err != nil {
		return "", fmt.Errorf("get restaurant timezone: %w", err)
	}
	return timezoneForAustralianLocation(city, state, country), nil
}

func timezoneForAustralianLocation(city, state, country string) string {
	normalize := func(value string) string {
		value = strings.ToLower(strings.TrimSpace(value))
		value = strings.NewReplacer(".", "", ",", "").Replace(value)
		return strings.Join(strings.Fields(value), " ")
	}

	city = normalize(city)
	state = normalize(state)
	country = normalize(country)
	if country != "" && country != "australia" && country != "au" && country != "aus" {
		return DefaultTimezone
	}

	switch city {
	case "sydney":
		return "Australia/Sydney"
	case "melbourne":
		return "Australia/Melbourne"
	case "brisbane":
		return "Australia/Brisbane"
	case "adelaide":
		return "Australia/Adelaide"
	case "perth":
		return "Australia/Perth"
	}

	switch state {
	case "new south wales", "nsw", "australian capital territory", "act":
		return "Australia/Sydney"
	case "victoria", "vic":
		return "Australia/Melbourne"
	case "queensland", "qld":
		return "Australia/Brisbane"
	case "south australia", "sa":
		return "Australia/Adelaide"
	case "western australia", "wa":
		return "Australia/Perth"
	default:
		return DefaultTimezone
	}
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
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get opening hours: %w", err)
	}

	values := map[string]json.RawMessage{}
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]string{}, nil
	}
	if err := json.Unmarshal(raw, &values); err != nil {
		return map[string]string{}, nil
	}

	// Places may also store non-string metadata such as open_now. Ignore that
	// metadata without discarding the weekday strings. Unknown/malformed hours
	// fail closed instead of inventing a generic service window.
	hours := make(map[string]string, len(values))
	for key, rawValue := range values {
		var value string
		if err := json.Unmarshal(rawValue, &value); err == nil {
			hours[key] = value
		}
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
	tx, err := repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Reservation{}, fmt.Errorf("begin reservation request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Serialize retries before slot capacity. This closes both the partial
	// unique-index race and concurrent overbooking without a global lock.
	if input.ClientRequestID != "" {
		if _, err := tx.Exec(ctx,
			`SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
			"reservation-client:"+restaurantID.String()+":"+input.ClientRequestID,
		); err != nil {
			return Reservation{}, fmt.Errorf("lock reservation retry: %w", err)
		}
	}
	slotKey := "reservation-slot:" + restaurantID.String() + ":" +
		input.ReservationDate.Format("2006-01-02") + ":" + input.ReservationTime
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, slotKey); err != nil {
		return Reservation{}, fmt.Errorf("lock reservation slot: %w", err)
	}

	if input.ClientRequestID != "" {
		const existingQuery = `
			SELECT id, restaurant_id, guest_name, guest_phone, guest_email, party_size,
			       reservation_date, reservation_time::text, status, source, notes,
			       COALESCE(client_request_id, ''), created_at, updated_at
			FROM reservations
			WHERE restaurant_id = $1 AND client_request_id = $2
			LIMIT 1`
		existing, lookupErr := scanReservation(tx.QueryRow(ctx, existingQuery, restaurantID, input.ClientRequestID))
		if lookupErr == nil {
			if !reservationMatchesInput(existing, input) {
				return Reservation{}, ErrClientRequestConflict
			}
			if err := tx.Commit(ctx); err != nil {
				return Reservation{}, fmt.Errorf("commit reservation retry lookup: %w", err)
			}
			return existing, nil
		}
		if !errors.Is(lookupErr, repository.ErrNotFound) {
			return Reservation{}, lookupErr
		}
	}

	var currentCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM reservations
		WHERE restaurant_id = $1
		  AND reservation_date = $2
		  AND reservation_time = $3::time
		  AND status IN ('pending', 'confirmed')`,
		restaurantID,
		input.ReservationDate.Format("2006-01-02"),
		input.ReservationTime,
	).Scan(&currentCount)
	if err != nil {
		return Reservation{}, fmt.Errorf("recheck reservation slot capacity: %w", err)
	}
	if currentCount >= DefaultMaxTablesPerSlot {
		return Reservation{}, ErrSlotUnavailable
	}

	const query = `
		INSERT INTO reservations (
			restaurant_id, guest_name, guest_phone, guest_email, party_size,
			reservation_date, reservation_time, status, source, notes, client_request_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7::time, $8, $9, $10, NULLIF($11, ''))
		RETURNING id, restaurant_id, guest_name, guest_phone, guest_email, party_size,
		          reservation_date, reservation_time::text, status, source, notes,
		          COALESCE(client_request_id, ''), created_at, updated_at`

	record, err := scanReservation(tx.QueryRow(ctx, query,
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
	if err != nil {
		return Reservation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Reservation{}, fmt.Errorf("commit reservation request: %w", err)
	}
	return record, nil
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
	timezone, err := repo.GetRestaurantTimezone(ctx, restaurantID)
	if err != nil {
		return AvailabilityResult{}, err
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("load restaurant timezone %q: %w", timezone, err)
	}
	serviceDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
	return computeAvailability(ctx, repo, restaurantID, serviceDate, partySize, timezone, location)
}

func computeAvailability(
	ctx context.Context,
	repo *Postgres,
	restaurantID uuid.UUID,
	date time.Time,
	partySize int,
	timezone string,
	location *time.Location,
) (AvailabilityResult, error) {
	hours, err := repo.GetOpeningHours(ctx, restaurantID)
	if err != nil {
		return AvailabilityResult{}, err
	}

	dayName := date.In(location).Weekday().String()
	hoursLine := hours[dayName]
	if hoursLine == "" {
		hoursLine = hours[strings.ToLower(dayName)]
	}
	result := AvailabilityResult{
		AvailableSlots: []string{},
		Timezone:       timezone,
		Date:           date.Format("2006-01-02"),
		PartySize:      partySize,
	}
	if hoursLine == "" {
		return result, nil
	}

	windows := parseHoursRangesInLocation(hoursLine, date, location)
	if len(windows) == 0 {
		return result, nil
	}

	now := time.Now().In(location)
	seen := make(map[string]struct{})
	for _, window := range windows {
		for slot := window.open; slot.Before(window.close); slot = slot.Add(SlotIntervalMinutes * time.Minute) {
			if slot.Before(now) {
				continue
			}

			slotISO := slot.Format(time.RFC3339)
			if _, duplicate := seen[slotISO]; duplicate {
				continue
			}

			slotTime := slot.Format("15:04:05")
			slotDate := time.Date(slot.Year(), slot.Month(), slot.Day(), 0, 0, 0, 0, location)
			count, err := repo.CountBySlot(ctx, restaurantID, slotDate, slotTime)
			if err != nil {
				return AvailabilityResult{}, err
			}
			if count >= DefaultMaxTablesPerSlot {
				continue
			}

			seen[slotISO] = struct{}{}
			result.AvailableSlots = append(result.AvailableSlots, slotISO)
		}
	}

	sort.Strings(result.AvailableSlots)
	return result, nil
}

type hoursWindow struct {
	open  time.Time
	close time.Time
}

func parseHoursRange(hoursLine string, date time.Time) (time.Time, time.Time, bool) {
	windows := parseHoursRangesInLocation(hoursLine, date, SydneyLocation)
	if len(windows) == 0 {
		return time.Time{}, time.Time{}, true
	}
	return windows[0].open, windows[0].close, false
}

func parseHoursRangesInLocation(hoursLine string, date time.Time, location *time.Location) []hoursWindow {
	hoursLine = strings.NewReplacer(
		"\u202f", " ",
		"\u00a0", " ",
		"—", "-",
		"–", "-",
	).Replace(hoursLine)
	hoursLine = strings.Join(strings.Fields(hoursLine), " ")
	line := strings.TrimSpace(strings.ToLower(hoursLine))
	if line == "" || line == "closed" || line == "temporarily closed" {
		return nil
	}
	if strings.Contains(line, "open 24 hours") || strings.TrimSpace(line) == "24 hours" {
		open := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
		return []hoursWindow{{open: open, close: open.AddDate(0, 0, 1)}}
	}

	segments := strings.FieldsFunc(hoursLine, func(r rune) bool {
		return r == ',' || r == ';'
	})
	windows := make([]hoursWindow, 0, len(segments))
	for _, segment := range segments {
		if strings.Contains(strings.ToLower(segment), "closed") {
			continue
		}
		parts := strings.SplitN(strings.TrimSpace(segment), "-", 2)
		if len(parts) != 2 {
			continue
		}

		window, ok := parseHoursWindow(parts[0], parts[1], date, location)
		if ok {
			windows = append(windows, window)
		}
	}
	return windows
}

func parseHoursWindow(openValue, closeValue string, date time.Time, location *time.Location) (hoursWindow, bool) {
	openValue = strings.TrimSpace(openValue)
	closeValue = strings.TrimSpace(closeValue)
	openCandidates := parseClockCandidates(openValue, clockMeridiem(closeValue), date, location)
	closeCandidates := parseClockCandidates(closeValue, clockMeridiem(openValue), date, location)
	if len(openCandidates) == 0 || len(closeCandidates) == 0 {
		return hoursWindow{}, false
	}

	var selected hoursWindow
	var selectedDuration time.Duration
	for _, open := range openCandidates {
		for _, close := range closeCandidates {
			if !close.After(open) {
				close = close.AddDate(0, 0, 1)
			}
			duration := close.Sub(open)
			if duration <= 0 || duration > 24*time.Hour {
				continue
			}
			if selectedDuration == 0 || duration < selectedDuration {
				selected = hoursWindow{open: open, close: close}
				selectedDuration = duration
			}
		}
	}
	if selectedDuration == 0 {
		return hoursWindow{}, false
	}
	return selected, true
}

func parseClock(value string, date time.Time) time.Time {
	return parseClockInLocation(value, date, SydneyLocation)
}

func parseClockCandidates(value, inheritedMeridiem string, date time.Time, location *time.Location) []time.Time {
	if clockMeridiem(value) != "" {
		if parsed := parseClockInLocation(value, date, location); !parsed.IsZero() {
			return []time.Time{parsed}
		}
		return nil
	}

	candidates := make([]time.Time, 0, 2)
	if inheritedMeridiem != "" {
		for _, meridiem := range []string{"AM", "PM"} {
			if parsed := parseClockInLocation(value+" "+meridiem, date, location); !parsed.IsZero() {
				candidates = append(candidates, parsed)
			}
		}
	}
	if len(candidates) == 0 {
		if parsed := parseClockInLocation(value, date, location); !parsed.IsZero() {
			candidates = append(candidates, parsed)
		}
	}
	return candidates
}

func clockMeridiem(value string) string {
	value = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), ".", ""))
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	last := fields[len(fields)-1]
	if last == "AM" || last == "PM" {
		return last
	}
	if strings.HasSuffix(value, "AM") {
		return "AM"
	}
	if strings.HasSuffix(value, "PM") {
		return "PM"
	}
	return ""
}

func parseClockInLocation(value string, date time.Time, location *time.Location) time.Time {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, ".", "")
	layouts := []string{"3:04 PM", "3:04PM", "15:04", "15:04:05", "3 PM", "3PM"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, location); err == nil {
			return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, location)
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
	return ParseReservationSlotInLocation(slotISO, SydneyLocation)
}

func ParseReservationSlotInLocation(slotISO string, location *time.Location) (time.Time, string, error) {
	t, err := time.Parse(time.RFC3339, slotISO)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid slot format: %w", err)
	}
	t = t.In(location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location), t.Format("15:04:05"), nil
}
