# VM Deployment Plan

Date: 2026-07-14
Status: VM stack deployed; durable scrape/OCR/outreach workflow pending controlled rollout

## Goal

Serve `tuvisolutions.com` and `www.tuvisolutions.com` from the VM at
`170.64.154.143`, using `tuvi-website/app` as the canonical corporate website.
Keep the standalone restaurant services catalog deployed on loopback for
internal verification while `/services/restaurants` is served by the corporate
website. Do not deploy `presentation/`.

## Current Repo Setup Found

The repo already has these deployable surfaces:

| Surface | Path | Runtime | Current state |
| --- | --- | --- | --- |
| Main API | `backend/cmd/api` | Go binary in Docker | Containerized through `infra/docker/Dockerfile.backend` and Compose profile `stack` |
| Worker | `backend/cmd/worker` | Go binary in Docker | Containerized through same backend image |
| Migrations | `backend/cmd/migrate` | Go binary in Docker | One-shot `migrate` service in Compose profile `stack` |
| PostgreSQL | `infra/docker/docker-compose.yml` | `postgres:16-alpine` | Persistent Docker volume `postgres_data` |
| Voice agent | `voice-sales-agent/` | Python/FastAPI in Docker | Containerized through Compose profile `voice` with Redis |
| Voice Redis | Compose service | `redis:7-alpine` | Persistent Docker volume `voice_sales_redis_data` |
| Restaurant services website | `apps/restaurant-services-catalog` | Vite static build | Containerized and exposed on VM loopback |
| Tuvi corporate website | `tuvi-website/app` | Next.js app | Canonical public website, containerized on VM loopback |
| Restaurant demo template | `template` | Next.js app on `3000` | Containerized in VM Compose and exposed only on VM loopback |
| City scrape worker | `automation/outreach/city_scrape_worker.py` | Python worker in Docker | Long-running, database-backed poller with no public port |
| OCR worker/job | `automation/outreach/verify_leads_from_db.py` | Background worker plus one-shot diagnostic job | `ocr-worker` polls durable email-only work; `ocr-job` remains in the Compose `jobs` profile |

Existing commands:

```bash
make up
make logs
make down
make voice-up
make voice-logs
make voice-down
make restaurant-services-catalog-build
make ocr-job
make test
```

The application Compose stack starts:

```text
postgres, redis -> migrate -> api -> worker
postgres, api -> scrape-worker
postgres -> ocr-worker
redis, api -> voice-agent
```

`ocr-worker` is the default long-running scheduler. `ocr-job` remains one-shot
for controlled diagnosis through the `jobs` profile; do not schedule both.

The VM deployment adds `infra/docker/docker-compose.vm.yml`,
`Dockerfile.catalog`, `Dockerfile.template`, and `Caddyfile.tuvi.example`.

## VM Audit Results

- VM: DigitalOcean Ubuntu 24.04 droplet, `ubuntu-s-1vcpu-2gb-syd1`.
- Docker: installed; Docker Compose plugin installed.
- Caddy: host-level systemd service already owns ports `80` and `443`.
- Existing services to preserve:
  - `tilnest.com`, `www.tilnest.com` -> `127.0.0.1:3001`
  - `api.sustainabilitywise.com.au` -> `127.0.0.1:3000`
  - `/insta/*` direct-IP route -> `127.0.0.1:7780`
  - n8n/Instagram Docker stack under `/opt/n8n-insta`
- Existing `/root/MonoRepo` is stale on `master`; deploy from a fresh checkout
  under `/opt/tuvi/MonoRepo`.
- Open loopback ports for Tuvi: `15173`, `18080`, `18000`, and `13000`.
- Current DNS before cutover:
  - `tuvisolutions.com` -> Vercel IP
  - `www.tuvisolutions.com` -> Vercel CNAME/IPs
  - `api.tuvisolutions.com`, `voice.tuvisolutions.com`, and
    `demo.tuvisolutions.com` have no records.
- Deployed Tuvi loopback services:
  - catalog: `127.0.0.1:15173`
  - corporate website: `127.0.0.1:15174`
  - API: `127.0.0.1:18080`
  - voice: `127.0.0.1:18000`
  - demo template: `127.0.0.1:13000`
- Caddy routes for the Tuvi domains are installed and validated; they will serve
  public HTTPS after DNS points to the VM.

## Recommended VM Layout

Use one application directory:

