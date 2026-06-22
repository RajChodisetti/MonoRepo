# Work Log — 22 June 2026

**Scope:** Backend foundation (P1-E01) and early auth prep  
**Area:** `backend/` only (no frontend changes)

This document summarizes everything delivered in today's session: what was built, why it matters, how to run it locally, and what comes next.

---

## Summary

Today's work took the repo from an initial scaffold to a **production-style backend foundation** that can start reliably against PostgreSQL, expose health endpoints, run migrations, and use a clean repository pattern for database access.

| Ticket | Title | Status |
|--------|-------|--------|
| P1-001 | Initialize repository and Go services | **Complete** |
| P1-002 | Typed configuration and startup validation | **Complete** |
| P1-003 | PostgreSQL migrations + local Docker | **Complete** |
| P1-004 | Database access layer / repository pattern | **Complete** |
| P1-005 | API router, middleware, health endpoints | **Complete** |
| P1-006 | Async job abstraction | Partial (in-memory queue exists; not formally closed) |
| P1-E02 prep | Users table + user repository | Partial (schema + repo only; no login API yet) |

---

## 1. API server — Fiber with graceful shutdown

The API still uses the existing `net/http` router and middleware, but it now runs through **Go Fiber** with graceful shutdown on `SIGINT` / `SIGTERM`.

**Why:** Matches the pattern used in the reference `sscbackend` project — Fiber handles listening, while our HTTP handlers stay unchanged and testable.

**How it works:**

- `backend/cmd/api/main.go` — thin entry point
- `backend/internal/app/api.go` — creates Fiber app, mounts router via `adaptor.HTTPHandler`, listens in a goroutine, shuts down with a 10s timeout, then closes the DB pool

**Important:** Do not run `go run internal/app/api.go`. Always use:

```bash
make api
# or
go run ./backend/cmd/api
```

---

## 2. P1-002 — Typed config with strict validation

Configuration lives in `backend/internal/platform/config/` with two files:

- `config.go` — typed structs and validation rules
- `env.go` — strict env parsing (no silent fallbacks on bad values)

**What was added / tightened:**

- All major sections: app, HTTP, logging, database, Redis, email, LLM, voice, storage, token, jobs
- **Strict parsing** — invalid integers, durations, booleans, and ports fail at startup with clear errors
- **Provider validation** — when email/LLM/voice/storage is not `disabled`, required credentials must be set
- **Production + staging** — require explicit `DATABASE_URL`, `REDIS_URL`, and `TOKEN_SECRET`
- **Local defaults** — safe: providers disabled, dev token placeholder, DB optional in local env
- **Dotenv loading** — reads `.env`, `backend/.env`, and `../.env`

**New env key added today:**

- `APP_ROLE` — controls access to `/healthz` and `/readyz` (see section 8)

**Tests:** 13 tests in `config_test.go` covering defaults, production/staging rules, malformed values, and provider validation.

---

## 3. P1-003 — Docker PostgreSQL + migrations

### Docker Compose

File: `infra/docker/docker-compose.yml`

- PostgreSQL 16 on `localhost:5432`
- Database: `restaurant_platform`
- User / password: `postgres` / `postgres`
- Named volume `postgres_data` for persistence
- Healthcheck so `make db-up` waits until DB is ready

### Makefile database commands

| Command | What it does |
|---------|----------------|
| `make db-up` | Start Postgres container and wait until healthy |
| `make db-down` | Stop container (data kept) |
| `make db-reset` | Stop container and delete volume (fresh DB) |
| `make migrate-up` | Apply pending SQL migrations |
| `make migrate-down` | Roll back the last applied migration |
| `make setup` | `db-up` + `migrate-up` |
| `make dev` | `setup` + `api` |

### Migrations

| File | Purpose |
|------|---------|
| `000001_foundation.up.sql` | `app_metadata`, `job_runs` tables + schema baseline row |
| `000002_auth_users.up.sql` | `users` table with roles: `admin`, `user`, `developer` |

Migration runner: `backend/cmd/migrate` — requires DB connection before running.

---

## 4. P1-004 — Database layer and repository pattern

### Connection handling

`backend/internal/platform/db/`:

- `ConnectRequired()` — retries ping until DB is ready (or timeout)
- `ConnectRequiredLogged()` — structured logs: `database_connecting`, `database_connected_successfully`, `database_connection_failed`, `database_disconnected`
- API and worker **refuse to start** if the database is unreachable

### Repository pattern

Established under `backend/internal/repositories/`:

| Package | Interface | Implementations |
|---------|-----------|-----------------|
| `metadata` | `Get(ctx, key)` | `Postgres`, `Mock` |
| `user` | `GetByID`, `GetByEmail`, `Create` | `Postgres`, `Mock` |

Shared errors in `backend/internal/platform/repository/errors.go`:

- `ErrNotFound`
- `ErrConflict`

### Store (wiring layer)

`backend/internal/store/store.go` wires all repositories from a single connection pool and runs startup checks:

- `VerifyFoundation()` — confirms `app_metadata` baseline exists (migration 000001 applied)
- `verifyUsersTable()` — confirms `users` table exists (migration 000002 applied)
- `VerifyStartup()` — runs both checks before API/worker accept traffic

**Why a separate `store` package:** Avoids import cycles between `db` and repository packages.

### Tests

Mock-based tests (no real DB required for CI):

