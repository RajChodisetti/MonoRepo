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
                 -> one provider attempt per durable job activation
                 -> 40 persisted slots per account across an 8-hour window
                 -> rotate accounts after slot 40
                 -> 24-hour PostgreSQL cooldown and automatic continuation
                 -> confirmed send updates the delivery ledger and restaurant counter
```

Creating a scrape job, running OCR, or preparing drafts does not send email.
Bulk outreach still requires all three record approvals and an internal
administrator enabling the persisted email job from the Outreach UI. Internal
administrators can also send bounded selective/manual emails from the restaurant
list or a restaurant detail page; those ad hoc sends bypass the bulk email-job
flag and the generic `EMAIL_DISABLE_SENDING` adapter flag, but still require a
configured Gmail outreach sender, a contact email, and a recipient that has not
opted out.

## Required deployment order

1. Keep `LEAD_OCR_VERIFICATION_ENABLED=false`, keep the Outreach UI email job
   disabled, and retain `EMAIL_DISABLE_SENDING=true` for the generic adapter
   during rollout.
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
6. Apply migrations `000015` through `000029`, in order. `000019` depends on
   the review-audit columns introduced by `000018`; `000021` binds tracking and
   delivery audit rows to the immutable address actually used for outreach.
   `000022` returns legacy automatic artifacts to draft, queues fresh preparation,
   and binds new artifacts to both OCR input and the exact identity/public payload.
   `000023` makes maximum-depth coverage gaps a durable, visible waiting state;
   `000024` adds durable eight-hour Gmail pacing; and `000029` adds daily Gmail
   health evidence plus signed-demo engagement sessions/transcripts.
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

EMAIL_PROVIDER=disabled
EMAIL_FROM_ADDRESS=
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
OUTREACH_SEND_WINDOW=8h
OUTREACH_SEND_JITTER_MIN=2m
OUTREACH_SEND_JITTER_MAX=5m
OUTREACH_ZOHO_ACCOUNTS_JSON=[]
OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON=[{"key":"workspace-sales-1","mailbox_email":"sales1@example.com","from_email":"sales1@example.com","client_id":"<oauth client id>","client_secret":"<secret>","refresh_token":"<offline refresh token>"}]
OUTREACH_EMAIL_HEALTH_ENABLED=true
OUTREACH_EMAIL_HEALTH_RECIPIENT=rajchodisetti@gmail.com
OUTREACH_EMAIL_HEALTH_INTERVAL=24h
```

The singleton `ZOHO_*` values configure the generic Zoho adapter used by other
email flows; quota-managed restaurant outreach uses only the Gmail accounts in
`OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`. Add another JSON entry to add a sender
without a code change. If the generic adapter is intentionally unused,
`EMAIL_PROVIDER=disabled` is valid while the Gmail outreach pool is configured;
the restaurant outreach job is enabled from the admin UI.

The admin photo viewer resolves temporary Google Places media URLs through the
Go API. Keep its key in `/opt/tuvi/env/places-api.env` (mode `0600`) so the API
does not receive the Apollo or Hugging Face secrets from `ingestion.env`:

```text
GOOGLE_PLACES_API_KEY=<server-side Places API key>
PLACES_API_BASE_URL=https://places.googleapis.com/v1
PLACES_PHOTO_LIMIT=10
PLACES_PHOTO_MAX_WIDTH=1600
PLACES_API_TIMEOUT=20s
```

The response is deliberately `no-store`; media URLs are refreshed on demand and
must be displayed with the author attribution returned alongside each URL.

For Google Workspace, create one OAuth web application, request only
`https://www.googleapis.com/auth/gmail.send`, obtain offline consent separately
for each mailbox, and store that mailbox's refresh token in its entry.
`mailbox_email` is the primary mailbox authorized by the refresh token and is
the durable quota identity. `from_email` defaults to it; use a different value
only when that send-as alias is configured in Gmail. Token and Gmail endpoints
are fixed in code to Google's HTTPS hosts so credentials cannot be redirected.