```text
/opt/tuvi/
  MonoRepo/                         # git checkout
  env/
    stack.env                       # backend/api/worker/migrations secrets
    voice.env                       # Twilio/STT/TTS/LLM/call secrets
    template.env                    # Next server env, if deployed
  backups/
    postgres/
  logs/
```

Keep database and Redis data in Docker named volumes. Do not expose Postgres,
Redis, API, voice, template, or catalog directly; only Caddy is public.

## Public Routing

Use the existing host-level Caddy service on ports `80` and `443`.

Recommended hostnames:

| Hostname | Target |
| --- | --- |
| `tuvisolutions.com` | corporate website at `127.0.0.1:15174` |
| `www.tuvisolutions.com` | corporate website at `127.0.0.1:15174` |
| `api.tuvisolutions.com` | Go API container at `127.0.0.1:18080` |
| `voice.tuvisolutions.com` | voice agent container at `127.0.0.1:18000` |
| `demo.tuvisolutions.com` | template container at `127.0.0.1:13000` |

Subdomains are cleaner because the voice agent needs public WebSocket support
for Twilio Media Streams.

## Environment Required

### Main API / Worker / Migrate

Store in `/opt/tuvi/env/stack.env`:

```text
APP_ENV=production
APP_NAME=restaurant-platform
HTTP_ADDR=:8080
PUBLIC_BASE_URL=https://api.tuvisolutions.com
PUBLIC_WEB_URL=https://demo.tuvisolutions.com
PUBLIC_MARKETING_URL=https://tuvisolutions.com
PRESENTATION_SITE_URL=https://tuvisolutions.com/services/restaurants
CORS_ALLOWED_ORIGINS=https://tuvisolutions.com,https://www.tuvisolutions.com,https://demo.tuvisolutions.com
POSTGRES_USER=tuvi
POSTGRES_PASSWORD=<strong password>
POSTGRES_DB=restaurant_platform
DATABASE_URL=postgres://tuvi:<URL-encoded same password>@postgres:5432/restaurant_platform?sslmode=disable
REDIS_URL=redis://redis:6379
TOKEN_SECRET=<32+ chars>
DEMO_TOKEN_TTL=720h
TUVI_API_TOKEN=<32+ chars>
CALL_API_SECRET=<same server-side value used by voice/template services>
CONSULTATION_NOTIFY_EMAIL=<team email>
CONSULTATION_TIMEZONE=Australia/Sydney
# PostgreSQL is the active slot ledger; keep Calendar disabled for this release.
CONSULTATION_GOOGLE_CALENDAR_DISABLED=true
EMAIL_PROVIDER=disabled
EMAIL_DISABLE_SENDING=true

OUTREACH_BULK_MAX=150
OUTREACH_EMAILS_PER_ACCOUNT=40
OUTREACH_EMAIL_COOLDOWN=24h
OUTREACH_SEND_WINDOW=8h
OUTREACH_SEND_JITTER_MIN=2m
OUTREACH_SEND_JITTER_MAX=5m
OUTREACH_ZOHO_ACCOUNTS_JSON=[]
OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON=[]
```

Keep scrape/OCR credentials in a separate host-readable file
(`/opt/tuvi/env/ingestion.env`, mode `0600`). The Compose services read this file
directly and use the same container-network `DATABASE_URL` as the API:

```text
GOOGLE_PLACES_API_KEY=<restricted Places API key>
PLACES_API_BASE_URL=https://places.googleapis.com/v1
APOLLO_API_KEY=<restricted Apollo API key>
APOLLO_API_BASE_URL=https://api.apollo.io/api/v1
APOLLO_ENRICHMENT_ENABLED=true

SCRAPE_WORKER_POLL_SECONDS=15
SCRAPE_JOB_LEASE_SECONDS=900
SCRAPE_INITIAL_GRID_ROWS=4
SCRAPE_INITIAL_GRID_COLUMNS=4
SCRAPE_CELL_PAGE_LIMIT=3
SCRAPE_GRID_MAX_DEPTH=12

LEAD_OCR_VERIFICATION_ENABLED=false
LEAD_OCR_BATCH_SIZE=50
LEAD_OCR_DAILY_REQUEST_LIMIT=200
OCR_WORKER_POLL_SECONDS=900
MENU_OCR_TIMEOUT=45
HUGGING_FACE_API_KEY=<vision provider key>
HF_VISION_MODEL=<supported vision model>
```

