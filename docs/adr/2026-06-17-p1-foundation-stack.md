# ADR: P1 Foundation Stack

Date: 2026-06-17
Status: Proposed

## Context

P1-E01 needs a small backend foundation for API startup, worker startup,
configuration, PostgreSQL connectivity, migrations, health checks, logging, and
jobs. The repository currently contains planning docs and no application code.

## Decision

Use the Go standard library `net/http` router and middleware for the first API
foundation. Use `pgxpool` for PostgreSQL connection pooling. Use a small internal
migration runner over SQL files in `backend/migrations`. Use an in-memory job
queue for the first worker path, with a database `job_runs` baseline table added
for later durable job work.

## Options Considered

- `net/http` versus chi/gin/fiber for routing.
- `pgxpool` versus an ORM or `database/sql`.
- Internal migration runner versus goose/atlas/golang-migrate.
- In-memory queue versus Redis/asynq or a Postgres-backed queue.

## Consequences

This keeps P1-E01 dependency-light and easy to run locally. It postpones routing
framework, ORM, and durable queue commitments until more domain APIs exist. The
worker interface and migration table leave room to move to Redis/asynq,
Postgres-backed jobs, or Temporal later.

## Rollback / Revisit Trigger

Revisit if route patterns become complex, if SQL generation becomes valuable
after restaurant CRUD, or if jobs need durability before outreach/content work.
