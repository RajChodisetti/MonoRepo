# Service inventory

Run commands from the repository root unless noted otherwise.

## Ports

| Port | Service |
| --- | --- |
| `3000` | Restaurant template |
| `3001` | Tuvi corporate website and public AI review |
| `3002` | Internal admin portal |
| `5432` | PostgreSQL |
| `8080` | Main Go API |
| `8081` | Swagger UI |
| `8000` | Voice sales agent |
| `5173` | Restaurant services catalog |

## Long-running services

| Service | Start | Purpose |
| --- | --- | --- |
| PostgreSQL | `make db-up` | Source of truth for restaurants, sequence state, jobs, quotas, and consultations. |
| API | `make api` | Private/admin APIs, public restaurant reports, unsubscribe confirmation, and delivery controls. |
| Worker | `make worker` | Durable jobs and quota-managed plain-text outreach. |
| Scrape worker | Compose `scrape-worker` | Places-first discovery, targeted Apollo enrichment, and direct import. |
| Admin | `cd apps/web && npm run dev` | Sequence editor, recipient progress, restaurants, scrape jobs, and controls. |
| Corporate website | `npm --prefix web run dev -- -p 3001` | Marketing pages, consultation booking, and digital-footprint AI review. |
| Restaurant template | `cd template && npm run dev` | Public restaurant/demo renderer. |
| Voice agent | `cd voice-sales-agent && make dev` | Corporate and restaurant browser voice runtime. |

OCR is retired. There is no OCR service, one-shot OCR job, OCR cron, or OCR
provider configuration.

## Stack commands

| Command | Starts |
| --- | --- |
| `make setup` | PostgreSQL and migrations |
| `make dev` | PostgreSQL, migrations, and API |
| `make start` | PostgreSQL, migrations, API, and worker |
| `make up` | Full Docker stack including scrape worker |
| `make swagger` | Local OpenAPI viewer |

## Data and outreach flow

```text
scrape-worker
  -> Google Places discovery
  -> optional targeted Apollo work-email/owner enrichment
  -> PostgreSQL restaurant + inferred-business source evidence
  -> active approved sequence enrollment when name + valid email exist

worker (only when persisted email job is enabled)
  -> due follow-ups first
  -> then new recipients
  -> Gmail quota claim + idempotent delivery attempt
  -> confirmed send advances integer sequence step and next-due timestamp
  -> failure/unknown leaves the step unchanged

public AI review
  -> Places details/reviews/media
  -> concurrent website capture and low-latency vision review
  -> deterministic partial fallback within the response deadline
```

The outreach path does not require image analysis, a generated profile, a
published demo, or a restaurant-specific approved campaign. Suppression and
lifecycle gates still fail closed. Google listing media is resolved live with
attribution; owner/licensed media requires explicit admin approval.

See [lead-scrape-outreach.md](./runbooks/lead-scrape-outreach.md) for deployment
and operational checks.