Keep OCR false during migration. Run the first reviewed one-shot with
`-e LEAD_OCR_VERIFICATION_ENABLED=true`; after that batch is accepted, set the
ingestion file to true and start the `ocr-worker` background service. Remove any
older OCR cron so there is only one scheduler.

The durable scrape worker runs Google Places first, then Apollo only for leads
still missing an owner or work email and only when a usable business domain is
available. Request reservations are persisted and capped at 500 combined
Places/Apollo calls per window. Grid cell, Places page token, candidate state,
request count, and `resume_at` survive process or VM restarts.

Before starting `scrape-worker`, explicitly remove or disable every legacy host
cron entry that invokes `automation/outreach/cron_lead_ingestion.sh`,
`daily_ingestion.py`, or `make ingest-daily`. Running either legacy path beside
the durable worker would bypass the shared scrape-job ledger and can duplicate
provider usage and lead ingestion.

Enable email only after confirming provider credentials and sender domain:

```text
EMAIL_PROVIDER=resend
EMAIL_API_KEY=<secret>
EMAIL_API_BASE_URL=https://api.resend.com
EMAIL_FROM_ADDRESS=<verified sender>
EMAIL_DISABLE_SENDING=false
```

For quota-managed bulk outreach with Google Workspace, first enable the Gmail
API in the Google Cloud project. Configure an OAuth consent screen/application,
request only `https://www.googleapis.com/auth/gmail.send`, obtain offline consent
and a refresh token separately for each mailbox, then configure:

```text
EMAIL_PROVIDER=disabled
EMAIL_FROM_ADDRESS=
EMAIL_FROM_NAME=Tuvi Solutions
OUTREACH_ZOHO_ACCOUNTS_JSON=[]
OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON=[{"key":"workspace-sales-1","mailbox_email":"sales1@example.com","from_email":"sales1@example.com","client_id":"<oauth client id>","client_secret":"<secret>","refresh_token":"<offline refresh token>"}]
EMAIL_DISABLE_SENDING=false
```

`mailbox_email` must be the primary mailbox that authorized the refresh token.
Use a different `from_email` only for a send-as alias already configured in
Gmail. Gmail delivery uses the HTTPS `users.messages.send` API; SMTP is not used.

Zoho remains an optional quota-managed provider when explicitly configured:

```text
EMAIL_PROVIDER=zoho
EMAIL_FROM_ADDRESS=<verified fallback sender>
ZOHO_ACCOUNT_ID=<default account id>
ZOHO_FROM_EMAIL=<verified sender>
ZOHO_REGION=com.au
ZOHO_CLIENT_ID=<secret>
ZOHO_CLIENT_SECRET=<secret>
ZOHO_REFRESH_TOKEN=<secret>
OUTREACH_ZOHO_ACCOUNTS_JSON=[{"key":"sales-au-1","account_id":"<id>","from_email":"<verified sender>","client_id":"<secret>","client_secret":"<secret>","refresh_token":"<secret>","region":"com.au"}]
EMAIL_DISABLE_SENDING=false
```

The singleton `ZOHO_*` values configure the generic Zoho email adapter; the
JSON array configures the independently rotated, quota-managed outreach pool.
The stable non-secret account `key` identifies the PostgreSQL quota row across
credential rotations. PostgreSQL conservatively reserves 40 delivery attempts
per account cycle. It spreads those reservations across 40 persisted slots in
an 8-hour window (about 12 minutes per slot) and adds a random persisted 2-5
minute offset. On-time cycles normally produce effective gaps of about 9-15
minutes. A global 2-5 minute minimum-delay guard also spans account changes and
prevents restart catch-up bursts. Each durable job activation makes at most one
provider attempt, releases the worker, and resumes at the database
`available_at`; it does not sleep in a worker process. After slot 40, that
account enters its 24-hour cooldown and the workflow rotates or waits
automatically.

Forty is an upper bound on reserved attempts, not a guarantee of 40 accepted or
delivered emails. A shortage of fully approved eligible leads, provider
rejections, or ambiguous outcomes can produce fewer confirmed sends. Confirmed
deliveries increment `email_send_count` and record the global
`last_email_send_sequence`; ambiguous provider outcomes use campaign status
`send_unknown`, consume their reserved slot, and are not retried automatically.

Only HTTP(S) provider APIs are supported. `EMAIL_PROVIDER=smtp` is rejected at
configuration validation. Keep sending disabled until the sender identities,
human approval workflow, and all four canonical production URLs above have been
reviewed.

