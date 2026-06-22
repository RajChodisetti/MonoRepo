# Restaurant Platform

Sales-first restaurant platform for personalized demo websites, tracked outreach,
reservation requests, inbound AI receptionist prototypes, and lightweight content
automation.

## Local Setup

1. Install Go 1.26 or newer.
2. Copy `.env.example` to `.env` and adjust local values.
3. Start PostgreSQL and create a local database:

   ```bash
   createdb restaurant_platform
   ```

4. Export the environment variables you need, or source `.env` in your shell.
5. Run migrations:

   ```bash
   make migrate-up
   ```

6. Start the API:

   ```bash
   make api
   ```

7. Start the worker in another shell:

   ```bash
   make worker
   ```

## Commands

```bash
make api          # run the Go API on HTTP_ADDR
make worker       # run the Go worker with the local in-memory queue
make test         # run backend tests
make fmt          # format backend Go files
make migrate-up   # apply SQL migrations from backend/migrations
make migrate-down # roll back the most recently applied migration
```

## Health Checks

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

`/healthz` reports process health. `/readyz` verifies PostgreSQL readiness when
`DATABASE_URL` is configured.

## Configuration Notes

Local defaults are safe for development. Production startup requires explicit
database and token settings. Provider credentials are optional while providers
are set to `disabled`.
