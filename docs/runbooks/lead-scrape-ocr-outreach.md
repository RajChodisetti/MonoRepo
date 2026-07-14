# Lead scraping, OCR, and outreach runbook

Date: 2026-07-14

This runbook operates the Places-first city workflow, targeted Apollo contact
enrichment, OCR verification, human review, and quota-managed invite email.
All private API calls require an `internal_admin` bearer token.

## Workflow and safety gates

```text
POST city scrape job
  -> Places grid discovery and Place-ID deduplication
  -> Places details (no arbitrary restaurant-website fetch in the durable worker)
  -> Apollo only for missing owner or work-email details
  -> PostgreSQL restaurant/profile import (OCR pending)
  -> scheduled OCR job
       -> no_images / failed (stop; not eligible)
       -> verified
            -> Go worker creates demo draft + campaign draft
            -> human approves profile
            -> human publishes demo
            -> human approves campaign
            -> human starts bulk outreach
                 -> 40 reserved delivery attempts per email account
                 -> rotate accounts
                 -> 24-hour PostgreSQL cooldown and automatic continuation
                 -> confirmed send updates the delivery ledger and restaurant counter
```

Creating a scrape job, running OCR, or preparing drafts does not send email.
Real email still requires all three record approvals, email sending enabled, and
an internal administrator starting the bulk workflow.

## Required deployment order

1. Keep `LEAD_OCR_VERIFICATION_ENABLED=false` and
   `EMAIL_DISABLE_SENDING=true` during rollout.
2. Remove or disable every legacy `cron_lead_ingestion.sh` /
   `daily_ingestion.py` crontab entry so it cannot race the durable request
   ledger during or after rollout.
3. Quiesce writes before migration: stop the old API, Go worker, and any scrape
   or OCR process. In particular, a pre-`000020` worker must not remain active
   while lease columns and ambiguous delivery states are reconciled.
4. Back up PostgreSQL. Before migrating, inventory `job_runs` in
   `queued`/`running` and campaigns in `sending`. If `email_delivery_attempts`
   already exists from a partial rollout, also inventory its `sending` rows.
   These ambiguous legacy states are deliberately reconciled fail-closed by
   `000020`.
5. Deploy/build the updated API, Go worker, scrape-worker, and OCR-job images,
   but do not start their application processes yet.
6. Apply migrations `000015` through `000023`, in order. `000019` depends on
   the review-audit columns introduced by `000018`; `000021` binds tracking and
   delivery audit rows to the immutable address actually used for outreach.
   `000022` returns legacy automatic artifacts to draft, queues fresh preparation,
   and binds new artifacts to both OCR input and the exact identity/public payload.
   `000023` makes maximum-depth coverage gaps a durable, visible waiting state.
7. Start the updated API and Go worker before enabling OCR. The updated Go worker must
   know the `lead.prepare` job type before OCR can enqueue it.
   Migration `000018` resets pre-audit profile approvals and demo publications
   to `draft`, because no trustworthy reviewer/publisher identity existed; record
   those human gates again through the APIs below.
   Migration `000022` likewise requires fresh review of pre-provenance automatic
   drafts; do not manually restore their old approval/publication state.
8. Start the updated `scrape-worker` service and confirm it is polling.
9. Run one small OCR batch and inspect the generated drafts.
10. Configure and review production email links and HTTP-provider credentials.
11. Enable email only after the approval and rendered-email preflight checks have
   been exercised with `EMAIL_REDIRECT_TO` in a non-production environment.

## Environment

In `/opt/tuvi/env/stack.env`:

```text
APP_ENV=production
POSTGRES_USER=tuvi
POSTGRES_PASSWORD=<strong password>
POSTGRES_DB=restaurant_platform
DATABASE_URL=postgres://tuvi:<URL-encoded same password>@postgres:5432/restaurant_platform?sslmode=disable
TOKEN_SECRET=<32+ character authentication secret>
DEMO_TOKEN_TTL=720h
TUVI_API_TOKEN=<32+ character server-side secret>
CALL_API_SECRET=<32+ character server-side secret shared with voice/template services>
CORS_ALLOWED_ORIGINS=https://tuvisolutions.com,https://www.tuvisolutions.com,https://demo.tuvisolutions.com

PUBLIC_BASE_URL=https://api.tuvisolutions.com
PUBLIC_WEB_URL=https://demo.tuvisolutions.com
PUBLIC_MARKETING_URL=https://tuvisolutions.com
PRESENTATION_SITE_URL=https://tuvisolutions.com/services/restaurants

EMAIL_PROVIDER=zoho
EMAIL_FROM_ADDRESS=<verified fallback sender>
EMAIL_FROM_NAME=Tuvi Solutions
EMAIL_DISABLE_SENDING=true
EMAIL_REDIRECT_TO=
EMAIL_OPEN_TRACKING_ENABLED=true

ZOHO_ACCOUNT_ID=<default account id>
ZOHO_FROM_EMAIL=<verified sender>
ZOHO_REGION=com.au
ZOHO_API_BASE_URL=https://mail.zoho.com/api/accounts
ZOHO_CLIENT_ID=<secret>
ZOHO_CLIENT_SECRET=<secret>
ZOHO_REFRESH_TOKEN=<secret>

OUTREACH_BULK_MAX=150
OUTREACH_EMAILS_PER_ACCOUNT=40
OUTREACH_EMAIL_COOLDOWN=24h
OUTREACH_SEND_INTERVAL=2s
OUTREACH_ZOHO_ACCOUNTS_JSON=[{"key":"sales-au-1","account_id":"<id>","from_email":"<verified sender>","client_id":"<secret>","client_secret":"<secret>","refresh_token":"<secret>","region":"com.au"}]
```

The singleton `ZOHO_*` values configure the generic Zoho adapter used by other
email flows; `OUTREACH_ZOHO_ACCOUNTS_JSON` configures the independently rotated
bulk-outreach pool. If the generic adapter is intentionally unused,
`EMAIL_PROVIDER=disabled` is valid while the rotating outreach pool is enabled
through `EMAIL_DISABLE_SENDING=false`. The account `key` is a stable,
non-secret identity. Do not change it when a refresh token is rotated or when
array order changes. Each account has at most 40 reserved attempts per cycle
and becomes available 24 hours after its 40th reservation. Zoho sends through
OAuth and its HTTPS Mail API; SMTP is rejected. Configuration rejects duplicate
normalized `(region, account_id)` identities even when aliases use different
keys, and PostgreSQL uniquely persists that provider identity. If a key is
renamed, the same quota row and usage/cooldown are retained rather than reset,
so one Zoho credential cannot receive multiple independent quotas.

Token-gated demos do not use a global signing secret. Each demo receives a
cryptographically random opaque token; `demo_sites` stores its bcrypt hash and
the campaign stores the current token needed to construct its reviewed tracked
link. `DEMO_TOKEN_TTL` controls expiry, and regeneration rotates the token,
updates the hash, and clears publication/campaign approval.

In `/opt/tuvi/env/ingestion.env` (mode `0600`):

```text
GOOGLE_PLACES_API_KEY=<restricted Places API (New) key>
PLACES_API_BASE_URL=https://places.googleapis.com/v1
APOLLO_API_KEY=<restricted Apollo key>
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
LEAD_OCR_MAX_ATTEMPTS=3
LEAD_OCR_RETRY_AFTER_HOURS=24
HUGGING_FACE_API_KEY=<vision provider key>
HF_VISION_MODEL=<supported vision model>
```

`OPENAI_API_KEY` or `GEMINI_API_KEY` may be used instead of Hugging Face. Keep
provider keys out of logs and source control.

## Migrate and start services in lease-safe order

After disabling legacy ingestion, stopping the old API/workers, and taking the
backup described above, run these as separate phases from
`/opt/tuvi/MonoRepo`. Do not collapse them into one `up` command during the
`000020` rollout.

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml --profile jobs \
  build migrate api worker scrape-worker ocr-job

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d postgres redis

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml run --rm --no-deps migrate

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d --no-deps api

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d --no-deps worker

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d --no-deps scrape-worker
```

The scrape worker exposes no host port. It polls due `scrape_jobs` through the
same PostgreSQL database as the API. Keep OCR disabled in `ingestion.env`; for
the controlled batch below, enable it only on that one-shot container:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml --profile jobs run --rm \
  -e LEAD_OCR_VERIFICATION_ENABLED=true ocr-job
```

## Authenticate

Use a seeded internal administrator. Keep the token in the current shell only:

```bash
export TUVI_API=https://api.tuvisolutions.com
export ADMIN_TOKEN="$(curl -fsS -X POST "$TUVI_API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  --data '{"email":"<admin email>","password":"<admin password>"}' | jq -r '.access_token')"
```

## Trigger and monitor Melbourne scraping

Triggering is idempotent while Melbourne has a queued, running, or waiting
restaurant job:

```bash
curl -fsS -X POST "$TUVI_API/api/v1/scrape-jobs" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data '{"city":"Melbourne","niche":"restaurant"}' | jq
```

Save the returned job ID and inspect it:

```bash
export SCRAPE_JOB_ID=<uuid>
curl -fsS "$TUVI_API/api/v1/scrape-jobs/$SCRAPE_JOB_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

curl -fsS "$TUVI_API/api/v1/scrape-jobs?limit=20" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Healthy long-running states:

- `running`: a worker owns a renewable lease.
- `waiting` + `request_limit`: exactly 500 combined Places/Apollo calls have
  been reserved; the cell, page token, and candidates remain checkpointed until
  `resume_at` 24 hours later.
- `waiting` + `revisit`: the current city coverage cycle completed; leaf cells
  restart 24 hours later to find Place IDs added since the previous cycle.
- `waiting` + `provider_error`: state is preserved and retried after 24 hours;
  inspect `last_error` before leaving it unattended.
- `waiting` + `coverage_incomplete`: even the deepest configured leaf still
  reached the Places result cap. That leaf is parked in the explicit
  `coverage_incomplete` cell state so later siblings can finish; the city cycle
  is not recorded complete, and capped leaves retry after 24 hours. Inspect
  `progress.cells_saturated` and increase `SCRAPE_GRID_MAX_DEPTH` only after a
  request-cost review if this repeats.
- `failed`: correct the permanent configuration/provider error, then requeue
  the same durable job with:

```bash
curl -fsS -X POST "$TUVI_API/api/v1/scrape-jobs/$SCRAPE_JOB_ID/retry" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

The retry endpoint accepts only `failed` jobs. It retains imported candidates
and coverage history, resets failed cells/candidates, and keeps a still-active
request-window count; it does not create a replacement city job.

Saturated cells dynamically subdivide through the configured depth (default
`12`). A leaf that remains saturated is never counted as completed coverage;
`progress.cells_saturated`, `waiting_reason=coverage_incomplete`, and
`last_error` keep the gap visible while preserving it for retry.

The job intentionally does not become terminal `completed`: recurring revisits
are how the system discovers newly opened/listed restaurants. Completion of a
coverage pass is recorded in `last_cycle_completed_at`.

Useful container logs:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml logs -f --tail=200 scrape-worker
```

## Run and schedule OCR

Run one database-networked OCR batch:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml --profile jobs run --rm \
  -e LEAD_OCR_VERIFICATION_ENABLED=true ocr-job
```

Inspect this controlled batch first. Only after its state transitions and draft
artifacts are accepted, set `LEAD_OCR_VERIFICATION_ENABLED=true` in
`/opt/tuvi/env/ingestion.env` and schedule the same one-shot job hourly with a
non-overlapping lock (adjust the path if needed):

```cron
15 * * * * cd /opt/tuvi/MonoRepo && flock -n /var/lock/tuvi-ocr.lock docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml --profile jobs run --rm ocr-job >> /opt/tuvi/logs/ocr.log 2>&1
```

OCR state meanings:

- `pending`: ready to claim.
- `running`: claimed; stale claims older than two hours can be reclaimed.
- `verified`: at least one image was analyzed successfully; `lead.prepare` is
  queued in the same transaction.
- `no_images`: no trusted Google-hosted direct image or Places photo resource was available;
  it is terminal and not outreach eligible.
- `failed`: resolution or OCR failed and is not eligible. An unchanged input is
  retried after `LEAD_OCR_RETRY_AFTER_HOURS` until
  `LEAD_OCR_MAX_ATTEMPTS` is reached.

The OCR process makes a fresh Place Details request immediately before analysis,
then resolves current photo resources to short-lived HTTPS `photoUri` values.
Expirable photo resource names and Photo Media URLs are never written back to
PostgreSQL. These OCR-only Place Details/Photo Media calls are separate from the
city scrape job's 500-call Places/Apollo request window. Unattended OCR
prioritizes these resources and only falls back to HTTPS direct images on
Google-controlled image hosts; arbitrary restaurant/CDN URLs are not fetched
from the VM.

After the automatic attempt limit, inspect `ocr_verification_errors` and fix the
provider/input issue. To deliberately requeue that unchanged profile (or a
`no_images` profile after confirming that images now exist), use this reviewed
database operation with the exact restaurant UUID:

```sql
UPDATE restaurant_profiles
SET ocr_status = 'pending', ocr_verified = false, ocr_verified_at = NULL,
    ocr_started_at = NULL, ocr_completed_at = NULL, ocr_attempts = 0,
    ocr_claim_id = NULL, ocr_claim_fingerprint = NULL,
    ocr_verification_errors = '[]'::jsonb, updated_at = now()
WHERE restaurant_id = '<restaurant uuid>'::uuid
  AND ocr_status IN ('failed', 'no_images');
```

Confirm exactly one row changed. Do not reset `verified` or `running` rows.

## Review and approve a prepared lead

Find the restaurant, demo, and campaign IDs from the admin APIs or the review
query below. Review each gate in order because approving a new profile decision
invalidates any older demo publication and campaign approval. The protected
preview endpoints do not return the stored bearer token or token hash.

```bash
export RESTAURANT_ID=<uuid>
curl -fsS "$TUVI_API/api/v1/restaurants/$RESTAURANT_ID/campaigns" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Inspect the OCR-verified restaurant/profile and bind the decision to both
returned versions:

```bash
PROFILE_REVIEW_JSON=$(curl -fsS \
  "$TUVI_API/api/v1/restaurants/$RESTAURANT_ID/profile/review-preview" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
printf '%s\n' "$PROFILE_REVIEW_JSON" | jq
PROFILE_RESTAURANT_UPDATED_AT=$(printf '%s' "$PROFILE_REVIEW_JSON" | jq -r '.restaurant_updated_at')
PROFILE_UPDATED_AT=$(printf '%s' "$PROFILE_REVIEW_JSON" | jq -r '.profile_updated_at')

curl -fsS -X PATCH "$TUVI_API/api/v1/restaurants/$RESTAURANT_ID/profile/review" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data "$(jq -nc \
    --arg status approved \
    --arg restaurant_version "$PROFILE_RESTAURANT_UPDATED_AT" \
    --arg profile_version "$PROFILE_UPDATED_AT" \
    '{status:$status,expected_restaurant_updated_at:$restaurant_version,expected_profile_updated_at:$profile_version}')" | jq
```

Next, inspect the exact allowlisted demo payload. Publish the reviewed version;
the API refuses stale content and also requires current OCR/profile approval:

```bash
export DEMO_SITE_ID=<uuid>
DEMO_REVIEW_JSON=$(curl -fsS \
  "$TUVI_API/api/v1/demo-sites/$DEMO_SITE_ID/review-preview" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
printf '%s\n' "$DEMO_REVIEW_JSON" | jq
DEMO_UPDATED_AT=$(printf '%s' "$DEMO_REVIEW_JSON" | jq -r '.updated_at')

curl -fsS -X PATCH "$TUVI_API/api/v1/demo-sites/$DEMO_SITE_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data "$(jq -nc --arg status published --arg version "$DEMO_UPDATED_AT" \
    '{status:$status,expected_updated_at:$version}')" | jq
```

Finally, inspect the recipient, subject, HTML, text, links/placeholders, and
opt-out content, then approve that exact campaign version:

```bash
export CAMPAIGN_ID=<uuid>
CAMPAIGN_REVIEW_JSON=$(curl -fsS "$TUVI_API/api/v1/campaigns/$CAMPAIGN_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
printf '%s\n' "$CAMPAIGN_REVIEW_JSON" | jq
CAMPAIGN_UPDATED_AT=$(printf '%s' "$CAMPAIGN_REVIEW_JSON" | jq -r '.campaign.updated_at')

curl -fsS -X POST "$TUVI_API/api/v1/campaigns/$CAMPAIGN_ID/approve" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data "$(jq -nc --arg version "$CAMPAIGN_UPDATED_AT" \
    '{expected_updated_at:$version}')" | jq
```

Changing the profile decision invalidates published demos and approved campaigns.
Returning only the demo to `draft` also immediately removes send eligibility.

If a token-gated demo has expired, or an older draft contains stale template/link
content, unpublish the demo and regenerate both the opaque token and current
campaign content. Regeneration is administrator-only, records the actor in a
`draft_regenerated` event, atomically binds automatic artifacts to the current
verified OCR/profile fingerprints, and clears campaign approval:

```bash
curl -fsS -X PATCH "$TUVI_API/api/v1/demo-sites/$DEMO_SITE_ID/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  --data '{"status":"draft"}' | jq

curl -fsS -X POST "$TUVI_API/api/v1/campaigns/$CAMPAIGN_ID/regenerate" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Inspect the regenerated safe demo payload and campaign detail through the two
protected GET endpoints above, then publish the demo and approve the campaign
again. An expired demo cannot be published or selected for quota, and a campaign
token must match the selected demo token hash.

When a later scrape changes the OCR image fingerprint, the system resets the
profile review, unpublishes the demo, and returns only an `approved` campaign
to draft. In-flight, sent, stopped, and `send_unknown` campaigns are retained
for explicit reconciliation. A stale OCR worker cannot finalize over the new
input.

## Enable and start outreach

Before production enabling:

1. Confirm the four production URL values exactly match the values above.
2. Confirm every configured sender is verified by Zoho.
3. Confirm `EMAIL_REDIRECT_TO` is empty in production.
4. Regenerate and re-review any old draft containing a local URL using the
   administrator sequence above. Production
   rendered-email preflight rejects unresolved placeholders, HTTP links,
   localhost/loopback links, and unapproved hosts before a provider call.
5. Set `EMAIL_DISABLE_SENDING=false` and restart both API and Go worker.

Start one approved bulk workflow:

```bash
curl -fsS -X POST "$TUVI_API/api/v1/outreach/bulk-send" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

curl -fsS "$TUVI_API/api/v1/outreach/bulk-send/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Only OCR-verified, profile-approved leads with a published demo, separately
approved campaign, non-empty unsuppressed email, and no prior confirmed send
are selected. The obsolete individual `send-step` route is not registered;
outreach can only enter the quota-managed bulk path.

`OUTREACH_BULK_MAX` is the per-worker-execution slice (150 by default), not a
manual-retrigger boundary. The same approved bulk job requeues itself while
eligible leads remain. It uses any still-available account slots first, then
sets its next run to the earliest PostgreSQL `available_at` after every account
is cooling down.

After each provider-accepted send:

- `email_delivery_attempts.status = sent` with global `send_sequence`, account
  cycle, and account sequence;
- one `email_events.sent` row is linked to the attempt;
- the campaign becomes `sent`;
- `restaurants.email_sent = true`;
- `restaurants.email_send_count` increments (first confirmed invite is `1`);
- `restaurants.last_email_send_sequence` records the global sequence;
- `restaurants.last_email_recipient` records the normalized immutable address
  from the accepted delivery attempt (inspect in PostgreSQL; never log it);
- the restaurant lifecycle moves from `lead`/`demo_ready` to `emailed`.

An ambiguous provider timeout becomes `send_unknown`; it consumes the reserved
account slot and is not automatically retried. Reconcile it with the provider
before any manual repair.

Every claimed delivery has a five-minute database lease. If a worker exits
before finalization, the next worker/startup/status check converts the expired
attempt to `unknown` and the campaign to `send_unknown`; it never retries an
ambiguous external send automatically.

Counter meanings are intentionally different:

- `email_delivery_attempts.send_sequence` is a global audit sequence allocated
  for every reserved attempt, so failed/unknown/skipped attempts may create gaps
  between confirmed sends;
- `account_sequence` is `1..40` inside one account cycle;
- `restaurants.email_send_count` counts confirmed provider-accepted sends for
  that restaurant and becomes `1` for its first invite.

## Email content and populated links

Source templates:

- `backend/internal/campaigns/templates/outreach.html`
- `backend/internal/campaigns/templates/outreach.txt`

Subject:

```text
A live demo for <Restaurant Name> — AI receptionist, website & more
```

The invite says that missed calls and outdated websites can make it harder for
guests to request a table, says a live preview plus a reservation-request form
has already been prepared, and presents:

- AI Voice Receptionist — “24/7 calls, reservation requests & callbacks” —
  `https://tuvisolutions.com/services/restaurants`;
- Presentation Websites — “Modern sites from your real menu & photos” —
  `https://tuvisolutions.com/services/restaurants`;
- Reservation Requests — “Guests submit table requests on your demo site” —
  its own tracked token-gated template-3 demo link;
- Custom Apps — “QR ordering, loyalty & more” —
  `https://tuvisolutions.com`.

The closing invitation is: “Reply anytime — happy to walk you through it in 10
minutes.”

The primary “Open <Restaurant Name> demo” CTA uses an API click-tracking URL on
`https://api.tuvisolutions.com/t/click/<token>`, which redirects to the
restaurant preview on `https://demo.tuvisolutions.com` with a demo slug
and token plus `template=1`. That page fetches
`/api/public/v1/demo/<slug>?token=<token>` with
`no-store`; it never falls back to the ungated public restaurant-index API.
The Reservation Requests service entry receives a separate click token whose
target adds `template=3`. The token-gated payload supplies a server-derived
restaurant UUID (never a mutable payload-provided ID), so that template can
submit a `pending` reservation request for the correct restaurant.
The current outreach template does not render a template-2 link.
Click and open tracking tokens expire 30 days after creation or at the demo's
earlier expiry, whichever comes first. Unsubscribe tokens intentionally remain
valid so an older email retains a working opt-out. Each new tracking/delivery row
stores the normalized address actually used for that send, so a later restaurant
email edit cannot redirect an old opt-out to a different recipient. Legacy token
rows created before migration `000021` have no provable recipient; their opt-out
links use the restaurant's current normalized address as a legacy-only fallback
and record `recipient_snapshot=false` for audit. Reconcile those events against
historical provider records when available.
`DEMO_TOKEN_TTL` can keep the underlying token-gated demo alive longer, but it
does not extend an emailed click/open URL past that 30-day cap.
The legacy index-based voice widget is intentionally not mounted on token-gated
previews until it supports a stable token-scoped restaurant identity; the AI
service link instead opens the presentation page above. Unpublishing or
expiring the demo therefore revokes the emailed preview. The footer includes
`https://api.tuvisolutions.com/t/unsubscribe/<token>`. When enabled, a 1x1 open
tracking image uses `https://api.tuvisolutions.com/t/open/<token>`.

## Database diagnostics

Run through the VM Postgres container without printing secrets:

```sql
SELECT city, niche, status, cycle_number, requests_used_window,
       max_requests_per_window, resume_at, waiting_reason,
       last_cycle_completed_at, current_cell_id
FROM scrape_jobs
ORDER BY created_at DESC;

SELECT ocr_status, count(*)
FROM restaurant_profiles
GROUP BY ocr_status
ORDER BY ocr_status;

SELECT r.id, r.name, r.email AS recipient,
       rp.description, rp.opening_hours, rp.phone, rp.address, rp.cuisines,
       rp.ocr_status, rp.review_status,
       rp.reviewed_at, rp.reviewed_by,
       d.id AS demo_site_id, d.status AS demo_status,
       d.public_payload, d.published_at, d.published_by, d.expires_at,
       c.id AS campaign_id, c.status AS campaign_status,
       c.subject, c.body_html, c.body_text,
       c.approved_at, c.approved_by
FROM restaurants r
JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
LEFT JOIN LATERAL (
  SELECT id, status, public_payload, published_at, published_by, expires_at FROM demo_sites
  WHERE restaurant_id = r.id ORDER BY created_at DESC LIMIT 1
) d ON true
LEFT JOIN LATERAL (
  SELECT id, status, subject, body_html, body_text, approved_at, approved_by
  FROM email_campaigns
  WHERE restaurant_id = r.id ORDER BY created_at DESC LIMIT 1
) c ON true
ORDER BY r.created_at DESC;

SELECT account_key, position, usage_count, send_limit, cycle_number,
       available_at, last_used_at, enabled
FROM outreach_email_accounts
ORDER BY position;

SELECT send_sequence, campaign_id, restaurant_id, recipient_email, status, account_cycle,
       account_sequence, provider_message_id, error_code, sent_at
FROM email_delivery_attempts
ORDER BY send_sequence DESC
LIMIT 100;

SELECT id, name, email AS current_email, last_email_recipient,
       last_email_send_sequence, last_email_sent_at
FROM restaurants
WHERE email_sent = true
ORDER BY last_email_sent_at DESC;

SELECT id, job_type, status, last_error, updated_at
FROM job_runs
WHERE last_error LIKE 'Legacy %'
   OR status = 'running'
ORDER BY updated_at DESC;
```

After migration, every `send_unknown` campaign requires comparison with the
Zoho provider before an operator changes any campaign/restaurant state. Never
requeue it merely because the old worker job is now `failed`.

## Stop controls

- Disable new email immediately with `EMAIL_DISABLE_SENDING=true` and restart
  API/worker. Already accepted provider sends cannot be recalled.
- Stop an individual campaign with `POST /api/v1/campaigns/{id}/stop`.
- Stop new scrape execution by stopping `scrape-worker`; checkpoints remain in
  PostgreSQL for a later restart.
- Disable scheduled OCR with `LEAD_OCR_VERIFICATION_ENABLED=false` or remove the
  cron entry. Running OCR claims finish independently and do not send email.