The account `key` is a stable, non-secret identity. Do not change it when a
refresh token is rotated or when array order changes. Each account has at most
40 reserved attempts per cycle. The allowance is divided into 40 durable slots
over eight hours (about 12 minutes per slot), with a persisted random 2-5 minute
offset. During an on-time cycle, adjacent offsets normally yield effective gaps
of about 9-15 minutes. A separate global 2-5 minute minimum-delay guard spans
account transitions and prevents delayed/restarted workers from catching up in
a burst. After slot 40, the account becomes available again only after its
24-hour cooldown. Gmail uses OAuth and the HTTPS API; SMTP is rejected.
Configuration rejects duplicate provider identities even when aliases use
different keys, and PostgreSQL retains the usage/cooldown row across credential
rotation.

The worker sends one real Gmail health-check message per configured mailbox to
`OUTREACH_EMAIL_HEALTH_RECIPIENT` when the account is first registered and then
once every 24 hours. Successful provider message IDs and safe failures are shown
on the Outreach admin page. Health checks are controlled independently by
`OUTREACH_EMAIL_HEALTH_ENABLED`; the restaurant outreach run is controlled by
the persisted admin UI toggle.

The limit is conservative: it is at most 40 reserved provider attempts, not a
promise of 40 accepted or delivered emails. Fewer eligible approved leads,
skipped claims, provider rejection, or ambiguous outcomes can all reduce the
confirmed-send count. Ambiguous attempts still consume their reserved slot and
are never retried automatically.

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
LEAD_OCR_DAILY_REQUEST_LIMIT=200
OCR_WORKER_POLL_SECONDS=900
MENU_OCR_TIMEOUT=45
HUGGING_FACE_API_KEY=<vision provider key>
HF_VISION_MODEL=<supported vision model>
```

`OPENAI_API_KEY` or `GEMINI_API_KEY` may be used instead of Hugging Face. Keep
provider keys out of logs and source control.

### Administer the VM PostgreSQL database from the workstation

The dedicated `monorepo` database is exposed only on VM loopback port `15432`.
The ignored root `.env` contains the local-tunnel `DATABASE_URL` and standard
`PG*` variables; it must remain mode `0600`. Start the tunnel in one terminal:

```bash
ssh -i ~/.ssh/tuvi_vm_root_ed25519 \
  -L 127.0.0.1:15432:127.0.0.1:15432 \
  root@170.64.154.143 -N
```

Then connect from another terminal without copying the password into shell
history:

```bash
set -a
source .env
set +a
psql "$DATABASE_URL"
```

The VM application uses the separate internal URL stored in root-owned
`/opt/tuvi/env/monorepo.env`. Port `15432` must remain bound to `127.0.0.1`; do
not publish PostgreSQL on the VM's public interface.

## Migrate and start services in lease-safe order

After disabling legacy ingestion, stopping the old API/workers, and taking the
backup described above, run these as separate phases from
`/opt/tuvi/MonoRepo`. Do not collapse them into one `up` command during the
`000020` rollout.

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml --profile jobs \
  build migrate api worker scrape-worker ocr-worker ocr-job

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

docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d --no-deps ocr-worker
```

The scrape and OCR workers expose no host ports. Both poll durable PostgreSQL
state. Keep OCR disabled for the migration itself; enable the long-running OCR
worker only after its provider key, daily budget, and email-only candidate query
have been verified. The one-shot command below remains available for diagnosis.

The durable database OCR verifier ignores `MENU_OCR_MAX_IMAGES`: it refreshes
and attempts every discovered scraped photo. A restaurant reaches
`ocr_status=verified` only when every photo resolves and returns a successful
structured result. Partial provider, resolution, or parsing failures persist
processed/total counts and leave the restaurant `failed`; `no_images` is also
not verified.

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

## Run background OCR

Run one database-networked OCR batch:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml --profile jobs run --rm \
  -e LEAD_OCR_VERIFICATION_ENABLED=true ocr-job
```

After the controlled batch is accepted, set
`LEAD_OCR_VERIFICATION_ENABLED=true` in `/opt/tuvi/env/ingestion.env` and start
the durable background service:

```bash
docker compose --env-file /opt/tuvi/env/stack.env -p tuvi \
  -f infra/docker/docker-compose.vm.yml up -d --no-deps ocr-worker
