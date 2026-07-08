# Session Summary

**Latest (2026-07-08):** Tuvi VM stack is deployed on `root@170.64.154.143` from `/opt/tuvi/MonoRepo`; Caddy routes for Tuvi domains are installed and validated.

**Deployment state:** Catalog (`15173`), API (`18080`), voice (`18000`), template (`13000`), Postgres, Redis, worker, and migrations are running under Compose project `tuvi`.

**Public cutover:** DNS still points `tuvisolutions.com`/`www` at Vercel and Tuvi subdomains are missing; update DNS to `170.64.154.143` for public HTTPS.

**Voice state:** Voice service is running, but `/readyz/browser` reports missing Deepgram/OpenAI/Cartesia keys in `/opt/tuvi/env/voice.env`.

**Git state:** Local branch has deployment commits ahead of origin; GitHub HTTPS push failed with RPC 400 and SSH push is not authorized, so VM deploy used rsync.
