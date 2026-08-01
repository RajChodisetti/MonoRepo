# ADR: Hybrid Live and Durable Restaurant Media

Date: 2026-07-19
Status: Accepted

## Context

Restaurant demo pages lost their photos because Google Places Photo Media URLs
were correctly treated as temporary, while import replacement deleted existing
image rows before discovering there was no persistent replacement URL. A fresh
Places resolver existed only in the admin API. Copying Google photo bytes into
our own object storage would create an unnecessary licensing and attribution
risk, while relying on temporary provider URLs for restaurant-owned media would
leave published customer sites unstable.

Menu-document scans are useful for extracting structured menu text, but the
product must not display those scans on public restaurant websites.

## Decision

Use two source-aware media paths:

1. Google Places photos are resolved on demand by the Go API from the stable
   restaurant Place ID. Photo names, resolved URLs, and bytes are not persisted
   or copied. Public responses are `no-store`, images bypass the Next.js image
   optimizer, and Google Maps/source/author/report attribution is rendered with
   each displayed photo. Concurrent requests for the same restaurant share one
   in-flight resolver call without caching its result.
2. Restaurant-owned or separately licensed uploads are stored in an
   S3-compatible bucket/CDN and recorded in `restaurant_media_assets`, including
   explicit rights state. Uploads remain draft until the background vision
   worker classifies them. Approved assets use stable public URLs and normal
   Next.js optimization.

Admin-opened generated-site previews may request `preview_media=google_live`.
This is a no-store, attributed, live-only fallback used only when reviewed media
is empty, so lead-table previews can show current restaurant photos before the
OCR budget catches up. Published and token-gated demos continue to use the
reviewed public-media path.

OCR stores a one-way SHA-256 fingerprint derived from the Place ID and the
in-memory photo resource name. Freshly resolved Google photos must match that
fingerprint exactly; the resource name itself is never persisted or exposed.
Older positional-only classifications return to pending for a safe OCR refresh.
Unmatched, unclassified, and low-confidence Google photos fail closed and are
not displayed. Text-heavy `other` classifications also fail closed as a
defense against uncertain menu scans. A
`menu_document` classification is excluded at the service, public API, payload,
adapter, and template boundaries. Durable uploads detected as menu documents
are rejected from public use but retained for authenticated administrative
inspection.

For durable assets, OCR stores factual caption and alt text, tags, technical
quality and hero suitability scores, orientation, subject position, and content
flags. Templates use placement/type metadata to select hero images and provide
Food & drink, Space & atmosphere, and More gallery filters. Menu-document images
are never used for heroes, galleries, dish cards, or footer mosaics.

## Options Considered

- Persist current Google Photo Media URLs: rejected because the URLs expire and
  provider policy requires live handling and attribution.
- Copy all Google images into S3: rejected because technical stability does not
  establish redistribution rights.
- Resolve every image only from Google: useful for short-lived prospect demos,
  but unsuitable as the sole source for durable published customer sites.
- Store only owner/licensed media: legally and technically strongest for live
  customer sites, but would leave new prospect demos without visual context.

## Consequences

- Prospect demos regain photos after an exact OCR resource match exists, without
  storing temporary Google identifiers, resource names, URLs, or bytes.
- Lead-table generated previews can display attributed live Google photos before
  reviewed OCR media exists, at the cost of live Places media requests.
- Published sites can move toward durable owner/licensed media while retaining a
  compliant live fallback for reviewed demos.
- Object storage must expose uploaded objects through the configured HTTPS CDN
  or bucket policy; credentials remain server-side.
- Each live demo request consumes one Places Details call plus media resolution
  calls. Provider errors degrade to owned media or an empty visual state rather
  than failing the page.
- Owner upload OCR shares the existing PostgreSQL-enforced 200-request UTC-day
  ceiling with scraped-photo OCR.

## Rollback / Revisit Trigger

Rollback migration `000031`, disable `STORAGE_PROVIDER`, and remove the media
service injection to return to the legacy persisted gallery behavior. Revisit
when Google licensing terms change, live resolver cost/latency becomes material,
or the product obtains a separate licensed media feed suitable for durable
storage.