### Voice Agent

Store in `/opt/tuvi/env/voice.env`:

```text
ENVIRONMENT=production
PUBLIC_BASE_URL=https://voice.tuvisolutions.com
MONOREPO_API_URL=http://api:8080
TUVI_API_TOKEN=<same value as stack.env>
CALL_API_SECRET=<same value used by Next apps>
TWILIO_ACCOUNT_SID=<secret>
TWILIO_AUTH_TOKEN=<secret>
TWILIO_PHONE_NUMBER=<E.164 number>
DEEPGRAM_API_KEY=<secret>
OPENAI_API_KEY=<secret>
CARTESIA_API_KEY=<secret>
REDIS_URL=redis://redis:6379
CALL_LOG_DB=/app/data/calls.db
```

Twilio phone number webhook:

```text
POST https://voice.tuvisolutions.com/twiml
```

### Restaurant Services Catalog

This is a static site. It must not contain secrets. Only `VITE_*` public values
may be used. Build with:

```bash
npm --prefix apps/restaurant-services-catalog ci
npm --prefix apps/restaurant-services-catalog run build
```

Serve `apps/restaurant-services-catalog/dist`.

### Tuvi Corporate Website

Build `tuvi-website/app` with `infra/docker/Dockerfile.tuvi-website` and serve
the production Next.js process on VM loopback port `15174`.

### Restaurant Demo Template

```text
TEMPLATE=2
NEXT_PUBLIC_API_URL=https://api.tuvisolutions.com
NEXT_PUBLIC_VOICE_AGENT_URL=https://voice.tuvisolutions.com
VOICE_AGENT_URL=http://voice-agent:8000
CALL_API_SECRET=<same value as voice agent>
```

## VM Compose Service Inventory

Use `infra/docker/docker-compose.vm.yml`. It defines:

1. `postgres`, `redis`, `migrate`, `api`, and the Go `worker`.
2. Long-running `scrape-worker`, which uses the API database and exposes no port.
3. Long-running `ocr-worker`, plus one-shot `ocr-job` in the `jobs` profile for diagnosis.
4. `restaurant-services-catalog`, built from the Vite app and served by Nginx.
5. `tuvi-website`, built as the canonical corporate Next.js site.
6. `template`, built as a Next.js production server.
7. `voice-agent`, built from `voice-sales-agent/`.
8. Only loopback host ports; public traffic stays on host-level Caddy.

Minimal service dependency graph:

```text
Caddy :80/:443
  -> 127.0.0.1:15174 tuvi-website
  -> 127.0.0.1:18080 api
  -> 127.0.0.1:18000 voice-agent
  -> 127.0.0.1:13000 template

api -> postgres, redis
worker -> postgres, redis
migrate -> postgres, redis
scrape-worker -> postgres, api
ocr-worker -> postgres -> lead.prepare jobs -> worker
voice-agent -> redis, api
template -> api, voice-agent
```

## Deployment Sequence

1. Code checkout
   - Push the local `phase1_03/backend` deployment commit.
   - Clone or update `/opt/tuvi/MonoRepo` on `phase1_03/backend`.
   - Leave `/root/MonoRepo` untouched.

2. Secret setup
   - Create `/opt/tuvi/env/stack.env`, `/opt/tuvi/env/ingestion.env`, and the
     other required `/opt/tuvi/env/*.env` files with mode `0600`.
   - Do not commit or print secret values.
   - Ensure `TOKEN_SECRET`, `TUVI_API_TOKEN`, and
     `CALL_API_SECRET` are production values.

3. Quiesce, migrate, and start the updated VM stack
   - Follow the exact controlled-rollout order in
     `docs/runbooks/lead-scrape-ocr-outreach.md`; keep OCR and email disabled.
   - Remove/disable crontab entries for `cron_lead_ingestion.sh`,
     `daily_ingestion.py`, and `make ingest-daily` before starting the durable worker.
   - Stop the old API, Go worker, scrape worker, and any active OCR process before
     migration `000020`; a pre-lease worker must not race the schema/state
     reconciliation. Keep PostgreSQL and Redis running.
   - Back up PostgreSQL and inventory the ambiguous states listed in the detailed
     workflow runbook. Inventory `email_delivery_attempts` only if that table
     already exists from a partial rollout.
   - Run the separate build, database start, one-shot migration, API start, Go
     worker start, and scrape-worker start commands from the detailed workflow
     runbook. Do not replace that sequence with a broad `up -d --build` during
     the `000020`/`000021` rollout.
   - After the updated core processes are running, build/start the catalog,
     corporate website, template, and voice services; these do not replace the
     controlled migration sequence.
   - Confirm API, Go worker, scrape worker, Postgres, Redis, corporate website,
     catalog, template, and voice containers are running.
   - Confirm the scrape-worker logs show polling without credential, schema, or
     lease errors. Do not run the legacy ingestion scripts in parallel.