- `repositories/metadata/repository_test.go`
- `repositories/user/repository_test.go`
- `store/store_test.go`
- `platform/db/db_test.go` — connection failure cases

Run: `make test` or `go test ./backend/... -v`

---

## 5. P1-005 — Router, middleware, and health endpoints

File: `backend/internal/http/`

### Endpoints

| Route | Method | Purpose |
|-------|--------|---------|
| `/healthz` | GET | Service health (name, env, version) |
| `/readyz` | GET | Database ping with 2s timeout |

### Middleware chain (outer → inner)

1. **Request ID** — sets/propagates `X-Request-ID`
2. **Access log** — structured `http_request` log with method, path, status, bytes, duration
3. **CORS** — origins from `CORS_ALLOWED_ORIGINS` env
4. **Recovery** — catches panics, returns safe `500` without exposing stack to client

### Response helpers

`response.go` — consistent JSON success and error shapes.

### Tests

`router_test.go` covers healthz, readyz (OK / DB missing / DB error), recovery, and role gating.

---

## 6. Users schema and repository (P1-E02 prep)

Not full authentication yet — only the data layer foundation.

**Migration `000002_auth_users`:**

- `users` table with `email`, `password_hash`, `full_name`, `role`, `is_active`
- Roles: `admin`, `user`, `developer`
- Unique email constraint

**Code:**

- `backend/internal/auth/roles.go` — role constants and `ValidRole()`
- `backend/internal/repositories/user/` — Postgres + Mock implementations

**Still missing for P1-E02:**

- Login API
- Password hashing (bcrypt)
- JWT / session middleware
- Protected routes
- `restaurant_members` and tenant isolation

---

## 7. In-memory job queue (P1-006 — partial)

`backend/internal/jobs/` already has:

- Enqueue interface
- Worker with handler registration
- Retry support
- Tests for defaults, processing, and retries

Formal P1-006 sign-off was not the focus today; durable Postgres-backed jobs (`job_runs` table exists in migration 000001) are a follow-up.

---

## 8. APP_ROLE — health endpoint access gate

`APP_ROLE` in `.env` controls who can hit `/healthz` and `/readyz`:

| Value | Health endpoints |
|-------|------------------|
| `developer` | Allowed (200) |
| `admin` or `user` | Forbidden (403) — `"You are not a developer."` |

**Note:** This is a **local dev gate** on health routes only. It is separate from the `users.role` column in the database, which will power real RBAC after login is implemented.

Default in `.env.example`: `APP_ROLE=developer`

---

## 9. Environment files

`.env.example` and `backend/.env` were aligned to have **exactly the same keys** — no extras, no missing keys.

Key groups:

- App: `APP_NAME`, `APP_ENV`, `APP_VERSION`, `APP_ROLE`
- HTTP: `HTTP_ADDR`, `CORS_ALLOWED_ORIGINS`
- Logging: `LOG_LEVEL`, `LOG_FORMAT`
- Database pool settings
- Provider placeholders (disabled by default)
- Jobs: `JOB_BUFFER_SIZE`, `JOB_RETRY_DELAY`

---

## 10. How to run locally (quick reference)

From repo root:

```bash
# Full stack: DB + migrations + API
make dev

# Or step by step
make db-up
make migrate-up
make api

# Worker (second terminal)
make worker

# Tests
make test
go test ./backend/... -v

# Format Go code
make fmt
```

Health checks (requires `APP_ROLE=developer`):

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Connect with TablePlus / any SQL client:

- Host: `localhost:5432`
- User: `postgres`
- Password: `postgres`
- Database: `restaurant_platform`

---

## 11. Ticket verification notes

### P1-004 — "Tests with test DB or mocked repository"

- **Mocked repository tests:** Yes
- **Real PostgreSQL integration tests for repos:** Not yet (optional follow-up)

### P1-005 — Complete against backlog

All six acceptance criteria met. Minor optional gaps: no dedicated CORS or access-log unit tests; no `/api/v1` prefix yet (comes with domain APIs).

---

## 12. What comes next

Recommended build order:

1. **P1-007** — Admin login API, bcrypt password hashing, JWT/session
2. **P1-008** — Auth middleware and role checks on protected routes
3. **P1-009** — `restaurant_members` + tenant isolation
4. **P1-010** — Restaurant/profile/menu CRUD
5. Optional — Postgres integration tests for repositories
6. Optional — Durable job queue using `job_runs` table

---

## 13. Files touched / created (main areas)

```text
backend/cmd/api/
backend/cmd/worker/
backend/cmd/migrate/
backend/internal/app/          # API, worker, runtime
backend/internal/http/         # router, middleware, health access
backend/internal/platform/
  config/                      # config.go, env.go, tests
  db/                          # pool, connect logging, tests
  migrations/                  # migration runner
  repository/                  # shared errors
  logger/
backend/internal/repositories/
  metadata/
  user/
backend/internal/store/
backend/internal/auth/
backend/internal/jobs/
backend/migrations/            # 000001, 000002
infra/docker/docker-compose.yml
Makefile
.env.example
backend/.env
README.md
```

---

## 14. Checks run

- `make test` / `go test ./backend/...` — all backend unit tests pass
- `make migrate-up` — migrations 000001 and 000002 applied successfully
- `make api` — server starts after DB connect + `VerifyStartup()`
- Manual curl smoke on `/healthz` and `/readyz`
