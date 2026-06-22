# Restaurant Platform

Sales-first restaurant platform for personalized demo websites, tracked outreach,
reservation requests, inbound AI receptionist prototypes, and lightweight content
automation.

## Prerequisites

- Go 1.26 or newer
- Docker Desktop (for local PostgreSQL)

## Local Setup

Run these from the **repo root** (`MonoRepo/`).

### 1. Environment

Copy the example env file and adjust if needed:

```bash
cp .env.example .env
```

Or use `backend/.env` — the app loads `.env`, `backend/.env`, and `../.env` in that order.

Default database URL (matches Docker setup):

```text
DATABASE_URL=postgres://postgres:postgres@localhost:5432/restaurant_platform?sslmode=disable
```

**Important:** Set `APP_ROLE=developer` in your env file so `/healthz` and `/readyz`
are accessible during local development. Other roles (`admin`, `user`) receive 403 on
those endpoints.

### 2. Quick start (recommended)

Start database, run migrations, and launch the API in one go:

```bash
make dev
```

Or step by step:

```bash
make setup   # db-up + migrate-up
make api     # start API (requires database + migrations)
```

### 3. Worker (optional, second terminal)

```bash
make worker
```

## Commands

All commands below are run from the **repo root**.

### Database (Docker)

```bash
make db-up      # start PostgreSQL and wait until healthy
make db-down    # stop database container (data is kept)
make db-reset   # stop database and delete all data (fresh start)
```

PostgreSQL runs on `localhost:5432` via `infra/docker/docker-compose.yml`.

| Setting  | Value                 |
| -------- | --------------------- |
| Host     | `localhost`           |
| Port     | `5432`                |
| User     | `postgres`            |
| Password | `postgres`            |
| Database | `restaurant_platform` |

### Migrations

```bash
make migrate-up    # apply pending SQL migrations
make migrate-down  # roll back the most recently applied migration
```

Direct Go equivalent (no Make):

```bash
go run ./backend/cmd/migrate up
go run ./backend/cmd/migrate down
```

Migration files live in `backend/migrations/`.

Current migrations:

- `000001_foundation` — `app_metadata`, `job_runs`
- `000002_auth_users` — `users` table

**Rollback notes:**

- `migrate-down` only reverses the **last applied** migration.
- Each migration has a matching `.down.sql` file (for example `000001_foundation.down.sql`).
- On production, take a backup before rolling back schema changes.

### App

```bash
make api      # run the Go API (Fiber) on HTTP_ADDR
make worker   # run the Go worker with the in-memory job queue
make test     # run all backend tests
make fmt      # format backend Go files
```

Direct Go equivalents (no Make):

```bash
go run ./backend/cmd/api       # start API
go run ./backend/cmd/worker    # start worker
go test ./backend/...          # run tests
go test ./backend/... -v       # run tests with verbose output
```

**Note:** Always start the API via `cmd/api`, not `internal/app/api.go`:

```bash
# correct
go run ./backend/cmd/api

# wrong — package app is not main
go run ./backend/internal/app/api.go
```

### Shortcuts

```bash
make setup    # make db-up + make migrate-up
make dev      # make setup + make api
```

## Health Checks

Requires `APP_ROLE=developer` in your env file.

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

| Endpoint   | Purpose                                              |
| ---------- | ---------------------------------------------------- |
| `/healthz` | Process is running (returns service name, env, version) |
| `/readyz`  | PostgreSQL is connected and ready                    |

The API starts only after:

1. Database connection succeeds (with retry)
2. Foundation and users migrations are verified (`VerifyStartup`)

On startup you should see logs like `database_connected_successfully`.

## Configuration Notes

Local defaults are safe for development. Production and staging require explicit
`DATABASE_URL`, `REDIS_URL`, and `TOKEN_SECRET` settings. Provider credentials
are optional while providers are set to `disabled`.

Key environment variables:

| Variable              | Purpose                                      |
| --------------------- | -------------------------------------------- |
| `APP_ENV`             | `local`, `test`, `staging`, or `production`  |
| `APP_ROLE`            | `developer`, `admin`, or `user` (health gate)  |
| `HTTP_ADDR`           | API listen address (default `:8080`)         |
| `CORS_ALLOWED_ORIGINS`| Comma-separated allowed origins              |
| `DATABASE_URL`        | PostgreSQL connection string                 |
| `TOKEN_SECRET`        | Signing secret (min 32 chars in prod/staging)  |

See `.env.example` for the full list.

## Documentation

- [Today's work log (2026-06-22)](docs/work-log/2026-06-22-backend-foundation-session.md) — detailed session notes
- [Phase 1 technical backlog](PHASE1_TECHNICAL_BACKLOG.md) — tickets and acceptance criteria
- [AGENTS.md](AGENTS.md) — coding-agent operating contract