4. Start background OCR
   - Run one controlled batch and inspect its state transitions and generated drafts:

     ```bash
     docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
       -f infra/docker/docker-compose.vm.yml --profile jobs run --rm \
       -e LEAD_OCR_VERIFICATION_ENABLED=true ocr-job
     ```

   - Inspect that batch. Only after accepting its states and generated drafts,
     set `LEAD_OCR_VERIFICATION_ENABLED=true` in `ingestion.env`, remove any
     older OCR cron, and start the background worker:

     ```bash
     docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
       -f infra/docker/docker-compose.vm.yml up -d --no-deps ocr-worker
     ```

   - OCR states are `pending`, `running`, `verified`, `no_images`, and `failed`.
     Only `verified` enqueues `lead.prepare` and is eligible for later human
     approval. `no_images` must never be treated as verified. Only restaurants
     with a canonical email are claimed. `ocr_daily_request_usage` enforces the
     global 200-request UTC-day ceiling across restarts; timeout/429 claims are
     returned to pending without consuming an attempt.

5. Reverse proxy and TLS
   - Append `infra/docker/Caddyfile.tuvi.example` routes to `/etc/caddy/Caddyfile`.
   - Run `caddy validate --config /etc/caddy/Caddyfile`.
   - Reload Caddy only after validation passes.

6. DNS cutover
   - Set `tuvisolutions.com`, `www.tuvisolutions.com`,
     `api.tuvisolutions.com`, `voice.tuvisolutions.com`, and
     `demo.tuvisolutions.com` to `170.64.154.143`.
   - Caddy can issue public certificates only after DNS points to this VM.

7. Twilio setup
   - Configure Twilio inbound voice webhook after HTTPS works:
     `POST https://voice.tuvisolutions.com/twiml`.

8. Seed data
   - Run `seed-admin` once if no admin exists.
   - Run `seed-demo-fixture` or import restaurant data only if needed.

9. Smoke checks
    - `https://tuvisolutions.com`
    - `https://api.tuvisolutions.com/api/public/v1/site/restaurants`
    - `https://voice.tuvisolutions.com/readyz/browser`
    - `https://demo.tuvisolutions.com`
    - Tuvi booking availability and booking POST through the website.
    - Voice callback form through website to voice agent.
    - Twilio inbound call to `/twiml`.

## Operate the Lead Workflow

All workflow endpoints require an `internal_admin` bearer token. Triggering is
idempotent while the city/niche pair already has an active job:

```bash
export TUVI_API=https://api.tuvisolutions.com
export ADMIN_TOKEN=<short-lived internal-admin JWT>

curl -fsS -X POST "$TUVI_API/api/v1/scrape-jobs" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data '{"city":"Melbourne","niche":"restaurant"}'

curl -fsS "$TUVI_API/api/v1/scrape-jobs?limit=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -fsS "$TUVI_API/api/v1/scrape-jobs/<scrape-job-uuid>" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

The worker persists each grid cell, Places page token, candidate, current request
window, and resume time. `waiting` with `request_limit` resumes after 24 hours;
`waiting` with `revisit` begins another city coverage cycle after 24 hours so new
Place IDs can be discovered without duplicating already imported Place IDs.

After OCR has reached `verified` and the generated demo/campaign drafts have been
reviewed, record each required human gate separately. Use the exact
profile-preview → version-bound profile review → demo-preview → version-bound
publication → campaign-detail → version-bound campaign approval commands in
`docs/runbooks/lead-scrape-ocr-outreach.md`; unversioned approval requests are
rejected.

Generic `POST /api/v1/restaurants/{id}/demo-sites` creation is draft-only;
publishing through its generic create request is rejected. Profile review records
`reviewed_at/by`, demo publication records `published_at/by`, and campaign
approval records `approved_at/by`. Bulk eligibility requires all three audited,
still-current gates plus OCR `verified` status. A new profile decision returns
older demo/campaign approvals to draft. Only a draft campaign can receive a new
approval; `send_unknown` cannot be re-approved.

Keep `EMAIL_DISABLE_SENDING=true` until all content and links are approved. Then
restart the API and Go worker after changing it to `false`, and explicitly start
one bulk workflow:

```bash
curl -fsS -X POST "$TUVI_API/api/v1/outreach/bulk-send" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -fsS "$TUVI_API/api/v1/outreach/bulk-send/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