```

Do not keep the former host OCR cron beside `ocr-worker`. The worker polls every
`OCR_WORKER_POLL_SECONDS` and claims only restaurants whose canonical
`restaurants.email` is nonblank. Before every vision-provider request it makes
an atomic reservation in `ocr_daily_request_usage`. The enforced global limit
cannot exceed 200 requests per UTC day and survives container restarts. SDK
automatic retries are disabled so one reservation can cause at most one provider
request.

Provider calls and image downloads use `MENU_OCR_TIMEOUT`. A timeout, HTTP 429,
or temporary provider failure returns the active claim (and any unstarted batch
claims) to `pending` without consuming an OCR attempt. A provider request that
times out still counts against the daily limit because its remote outcome is
ambiguous.

OCR state meanings:

- `pending`: ready to claim.
- `running`: claimed; stale claims older than two hours can be reclaimed.
- `verified`: every discovered image resolved and was analyzed successfully;
  `lead.prepare` is queued in the same transaction.
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
2. Confirm every configured Gmail sender shows healthy in the Outreach UI.
3. Confirm `EMAIL_REDIRECT_TO` is empty in production.
4. Regenerate and re-review any old draft containing a local URL using the
   administrator sequence above. Production
   rendered-email preflight rejects unresolved placeholders, HTTP links,
   localhost/loopback links, and unapproved hosts before a provider call.
5. Open the Outreach admin page and confirm the job is disabled before starting.

Start one approved bulk workflow from the Outreach admin page by selecting
**Enable email job**. The equivalent authenticated API call is:

```bash
curl -fsS -X PATCH "$TUVI_API/api/v1/outreach/email-job" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled":true}' | jq

curl -fsS "$TUVI_API/api/v1/outreach/bulk-send/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq
```

Only OCR-verified, profile-approved leads with a published demo, separately
approved campaign, non-empty unsuppressed email, and no prior confirmed send
are selected. The obsolete individual `send-step` route is not registered;
outreach can only enter the quota-managed bulk path.

For the durable account pool, each job activation crosses the provider boundary
at most once. The same approved bulk job then releases its lease and requeues
itself for the persisted PostgreSQL `available_at`; the worker never sleeps for
minutes between sends. `OUTREACH_BULK_MAX` remains a safety bound for
non-durable processing and does not turn one durable activation into a blast.
The job continues while eligible leads remain, follows the 8-hour account
schedule, and resumes automatically after the 24-hour cooldown following slot
40. No manual retrigger is required.

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
has already been prepared, and presents two promotional links:

- Personalized demo websites — the tracked token-gated demo link;
- Services catalog — `https://tuvisolutions.com/services/restaurants`.

The closing invitation is: “Reply anytime — happy to walk you through it in 10
minutes.”

The “Personalized demo websites” link uses an API click-tracking URL on
`https://api.tuvisolutions.com/t/click/<token>`, which redirects to the
restaurant preview on `https://demo.tuvisolutions.com` with a demo slug
and token plus `template=1`. That page fetches
`/api/public/v1/demo/<slug>?token=<token>` with
`no-store`; it never falls back to the ungated public restaurant-index API.
The token-gated payload supplies a server-derived restaurant UUID (never a
mutable payload-provided ID), so the demo can submit a `pending` reservation
request for the correct restaurant. Current outreach preview/send paths
canonicalize the stored campaign into exactly three email anchors: Personalized
demo websites, Services catalog, and Unsubscribe. Migration `000036` rewrites
unsent draft/approved rows that still contain the old template-specific
placeholders, template-option labels, or duplicated demo buttons.
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
services catalog link instead opens the Tuvi services page above. Unpublishing
or expiring the demo therefore revokes the emailed preview. The footer includes
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

- Disable the email job from the Outreach admin page. The currently in-flight
  Gmail request is allowed to finish so its delivery state is recorded; no next
  provider request starts. Already accepted provider sends cannot be recalled.
- Stop an individual campaign with `POST /api/v1/campaigns/{id}/stop`.
- Stop new scrape execution by stopping `scrape-worker`; checkpoints remain in
  PostgreSQL for a later restart.
- Disable scheduled OCR with `LEAD_OCR_VERIFICATION_ENABLED=false` or remove the
  cron entry. Running OCR claims finish independently and do not send email.
