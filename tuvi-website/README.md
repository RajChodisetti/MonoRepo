# Tuvi Corporate Website

Marketing site for [Tuvi Solutions](https://www.tuvisolutions.com/) — custom software consultancy brand. Content sourced from the live site; design is a fresh animated dark-premium Next.js build.

## Quick start (one command)

From the **Tuvi** repo root:

```bash
./start-tuvi.sh
```

This starts Postgres, consultation API (`:8090`), voice agent Docker (`:8000`), and the website (`:3000`).

Stop background services:

```bash
./stop-tuvi.sh
```

## Quick start (website only)

```bash
cd tuvi-website/app
cp .env.example .env.local   # CONSULTATION_API_URL + TUVI_API_TOKEN
npm install
npm run dev
```

Open **http://localhost:3000**.

## Environment

| Variable | Purpose |
|----------|---------|
| `CONSULTATION_API_URL` | Go booking API (default `http://localhost:8090`) |
| `TUVI_API_TOKEN` | Bearer token for consultation API (server-only) |
| `NEXT_PUBLIC_CONTACT_EMAIL` | Footer & contact mailto |
| `NEXT_PUBLIC_LINKEDIN_URL` | Footer LinkedIn link |
| `NEXT_PUBLIC_VOICE_AGENT_URL` | Voice AI server (default `http://localhost:8000`) |

## Voice AI assistant

Bottom-right **Talk to Tuvi AI** uses `voice-sales-agent` with `?agent=corporate`.

**Recommended:** `./start-tuvi.sh` from the Tuvi repo root (starts Postgres, API, voice agent, and website).

Manual split terminals:

```bash
docker compose -f MonoRepo/infra/docker/docker-compose.yml up -d postgres
cd tuvi-website/backend && make migrate-up && make run
cd voice-sales-agent && docker compose up -d --build
cd tuvi-website/app && npm run dev
```

Set `TUVI_WEBSITE_API_URL` and `TUVI_API_TOKEN` in `voice-sales-agent/.env` (see `backend/README.md`).

**Book a Call** (`/book`) uses Google Calendar + SMTP via `tuvi-website/backend/`. Voice AI bookings use the same API.

## Structure

```
app/              # Next.js frontend
backend/          # Go consultation API (slots, GCal, SMTP)
```

## Commands

```bash
npm run dev      # dev server :3000
npm run build    # production build
npm run start    # serve build :3000
```

## Notes

- Leadership section shows **names only** (initials monogram) — no founder photos.
- Standalone frontend; no MonoRepo backend dependency.
