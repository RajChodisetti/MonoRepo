# Guide: Start a city scrape job on api.tuvisolutions.com

Base URL: `https://api.tuvisolutions.com`

## Docs (Swagger / OpenAPI)

| Resource | URL |
|----------|-----|
| Swagger UI | https://api.tuvisolutions.com/docs/ |
| OpenAPI YAML | https://api.tuvisolutions.com/openapi.yaml |
| Full runbook | [lead-scrape-ocr-outreach.md](./lead-scrape-ocr-outreach.md) |
| Postman collection | [../../postman/Tuvi-Scrape-Jobs.postman_collection.json](../../postman/Tuvi-Scrape-Jobs.postman_collection.json) |

In Swagger: **Authorize** → paste `Bearer <access_token>` (or just the token if the UI adds Bearer).

## Before you start (server-side)

Scraping is **not** finished by the API alone. You need:

1. API up at `https://api.tuvisolutions.com`
2. **`scrape-worker`** container/process running on the VM (same Postgres)
3. Places key configured on the worker (`PLACES_API` / Google Places)
4. An `internal_admin` user (seeded with `make seed-admin` / your prod admin)

If the job stays `queued` forever, the worker is not polling.

## Supported values

- **Cities:** Adelaide, Brisbane, Melbourne, Perth, Sydney  
- **Niches:** `restaurant` (default), `dentist`, `plumber`

## Option A — Postman (recommended)

1. Open Postman → **Import** → select  
   `MonoRepo/postman/Tuvi-Scrape-Jobs.postman_collection.json`
2. Collection variables:
   - `baseUrl` = `https://api.tuvisolutions.com`
   - `adminEmail` / `adminPassword` = your **production** internal admin
   - `city` = e.g. `Melbourne`
   - `niche` = `restaurant`
3. Run in order:
   1. **Login (internal_admin)** — saves `accessToken`
   2. **Admin me** — must be 200
   3. **Trigger city scrape job** — saves `scrapeJobId` (202 new / 200 existing)
   4. **Get scrape job progress** — poll every 30–60s
   5. **List recent scrape jobs** — overview
   6. **Retry** — only if status is `failed`

## Option B — curl (same sequence)

```bash
export TUVI_API=https://api.tuvisolutions.com

# 1) Login
export ADMIN_TOKEN="$(curl -fsS -X POST "$TUVI_API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  --data '{"email":"<admin email>","password":"<admin password>"}' \
  | jq -r '.access_token')"

# 2) Confirm admin
curl -fsS "$TUVI_API/api/v1/admin/me" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# 3) Trigger scrape (idempotent per city+niche)
curl -fsS -X POST "$TUVI_API/api/v1/scrape-jobs" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data '{"city":"Melbourne","niche":"restaurant"}' | jq

# 4) Poll progress (paste id from step 3)
export SCRAPE_JOB_ID=<uuid-from-previous-response>
curl -fsS "$TUVI_API/api/v1/scrape-jobs/$SCRAPE_JOB_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# 5) List jobs
curl -fsS "$TUVI_API/api/v1/scrape-jobs?limit=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

## Option C — Swagger UI

1. Open https://api.tuvisolutions.com/docs/
2. `POST /api/v1/auth/login` → Execute with admin email/password → copy `access_token`
3. Click **Authorize** → `Bearer <token>` → Authorize
4. `POST /api/v1/scrape-jobs` body:
   ```json
   { "city": "Melbourne", "niche": "restaurant" }
   ```
5. `GET /api/v1/scrape-jobs/{id}` to monitor

## How to read status

| Status | Meaning |
|--------|---------|
| `queued` | Waiting for scrape-worker to claim |
| `running` | Worker has a lease; Places grid running |
| `waiting` + `request_limit` | Hit ~500 Places/Apollo calls; resumes after `resume_at` (~24h) |
| `waiting` + `revisit` | Coverage pass done; next cycle later |
| `waiting` + `coverage_incomplete` | Dense cell hit Places cap at max depth |
| `failed` | Permanent error → use `/retry` after fixing config |

Job does **not** end as permanent `completed`; revisits are intentional.

## Important

- Creating a scrape job does **not** send email.
- Local `backend/.env` admin (`admin@local.test`) is for local only — use **production** admin on this host.
- Do not commit real passwords into the Postman collection file.
