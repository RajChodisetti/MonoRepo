# Tuvi Consultation API (Legacy)

The active consultation scheduler now lives in the main MonoRepo API on
`http://localhost:8080` under:

```text
GET  /api/v1/company/consultations/availability
GET  /api/v1/company/consultations/availability/check
POST /api/v1/company/consultations
```

This standalone backend is kept only as legacy reference while callers migrate.

Go backend for Tuvi website consultation booking: slot availability, Google Calendar events, and SMTP notifications.

## Prerequisites

- Go 1.22+
- PostgreSQL (reuse MonoRepo docker Postgres or local install)
- Google Workspace service account with Calendar API enabled (production)
- SMTP credentials (production)

## Quick start (local)

```bash
# 1. Start Postgres
docker compose -f ../../infra/docker/docker-compose.yml up -d postgres

# 2. Create database
createdb -h localhost -U postgres tuvi_website
# or: psql -h localhost -U postgres -c "CREATE DATABASE tuvi_website;"

# 3. Configure env
cp .env.example .env
# Edit .env — for local dev keep GOOGLE_CALENDAR_DISABLED=true and EMAIL_DISABLED=true

# 4. Migrate and run
make migrate-up
make run   # listens on :8090
```

## Environment

| Variable | Description |
|----------|-------------|
| `HTTP_ADDR` | Listen address (default `:8090`) |
| `TUVI_API_TOKEN` | Bearer token for API auth (required) |
| `DATABASE_URL` | Postgres connection string |
| `TIMEZONE` | Business timezone (default `Australia/Sydney`) |
| `BUSINESS_HOUR_START` / `BUSINESS_HOUR_END` | Consultation hours (default 9–17) |
| `SLOT_DURATION_MINUTES` | Slot length (default 30) |
| `GOOGLE_CALENDAR_ID` | Shared team calendar ID |
| `GOOGLE_SERVICE_ACCOUNT_JSON` | Path to service account JSON key |
| `GOOGLE_CALENDAR_DISABLED` | `true` = DB-only mode (local dev) |
| `NOTIFY_EMAIL` | Booking notification recipient (default `contact@tuvisolutions.com`) |
| `SMTP_*` / `EMAIL_*` | SMTP settings |
| `EMAIL_DISABLED` | `true` = skip sending email |

## Google Calendar setup (service account)

1. In [Google Cloud Console](https://console.cloud.google.com/), create or select a project.
2. Enable **Google Calendar API**.
3. Create a **Service Account** → Keys → Add key → JSON. Save as `secrets/google-service-account.json` (gitignored).
4. In Google Calendar, create or open the team consultations calendar.
5. **Share** the calendar with the service account email (`...@....iam.gserviceaccount.com`) with permission **Make changes to events**.
6. Copy the calendar ID (Calendar settings → Integrate calendar) into `GOOGLE_CALENDAR_ID`.
7. Set `GOOGLE_CALENDAR_DISABLED=false` in `.env`.

## API

All `/api/v1/*` routes require `Authorization: Bearer <TUVI_API_TOKEN>`.

### `GET /healthz`

Health check (no auth).

### `GET /api/v1/consultations/availability`

Query: `date` (optional YYYY-MM-DD), `days` (optional, default 5).

Returns free weekday slots within business hours.

### `GET /api/v1/consultations/availability/check`

Query: `date`, `time` — checks one slot; returns `alternatives` if busy.

### `POST /api/v1/consultations`

```json
{
  "date": "2026-07-10",
  "time": "14:00",
  "prospect_name": "Jane Doe",
  "prospect_email": "jane@example.com",
  "source": "voice"
}
```

- `201` — booking created (Google event + DB row + email)
- `409` — slot conflict with `alternatives` array

## Smoke test

```bash
export TOKEN=local-dev-tuvi-api-token-change-me

curl -s http://localhost:8090/healthz

curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8090/api/v1/consultations/availability?days=3"

curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"date":"2026-07-10","time":"14:00","prospect_name":"Test User","source":"voice"}' \
  http://localhost:8090/api/v1/consultations
```

## Voice agent integration

For active development, the voice agent should use the main MonoRepo API, not
this legacy service. Set in `voice-sales-agent/.env`:

```
TUVI_WEBSITE_API_URL=http://localhost:8080
TUVI_API_TOKEN=<same as backend TUVI_API_TOKEN>
```

The corporate voice agent (`?agent=corporate`) calls the main API company
consultation endpoints via `tuvi_api_client.py`.

Start the voice agent with Docker (recommended):

```bash
cd voice-sales-agent && make dev
```
