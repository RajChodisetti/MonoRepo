# Session Summary

**Latest (2026-07-06):** Tuvi scheduler and corporate voice bookings now call the main MonoRepo company consultation API for availability, slot checks, and booking creation.

**Endpoint state:** `GET /api/v1/company/consultations/availability`, `GET /api/v1/company/consultations/availability/check`, and `POST /api/v1/company/consultations` are shared by the web scheduler and corporate voice agent; source is recorded as `web` or `voice`.

**Config needed:** Set the same `TUVI_API_TOKEN` in the main API, Tuvi website server env, and corporate voice-agent env.

**Verification:** Backend tests, Python syntax compile, Tuvi TypeScript check, Node 22 Next build, and normal Tuvi `npm run build` all passed.