The obsolete `/campaigns/{id}/send-step` route is not registered, so outreach
can only use the durable, paced 40-attempt/account quota and 24-hour continuation
path. Keep `EMAIL_DISABLE_SENDING=true` until the sender identities, content,
links, and human approval workflow have been reviewed.

If a campaign contains stale links or its token-gated demo has expired, first set
the demo back to `draft`, then call
`POST /api/v1/campaigns/{id}/regenerate`. This rotates the opaque demo
token/expiry, renders current links, records the administrator, and clears
campaign approval; publish and approve again before outreach.

## Backup And Rollback

Backups:

```bash
cd /opt/tuvi/MonoRepo
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml exec -T postgres \
  pg_dump -U "${POSTGRES_USER:-tuvi}" "${POSTGRES_DB:-restaurant_platform}" \
  > /opt/tuvi/backups/postgres/restaurant_platform-$(date +%F-%H%M).sql
```

Schedule daily backups and keep at least 7 daily and 4 weekly copies.

Rollback:

1. Keep the previously deployed git commit recorded in `/opt/tuvi/current-release`.
2. Before deploy, take a Postgres backup.
3. If app deploy fails, check out previous commit and rerun Compose.
4. If migration fails, do not start new API; restore from backup or run the
   matching down migration only after reviewing data impact.

## Observability

Minimum:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml ps
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml logs -f --tail=200 api worker scrape-worker voice-agent restaurant-services-catalog template
docker stats
df -h
```

OCR is one-shot, so inspect `/opt/tuvi/logs/ocr.log` (or the host timer journal)
and alert if scheduled runs stop completing. Monitor `scrape_jobs.resume_at`,
`last_error`, failed cells, and `outreach_email_accounts.available_at` in
PostgreSQL in addition to container health.

Add later:

- Uptime checks for website, API, voice readiness, and demo template.
- Log rotation for Docker.
- Disk alert for Postgres volume.
- Backup success alert.

## Security Checklist

- Do not expose Postgres or Redis publicly.
- Do not commit `stack.env`, `voice.env`, `template.env`, `.env.local`, or provider
  credentials.
- Set production secrets before `APP_ENV=production`.
- Use HTTPS for all public sites.
- Put `CALL_API_SECRET` only in server-side env files and keep its value aligned
  across `stack.env`, `voice.env`, and `template.env`.
- Keep `TUVI_API_TOKEN` server-side only.
- Restrict CORS to actual public domains.
- Keep email sending disabled until sender domain and review flow are approved.
- Require audited profile approval, demo publication, and campaign approval for
  every real outreach recipient; `no_images`, `failed`, and `send_unknown` are
  not automatically sendable/retryable states.
- Keep `PUBLIC_BASE_URL`, `PUBLIC_WEB_URL`, `PUBLIC_MARKETING_URL`, and
  `PRESENTATION_SITE_URL` on the canonical HTTPS production values documented
  above. Production startup/render validation must fail before a provider call
  if a link is HTTP, local, unresolved, or on an unapproved host.
- Do not configure SMTP; use only supported HTTP(S) provider APIs.
- Keep legacy daily ingestion cron disabled while `scrape-worker` is active.
- Keep outbound calls disabled or tightly gated until compliance rules are
  confirmed.

## Remaining External Steps

1. Change DNS records from Vercel to `170.64.154.143`.
2. Add real `voice.env` provider secrets for Twilio, Deepgram, Cartesia, and
   OpenAI before expecting `/readyz/browser` or calls to pass.
3. Configure the OCR provider and install the locked one-shot schedule.
4. Confirm legacy ingestion cron entries are removed, then trigger the first
   city job through `POST /api/v1/scrape-jobs`.
5. Enable email and Google Calendar only after the Gmail API/OAuth mailbox credentials, canonical
   HTTPS links, and the audited human approval flow are reviewed and configured.
