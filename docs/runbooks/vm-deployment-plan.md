# VM Deployment Plan

Date: 2026-07-08
Status: VM stack deployed; public DNS cutover pending

## Goal

Move `tuvisolutions.com` and `www.tuvisolutions.com` from the older Vercel site
to the VM at `170.64.154.143`, serving `apps/restaurant-services-catalog` as the
canonical Tuvi website. Do not deploy `presentation/` or make the older
`tuvi-website/app` the public homepage.

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
| Restaurant services website | `apps/restaurant-services-catalog` | Vite static build | Build exists; currently not in VM Compose |
| Tuvi corporate website | `tuvi-website/app` | Next.js app on `3001` | Legacy for this deployment; do not use as the public homepage |
| Restaurant demo template | `template` | Next.js app on `3000` | Runtime depends on main API and voice agent; currently not in VM Compose |
| Automation jobs | `automation/outreach` | Python scripts | One-shot/manual jobs; not a long-running web service |

Existing commands:

```bash
make up
make logs
make down
make voice-up
make voice-logs
make voice-down
make restaurant-services-catalog-build
make test
```

The local Compose stack starts only:

```text
postgres -> migrate -> api -> worker
voice-sales-redis -> voice-sales-agent
```

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
| `tuvisolutions.com` | catalog container at `127.0.0.1:15173` |
| `www.tuvisolutions.com` | catalog container at `127.0.0.1:15173` |
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
PUBLIC_WEB_URL=https://tuvisolutions.com
CORS_ALLOWED_ORIGINS=https://tuvisolutions.com,https://www.tuvisolutions.com,https://demo.tuvisolutions.com
DATABASE_URL=postgres://postgres:<password>@postgres:5432/restaurant_platform?sslmode=disable
REDIS_URL=redis://redis:6379
TOKEN_SECRET=<32+ chars>
DEMO_TOKEN_SECRET=<32+ chars>
TUVI_API_TOKEN=<32+ chars>
CONSULTATION_NOTIFY_EMAIL=<team email>
CONSULTATION_TIMEZONE=Australia/Sydney
CONSULTATION_GOOGLE_CALENDAR_DISABLED=true
EMAIL_PROVIDER=disabled
EMAIL_DISABLE_SENDING=true
```

Enable email only after confirming provider credentials and sender domain:

```text
EMAIL_PROVIDER=resend
EMAIL_API_KEY=<secret>
EMAIL_FROM_ADDRESS=<verified sender>
EMAIL_DISABLE_SENDING=false
```

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
REDIS_URL=redis://voice-sales-redis:6379
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

Do not deploy as the public homepage in this cutover.

### Restaurant Demo Template

```text
TEMPLATE=2
NEXT_PUBLIC_API_URL=https://api.tuvisolutions.com
NEXT_PUBLIC_VOICE_AGENT_URL=https://voice.tuvisolutions.com
VOICE_AGENT_URL=http://voice-sales-agent:8000
CALL_API_SECRET=<same value as voice agent>
```

## Compose Changes Needed

Use `infra/docker/docker-compose.vm.yml`. It defines:

1. `postgres`, `redis`, `migrate`, `api`, and `worker`.
2. `restaurant-services-catalog`, built from the Vite app and served by Nginx.
3. `template`, built as a Next.js production server.
4. `voice-agent`, built from `voice-sales-agent/`.
5. Only loopback host ports; public traffic stays on host-level Caddy.

Minimal service dependency graph:

```text
Caddy :80/:443
  -> 127.0.0.1:15173 restaurant-services-catalog
  -> 127.0.0.1:18080 api
  -> 127.0.0.1:18000 voice-agent
  -> 127.0.0.1:13000 template

api -> postgres, redis
worker -> postgres, redis
migrate -> postgres, redis
voice-agent -> redis, api
template -> api, voice-agent
```

## Deployment Sequence

1. Code checkout
   - Push the local `phase1_03/backend` deployment commit.
   - Clone or update `/opt/tuvi/MonoRepo` on `phase1_03/backend`.
   - Leave `/root/MonoRepo` untouched.

2. Secret setup
   - Create `/opt/tuvi/env/*.env`.
   - Do not commit or print secret values.
   - Ensure `TOKEN_SECRET`, `DEMO_TOKEN_SECRET`, `TUVI_API_TOKEN`, and
     `CALL_API_SECRET` are production values.

3. Build and start VM stack
   - Run `docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml up -d --build`.
   - Confirm API, worker, Postgres, Redis, catalog, template, and voice
     containers are running.

4. Reverse proxy and TLS
   - Append `infra/docker/Caddyfile.tuvi.example` routes to `/etc/caddy/Caddyfile`.
   - Run `caddy validate --config /etc/caddy/Caddyfile`.
   - Reload Caddy only after validation passes.

5. DNS cutover
   - Set `tuvisolutions.com`, `www.tuvisolutions.com`,
     `api.tuvisolutions.com`, `voice.tuvisolutions.com`, and
     `demo.tuvisolutions.com` to `170.64.154.143`.
   - Caddy can issue public certificates only after DNS points to this VM.

6. Twilio setup
   - Configure Twilio inbound voice webhook after HTTPS works:
     `POST https://voice.tuvisolutions.com/twiml`.

7. Seed data
   - Run `seed-admin` once if no admin exists.
   - Run `seed-demo-fixture` or import restaurant data only if needed.

8. Smoke checks
    - `https://tuvisolutions.com`
    - `https://api.tuvisolutions.com/api/public/v1/site/restaurants`
    - `https://voice.tuvisolutions.com/readyz/browser`
    - `https://demo.tuvisolutions.com`
    - Tuvi booking availability and booking POST through the website.
    - Voice callback form through website to voice agent.
    - Twilio inbound call to `/twiml`.

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
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml logs -f --tail=200 api worker voice-agent restaurant-services-catalog template
docker stats
df -h
```

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
- Put `CALL_API_SECRET` only in server-side env files.
- Keep `TUVI_API_TOKEN` server-side only.
- Restrict CORS to actual public domains.
- Keep email sending disabled until sender domain and review flow are approved.
- Keep outbound calls disabled or tightly gated until compliance rules are
  confirmed.

## Remaining External Steps

1. Change DNS records from Vercel to `170.64.154.143`.
2. Add real `voice.env` provider secrets for Twilio, Deepgram, Cartesia, and
   OpenAI before expecting `/readyz/browser` or calls to pass.
3. Enable email and Google Calendar only after those provider credentials are
   reviewed and configured.
