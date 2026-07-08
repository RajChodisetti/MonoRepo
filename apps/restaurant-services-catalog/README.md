# Tuvi Restaurant Growth Website

Primary Vite website for Tuvi Solutions and the restaurant services package.
This is the merged replacement for the older Tuvi corporate/presentation site
and the restaurant services catalog.

It includes:

- Tuvi company positioning, proof points, and guarantee
- Restaurant services navigation with the current `Restaurant` service category
- Demo websites, QR ordering, rewards, AI voice receptionists, reservations,
  outreach, and content automation
- A safe static contact form that opens a prefilled email request

## Start Locally

From the repo root:

```bash
make restaurant-services-catalog-dev
```

Or from this app directory:

```bash
npm install
npm run dev
```

Open http://127.0.0.1:5173.

## Public Configuration

Copy `.env.example` to `.env.local` only if you need to override public contact
links:

```bash
cp apps/restaurant-services-catalog/.env.example apps/restaurant-services-catalog/.env.local
```

Only `VITE_` variables are read by this static app. Do **not** put
`TUVI_API_TOKEN` or other secrets in this app; those belong behind the backend or
a server-side proxy.

## Notes

This merged website is self-contained and does not depend on the outdated Tuvi
presentation website. The Tuvi brand link returns to the top of the merged site,
the Services navigation exposes Restaurant services, and booking CTAs route to
the local contact section.
