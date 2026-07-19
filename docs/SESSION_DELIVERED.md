# Session Delivered Log

This is the detailed running delivery log for coding-agent sessions. Update this
before the final response of every implementation or planning session.

Each entry should explain:

- what was delivered;
- why it was delivered;
- tests or checks run;
- business value;
- how the work fits with the rest of the Phase 1 or Phase 2 plan;
- risks, gaps, or follow-ups.

## 2026-07-17 — Low-Cost OCR Selection and Full-Photo Verification

**Role:** AI Workflow, Backend, Frontend, Security, Test, Cost/Eval, and DevOps
Agent

**Model Evaluation:** Current Hugging Face provider metadata was checked through
Context7, official documentation, and the live OpenAI-compatible router model
catalog. Four inexpensive VLM routes were tested read-only against the same
licensed restaurant-menu, plated-food, and non-restaurant fixtures. The free
Ternary Bonsai route and `Qwen/Qwen3.5-9B:deepinfra` each completed only one of
three requests. `google/gemma-3-4b-it:deepinfra` and Gemma 3 12B both completed
and correctly classified all three; the 4B model extracted ten visible menu
items and was selected because its live listed price was lower at $0.05 per
million input tokens and $0.10 per million output tokens. Root-only benchmark
artifacts are under `/opt/tuvi/benchmarks/`.

**Delivered:** Durable database OCR now processes every discovered Google
Places and trusted direct scraped photo, ignoring the manual/file
`MENU_OCR_MAX_IMAGES` cap. Verification requires an exact full-success
contract: discovered, analyzed, and successful counts must match; there must be
zero resolution/model/parsing failures. Partial attempts persist their model,
processed/total/failed counts, token usage, and sanitized errors before ending
as `failed`. Migration 000028 resets any legacy partial-success verification
and adds a database check preventing `ocr_status=verified` unless
`menu_ocr.all_images_processed=true`. The restaurant Overview and Profile
Review UI now shows discovered/analyzed/successful/failed counts, the all-photo
result, model, and provider.

**Checks Run:** `go test ./backend/...`, `go vet ./backend/...`, and
`go build ./backend/cmd/...` passed. Admin `npm run lint` and
`npx tsc --noEmit --incremental false` passed; the production Node 22 Docker
build passed. Six new OCR contract/unit tests passed locally and in the exact
production OCR image. The full production-image automation suite ran 32 tests:
30 passed and only the same two obsolete `daily_ingestion.py` tests failed
because that retired entry point intentionally refuses to replace the durable
pipeline. Migration 000028 passed a transaction/rollback test against the live
database before it was applied.

**Production Deployment and Pilot:** Commit `fd2cc94` was pushed to `master`
and deployed at `/opt/tuvi/releases/monorepo-fd2cc94`; migration 000028 is
active. API, worker, and admin-web were recreated from the exact release, and
API/admin health checks return HTTP 200. A validated backup exists at
`/opt/tuvi/backups/pre-full-photo-ocr-fd2cc94.sql.gz` (mode 0600; 7,326,193
bytes uncompressed), with a matching root-owned ingestion-config backup.

The one-restaurant production pilot discovered, resolved, analyzed, and
successfully classified all ten photos in 51 seconds, so and only then did the
row become verified. It used 6,110 input and 1,068 output tokens, approximately
$0.00041 at the current listed rate, and extracted no menu items because none
of those ten photos was classified as a menu document. The downstream worker
created one demo draft and one campaign draft; nothing was approved, published,
or sent. Production is now 1 verified, 8 failed, and 470 pending.
`LEAD_OCR_VERIFICATION_ENABLED=false` remains persisted, there is no OCR cron,
and no OCR container remains running.

**Business Value / Follow-up:** OCR now proves full photo coverage instead of
mistaking one successful image for restaurant verification, while exposing the
evidence needed for human review and cost tracking. At the pilot rate, 470
similar pending restaurants would cost roughly $0.19 in model tokens but take
about 6.7 serial hours; image complexity and provider latency can vary. A bulk
run remains intentionally unscheduled so its batch size and operating window
can be chosen explicitly.

## 2026-07-17 — Approved Qwen2.5-VL Hyperbolic Production Pilot

**Role:** AI Workflow, Security, Cost/Eval, and DevOps Agent

**Approval and Scope:** The user explicitly approved changing the production
OCR route to `Qwen/Qwen2.5-VL-7B-Instruct:hyperbolic` and running a
five-restaurant pilot. The persistent OCR switch stayed disabled, no OCR
schedule was installed, and the one-shot container was hard-limited to five
pending profiles. Outreach remained disabled and no profile, demo, or campaign
approval gate was bypassed.

**Safety and Configuration:** Before the write, a validated live-database
backup was created at
`/opt/tuvi/backups/pre-qwen25-five-pilot-a2812de.sql.gz` (mode 0600;
7,325,725 bytes uncompressed) and the root-owned ingestion configuration was
backed up to
`/opt/tuvi/backups/ingestion.env.pre-qwen25-five-pilot-a2812de` (mode 0600).
Only `HF_VISION_MODEL` and `MENU_OCR_MODEL` were changed to the approved
route. `LEAD_OCR_VERIFICATION_ENABLED=false` remains persisted, there are zero
OCR cron entries, and no OCR container remains running.

**Pilot Result:** The worker claimed exactly five pending profiles and
resolved ten Google Places images for each. Hyperbolic returned HTTP 400 for
all 50 vision requests, so the bounded job exited after
`verified=0 failed=5`. The live database moved from 3 failed / 476 pending to
8 failed / 471 pending, with each new failure recorded once and the sanitized
error `No image could be analyzed successfully`. No profile remains in
`running`; `menu_images` and `gallery_images` remain empty; demo sites stayed
at one, campaigns stayed at zero, and no `lead.prepare` job or outreach action
was created. Because no inference completed, the pilot produced no useful
latency, token-usage, quality, or price sample; provider-side billing should be
checked separately rather than inferred from HTTP status alone.

**Root Cause and Checks:** Context7 and current official Hugging Face
documentation confirmed the provider-suffix routing contract before the
change. After the failed pilot, the live Hugging Face model metadata showed
that `Qwen/Qwen2.5-VL-7B-Instruct` is now mapped only to
`featherless-ai`; `hyperbolic` is no longer a live provider for this model.
The OpenAI-compatible router's current model list also does not advertise this
model. Production API and admin login checks both still return HTTP 200. The
root-only pilot log is
`/opt/tuvi/logs/qwen25-five-pilot-a2812de.log` (mode 0600).

**Business Value / Follow-up:** The explicit approval was exercised with a
bounded blast radius and left a complete audit/rollback trail, but this route
cannot currently run the OCR workload. OCR remains off. Switching to
Featherless or another current VLM/provider is a new production model-route
decision and should begin with a non-restaurant compatibility probe before
another database-backed pilot.

## 2026-07-17 — Restaurant Photo URLs, OCR Visibility, Real Website Preview, and Production OCR Preflight

**Role:** Full-Stack, Backend, Security, Test, and DevOps Agent

**Delivered:** The restaurant-specific admin view now resolves and displays up
to ten current Google Places photo URLs with dimensions and required author
attribution, shows the complete URL and an open-image action, and keeps the
Places credential server-side. The scraper had intentionally persisted the
Google Place ID and photo count rather than expiring media URLs; the new
`GET /api/v1/restaurants/{id}/images/google` adapter resolves fresh URLs on
demand and returns `Cache-Control: no-store`. Existing OCR-classified menu and
gallery images also show their stored URLs. The profile and overview tabs now
distinguish “not checked,” running, verified, no-images, and failed OCR states,
including attempts, start/completion timestamps, and sanitized errors.

The Demo tab now exposes the existing database-driven generator for each
restaurant UUID (mapped to the same current `?id=<index>&template=1|2|3`
URLs used by `demo.tuvisolutions.com`) and explains the separate token-gated
workflow. “Create restaurant demo draft” now snapshots the restaurant's actual
public-safe database payload instead of the old `Sample Cafe` default;
“Inspect payload” is read-only, “Publish” enforces OCR verification plus human
profile approval, and “Unpublish” immediately revokes public access while
retaining the draft. Campaign creation now supplies the required one-time demo
token and is disabled when that token is unavailable.

**Apollo / Photo Diagnosis:** The earlier scrape result was not an Apollo
adapter failure: 45 candidates were enriched, 159 returned no candidate, 33
were skipped because no usable domain existed, and the job recorded no Apollo
provider errors. All 237 imported profiles had ten Google photo resources
(2,370 total), but URLs were deliberately not persisted because Places media
URLs expire. Before this release all 237 profiles were OCR `pending` with zero
attempts and both `menu_images` and `gallery_images` empty.

**Checks Run:** `go test ./backend/...` (157 passed), `go vet ./backend/...`,
and `go build ./backend/cmd/...` passed. `npm run lint`,
`npx tsc --noEmit --incremental false`, and a production Next.js build under
Node 22 passed. The host's Node 23 reproduces the existing Next.js JSON-parser
failure even on unchanged commit `82eb2e6`, so it is not introduced here.
OpenAPI YAML parsing, Compose config validation, diff checks, Google Places
adapter tests (including key/error redaction and attribution), and the new
no-store handler test passed. In the production OCR dependency image, 24 of 26
legacy automation tests passed; two outdated `daily_ingestion.py` tests fail
because that retired entry point now intentionally refuses to run once the
durable city pipeline is installed.

**Production Deployment:** Commit `199241c` was fast-forwarded to the remote
default branch `master` and deployed from a `git archive` release at
`/opt/tuvi/releases/monorepo-199241c`. The API, worker, and admin-web containers
were rebuilt/recreated; all run the exact release images with zero restarts.
The new protected routes return 401 without authentication, the API root and
admin login return 200, and the API container has the isolated Places key from
root-owned `/opt/tuvi/env/places-api.env` rather than the Apollo/Hugging Face
ingestion file. No migration or email send occurred. Rollback images are tagged
`rollback-eaa525c`, `/opt/tuvi/previous-release-path` records the prior release,
and the validated pre-OCR database backup is
`/opt/tuvi/backups/pre-ocr-photo-ui-199241c.sql.gz` (mode 0600; 3,680,849 bytes
uncompressed).

**OCR Preflight / Blocker:** A deliberately limited three-profile OCR run was
started after the backup. Google Places successfully resolved ten images for
each profile, but all 30 Hugging Face calls returned HTTP 400. The live router
model list confirms the configured `Qwen/Qwen2-VL-7B-Instruct` route is no
longer served; current alternatives include
`Qwen/Qwen3-VL-30B-A3B-Instruct`. The three rows are now correctly visible as
`failed`, attempt 1, with sanitized errors; 234 remain pending and image tables
remain empty. `LEAD_OCR_VERIFICATION_ENABLED=false` remains in production and
no cron was installed. Changing the production vision model is an explicit
model-route/cost approval gate, so recurring OCR is intentionally paused until
that approval is received.

**Business Value / Plan Fit:** Internal operators can now see the source photo
URLs, immediately tell whether OCR actually checked a row, open the existing
restaurant website generator by restaurant record, and understand the gated
demo lifecycle. The controlled preflight converted a silent operational gap
into a bounded, auditable model-configuration decision without enabling email
or repeatedly billing a broken OCR route.

## 2026-07-17 — Lead Photo Management, Ad Hoc/Bulk Send with Preview, Demo Links, and Admin Portal Production Deployment

**Role:** Full-Stack and DevOps Agent

**Delivered:** In the admin portal (`apps/web`), the restaurant lead detail
view gained a Photos tab (list/hide/restore scraped menu and gallery images,
soft-hide via new `hidden_at`/`hidden_by` columns so the public site-image
endpoints and underlying scrape data are unaffected), a Demo links section,
and a header "Send email" button; the restaurants list gained checkbox
multi-select with a bulk send action bar. Both flows share a
`SendPreviewModal` that renders the exact subject/HTML/text before the send
is confirmed. Backend additions: `GET/DELETE/POST` image visibility
endpoints, `GET .../demo-links`, and a new ad hoc outreach send path
(`POST .../outreach/adhoc-send`, single and batch) — a deliberate product
decision, made explicitly with the user, to skip the OCR/profile/demo/
campaign approval gates the existing quota-managed bulk pipeline enforces,
while still keeping the `EMAIL_DISABLE_SENDING` kill switch and the
`email_suppressions` opt-out list as non-negotiable. The existing
`POST /api/v1/outreach/bulk-send` pipeline is unmodified. Migration 000026
adds the visibility columns; a pre-existing migration numbering collision
(both the current work and the earlier `admin_portal` merge had claimed
`000009`) was discovered and fixed by renumbering the unrelated,
never-applied `scrape_ledger` migration to `000027`.

**Deployed:** `apps/web` is now live at `https://api.tuvisolutions.com/admin`
— same host as the API, Caddy path-routed (`handle /admin*`), no new
subdomain, per explicit instruction. This is the admin portal's first
production deployment (previously undeployed). Two real bugs were found and
fixed only by deploying and testing against a running container, not by
documentation review: (1) `next start` re-evaluates `next.config.ts`
(and therefore `basePath`) at server boot, so `NEXT_PUBLIC_BASE_PATH` had to
be set in the Docker runtime stage, not only the builder stage that produces
the static assets — without it, every page under `/admin/*` 404'd while the
same paths worked fine unprefixed; (2) `apps/web`'s Next.js pin (15.5.0) had
a flagged CVE, upgraded to 15.5.20 in the same pass. `api` and `worker` were
rebuilt and recreated to ship the new Go endpoints; no other existing
service (postgres, redis, scrape-worker, voice-agent, tuvi-website, template,
catalog) was rebuilt or restarted.

**Checks Run:** `go build ./... && go vet ./... && go test ./...` — full
backend suite green, including 5 new tests for the ad hoc send path
(disabled-sending rejection, suppressed-email rejection, no-campaign-draft
rejection, successful send + restaurant record update, batch partial
results). `npx tsc --noEmit` clean on `apps/web` after every change (the
local host's Node v23 breaks `next build`'s SWC step for unrelated reasons —
confirmed pre-existing by testing the already-shipped `tuvi-website` app,
which fails identically — so real build verification was done via the
Node 22 Docker image, matching this repo's own Dockerfile convention).
Production verification: pre- and post-migration Postgres backups taken
against the correct live database (`monorepo`, not the legacy
`tuvi`/`restaurant_platform` role/db pair, which turned out to be an inert
leftover from before an earlier least-privilege database migration);
`schema_migrations` confirms 000026 and 000027 applied; `https://api.
tuvisolutions.com/admin/login` returns 200; `/admin/dashboard` without a
session correctly redirects to `/admin/login?next=...`; the Go API's
existing public endpoints, `/docs/`, `/openapi.yaml`, and the other four
Tuvi domains (corporate site, voice, demo, and the unrelated Tilnest/
SustainabilityWise/n8n sites sharing the VM) all verified unaffected
post-deploy.

**Business Value / Plan Fit:** Gives the internal_admin operator a working
way to curate scraped photos, see a lead's demo link(s) without leaving the
detail view, and send outreach to one or several hand-picked leads with a
mandatory preview — closing the gap between "leads exist in the database"
and "a human can act on them" that the admin portal's first (undeployed)
merge left open.

**Production Deployment:** Deployed via `git archive` snapshots into new
`/opt/tuvi/releases/monorepo-<sha>` directories (release directories on this
VM are not git checkouts), following the existing symlink-swap/rollback
convention. Final release: `eaa525c`. `/opt/tuvi/MonoRepo` now points there;
`/opt/tuvi/previous-release-path` records the prior release for rollback.

**Risks / Follow-ups:** `EMAIL_DISABLE_SENDING=true` remains the production
default (unchanged, per the existing human-approval gate in
`docs/runbooks/vm-deployment-plan.md`) — the new ad hoc send buttons will
503 in production until that separate decision is made, identical to
today's existing bulk-send button. Several intermediate release directories
from this session's iterative debugging (`monorepo-b19b870` through
`monorepo-a00160a`) remain on the VM under `/opt/tuvi/releases/`; harmless
(98GB free) but could be pruned. The bare `https://api.tuvisolutions.com/admin`
root (no further path) takes two redirect hops to reach the login page
instead of one (`/admin` → `/admin/dashboard` → `/admin/login`) — cosmetic
only, not a functional issue, not investigated further given time already
spent on the basePath root cause above.

## 2026-07-16 — Admin Portal Merge and Documentation Sync

**Role:** Documentation Agent

**Delivered:** Fetched `origin/master` and merged its 4 new commits — most
notably PR #8 (`admin_portal`) — into `agent/tuvi-oauth-homepage-verification`.
This replaces the `apps/web` placeholder with the real Next.js `internal_admin`
console (dashboard, scrape-jobs list/detail/retry, restaurant list/detail with
profile-OCR review + demo + campaign + member management, outreach bulk-send),
which talks to the Go API only through its own same-origin `/api/admin/*`
BFF/proxy route with an httpOnly session cookie. The merge also brought in
`automation/outreach/scrape_ledger.py`, `daily_pipeline.py`, `identity.py`, and
migration `000027_scrape_ledger` (renumbered from 000009 during deployment after discovering it collided with the already-applied production migration 000009_email_campaigns). Updated `docs/SERVICES.md` (port plan,
long-running services, interlinks, notes) and `AGENTS.md` LIVING MEMORY >
Current Repo Shape / Recent Agent Updates to describe `apps/web` as the live
admin UI instead of a placeholder. Reviewed `RTK.md` (local Codex CLI
token-saving wrapper doc) for accuracy; no change needed — content is unrelated
to this feature and still correct.

**Checks Run:** `git merge` completed with no conflicts (verified via
`git status` after merge). No code was written, so no test suite was run;
`apps/web/README.md` (already accurate from the source commit) was cross-checked
against the actual route files (`src/app/(admin)/**`, `src/lib/api.ts`,
`src/lib/client-api.ts`, `src/app/api/admin/proxy/[...path]/route.ts`,
`src/middleware.ts`) before writing doc updates.

**Business Value / Plan Fit:** Keeps agent-facing context (AGENTS.md,
docs/SERVICES.md) truthful about repo shape so future sessions don't
mis-describe `apps/web` as unimplemented. Surfaces a real gap: the admin portal
is not yet part of `infra/docker/docker-compose.vm.yml` or the VM Caddyfile, so
it is not reachable on any `tuvisolutions.com` subdomain yet — flagged as a
follow-up rather than actioned, since VM/Caddy changes are a production
deployment decision requiring explicit approval.

**Production Deployment:** None. This session only merged branches locally and
edited documentation; nothing was pushed or deployed.

**Follow-ups:** Decide and implement a VM domain (e.g.
`admin.tuvisolutions.com`) and add an `apps/web` service block to
`docker-compose.vm.yml` + the Caddyfile if/when the admin portal should be
reachable outside the VM's private network.

## 2026-07-15 — Equal Restaurant Demo Sizing and Deployment

**Role:** Frontend, DevOps, and Documentation Agent

**Delivered:** Corrected the alternating restaurant-feature grid so both demo
videos receive the same desktop width. The first row now assigns its left video
the larger `1.22fr` column, while the reversed row assigns that same `1.22fr`
width to its right video. Both retain the identical 16:9 frame, border, radius,
shadow, autoplay, looping, muted playback, and control-free presentation.

Published functional commit `4596e32` on
`agent/tuvi-oauth-homepage-verification`. Archived the exact remote-matched
commit, verified its checksum locally and on the VM, built only
`tuvi-website`, and recreated it with `--no-deps`. The prior `bc7f106` website
release and image are retained as the immediate rollback. No migration or API,
worker, scrape, OCR, email, database, Redis, voice, template, catalog, Caddy,
DNS, or environment action ran.

**Checks Run:** `git diff --check`, TypeScript validation, and local and VM Node
22 production builds passed with all 13 routes generated. Local and live HTML
contain both complementary grid templates, exactly two videos, matching
16:9 frames, autoplay and loop attributes, and no play/pause controls. The
public restaurant route returns HTTP 200. The running image matches `4596e32`,
has zero restarts, and all other services remain up.

**Business Value / Plan Fit:** Removes a visible inconsistency between the two
restaurant demos so the alternating product story feels deliberate and
balanced rather than making one capability appear less important.

**Risks / Follow-ups:** This corrects the deterministic desktop grid sizing;
both videos were already full-width and equal below the desktop breakpoint.
Rollback to `bc7f106` restores the previous asymmetric alternating grid.

## 2026-07-15 — Restaurant Demo Video Autoplay and Deployment

**Role:** Frontend, DevOps, and Documentation Agent

**Delivered:** Updated both `/services/restaurants` product demos to use the
same fixed 16:9 frame, widened the feature container, and increased the desktop
video column so the demonstrations render equally and slightly larger. Removed
the play/pause overlay and client-side playback state. Both videos now request
muted autoplay, loop continuously, play inline, preload their media, disable
picture-in-picture, and expose no native controls.

Published functional commit `bc7f106` on
`agent/tuvi-oauth-homepage-verification`. Archived the exact remote-matched
commit, verified its checksum locally and on the VM, built only
`tuvi-website`, and recreated it with `--no-deps`. The prior `371c062` website
release and image are retained as the immediate rollback. No migration or API,
worker, scrape, OCR, email, database, Redis, voice, template, catalog, Caddy,
DNS, or environment action ran.

**Checks Run:** `git diff --check`, TypeScript validation, and local and VM Node
22 production builds passed with all 13 routes generated. Local and live HTML
contain exactly two videos and two sets of autoplay, loop, muted, inline,
controls-list, and picture-in-picture attributes, with no play/pause button
language. The public restaurant route returns HTTP 200 and both CDN media URLs
return HTTP 200 with `video/mp4`. The running image matches `bc7f106`, has zero
restarts, and all other services remain up.

**Business Value / Plan Fit:** The restaurant demonstrations now behave like
consistent ambient product previews, with larger matching frames and no
interaction chrome competing with the service story.

**Risks / Follow-ups:** Muted autoplay is supported by current major browsers,
but individual browser, battery, or data-saving policies may still prevent it.
Removing the pause control creates a continuous-motion accessibility tradeoff;
rollback to `371c062` restores the prior user-controlled playback behavior.

## 2026-07-15 — Homepage Workspace Disclosure Removal and Deployment

**Role:** Frontend, Security, DevOps, and Documentation Agent

**Delivered:** Removed the compact Google Workspace disclosure from the public
homepage and deleted its now-unused component. The customer-facing AI and custom
software journey now flows directly from the contact section into the footer.
The dedicated `/google-workspace`, `/privacy`, and `/terms` pages remain live,
and the footer retains a small link to the Workspace app page.

Published functional commit `371c062` on
`agent/tuvi-oauth-homepage-verification`. Archived the exact remote-matched
commit, verified its checksum locally and on the VM, built only
`tuvi-website`, and recreated it with `--no-deps`. The prior `7651213` website
release and image are retained as the immediate rollback. No migration or API,
worker, scrape, OCR, email, database, Redis, voice, template, catalog, Caddy,
DNS, or environment action ran.

**Checks Run:** `git diff --check`, TypeScript validation, and the local and VM
Node 22 production builds passed with all 13 routes generated. Local and live
rendered HTML checks confirmed that the quoted Workspace disclosure text is
absent while the AI/custom-software hero remains. Loopback and public HTTPS
checks returned HTTP 200 for the homepage, Workspace, Privacy, Terms, booking,
and restaurant routes. The running image matches commit `371c062`, has zero
restarts, and all other services remain up.

**Business Value / Plan Fit:** Removes an internal application explanation from
the primary marketing experience and keeps the homepage focused on prospective
software and AI customers.

**Risks / Follow-ups:** This deliberately removes the homepage purpose statement
that Google cited in the previous branding-verification rejection. The dedicated
application and legal pages remain public, but Google may reject branding
re-verification again because the homepage itself no longer explains `tuvi`.
Rollback to `7651213` restores the compact disclosure if required.

## 2026-07-15 — Customer-First Homepage Refinement and Deployment

**Role:** Frontend, Security, Reviewer, DevOps, and Documentation Agent

**Delivered:** Refocused the Tuvi homepage on its customer proposition: practical
AI systems, custom software, websites, apps, automation, and integrations. The
hero now uses short intentional headlines and a restrained desktop-only solution
panel instead of the oversized, cropped logo artwork. Mobile, tablet, and small
laptop widths are copy-led with natural height; the broader Services menu and
What We Build section now reinforce the AI/software offer rather than implying
that Tuvi serves only restaurants.

Replaced the large homepage OAuth block with a compact, light disclosure after
the primary contact journey. It still names the exact `tuvi` application,
identifies Tuvi Solutions as operator, explains reviewed Gmail API delivery and
the `gmail.send` boundary, states that inbox content cannot be read, and links
the app details, Privacy, and Terms pages. The full explanation remains on
`/google-workspace`, while customer-facing homepage metadata now describes AI,
websites, and apps.

Published functional commit `7651213` on
`agent/tuvi-oauth-homepage-verification`. Archived the exact remote-matched
commit, verified the SHA-256 checksum locally and on the VM, built only
`tuvi-website`, and recreated it with `--no-deps`. The prior `35800dd` release
and image are retained as the immediate rollback; `4eaf7fa` remains a second
fallback. No migration or API, worker, scrape, OCR, email, database, Redis,
voice, template, catalog, Caddy, DNS, or environment action ran.

**Checks Run:** `git diff --check`, TypeScript validation, and the local Node 22
production build passed with all 13 routes generated. Rendered homepage checks
confirmed the customer-first title/copy and compact OAuth disclosure, and
confirmed the removed fact-card language is absent. Independent source reviews
found no hierarchy or responsive blockers at 375, 768, 1024, or 1440 pixels and
no OAuth-content blocker. The VM build passed; loopback and public HTTPS checks
returned 200 for all main, app, legal, booking, and restaurant routes. The live
image matches commit `7651213`, has zero restarts, and every non-website
container retained its prior ID and start time.

**Business Value / Plan Fit:** Visitors now understand Tuvi as an AI and custom
software partner immediately, without internal Workspace tooling competing with
the sales journey. The smaller disclosure preserves the public information
needed for OAuth review while keeping the homepage conversion-focused.

**Risks / Follow-ups:** The in-app visual browser was unavailable, so visual
confidence comes from the supplied screenshot, responsive source review,
production builds, and rendered/live HTML checks. Google verification remains
an external decision and is not guaranteed. Keep the consent-screen name exactly
`tuvi` and its Privacy/Terms URLs matched to the live links.

## 2026-07-15 — OAuth Homepage Identity, Responsive Polish, and Deployment

**Role:** Frontend, Security, Reviewer, DevOps, and Documentation Agent

**Delivered:** Made the public homepage explicitly identify the OAuth
application as **`tuvi`**, separate from its operator, Tuvi Solutions. Added a
prominent Google Workspace application section that explains the app's purpose,
authorized-user boundary, HTTPS Gmail sending flow, exact `gmail.send` scope,
and the Google data it cannot read. Replaced the conflicting "Tuvi Outreach"
name across the application page, metadata, Privacy Policy, Terms, navigation,
and footer. Centralized this identity and copy in the site-content model.

Polished the responsive layout by keeping the hero stacked through tablet and
small-laptop widths, balancing its desktop type and logo, removing excess form
spacing, tightening nested contact-card padding, and making mobile actions and
long email/link content wrap safely. Independent source reviews found no
release blockers at 375, 768, 1024, or 1440 pixels.

Published functional commit `35800dd` on the new
`agent/tuvi-oauth-homepage-verification` branch. Created a Git archive from that
exact remote-matched commit, verified its SHA-256 checksum locally and on the
VM, and built it from `/opt/tuvi/releases/35800dd...`. Retained the prior
`4eaf7fa` website source and image under immutable and rollback tags, then
recreated only `tuvi-website` with `--no-deps`. No migration or API, worker,
scrape, OCR, email, database, Redis, voice, catalog, Caddy, or DNS action ran.

**Checks Run:** `git diff --check` — passed;
`npx tsc --noEmit --incremental false --pretty false` — passed; production
Node 22 `next build` — passed with all 13 routes generated. A local production
server returned HTTP 200 for `/`, `/google-workspace`, `/privacy`, and `/terms`;
rendered HTML contained the exact `tuvi` identity, app-purpose heading,
`gmail.send only` disclosure, and all three public links. The in-app browser
surface was unavailable, so responsive verification combined static review,
production compilation, and rendered-HTML smoke checks. The VM build also
passed with all 13 routes generated. Loopback checks returned HTTP 200 for `/`,
`/google-workspace`, `/privacy`, `/terms`, `/book`, and
`/services/restaurants`; public HTTPS checks returned 200 on the root, `www`,
Workspace, Privacy, and Terms URLs. The live homepage contains the required app
name, purpose, permission, and legal-link text. The running image matches the
commit tag, its restart count is zero, and all non-website services remain up.

**Business Value / Plan Fit:** Directly addresses Google's prior branding
findings by making the app name and purpose public and consistent while
improving the homepage's usability and credibility without weakening Phase 1's
human-review boundary.

**Risks / Follow-ups:** The website change is live. Keep the Google Cloud app
name exactly `tuvi`, submit the final non-redirecting homepage/privacy/terms
URLs, verify domain ownership, use the matching logo, and request
re-verification. Approval is not guaranteed, and the separate Gmail
outreach-policy concern is unchanged. The Compose build still warns that
Buildx is unavailable and uses a floating `node:22-alpine` base; neither
affected this successful release.

## 2026-07-14 — Tuvi Brand Redesign and Legal-Site Deployment

**Role:** Frontend, Security, Reviewer, and DevOps Agent

**Delivered:** Redesigned the canonical Tuvi Solutions Next.js website around
the supplied logo and its exact ivory/forest palette. The corporate homepage
and restaurant-services experience now share an editorial brand system,
responsive navigation, accessible line icons, calmer motion, clearer service
language, optimized restaurant video controls, and accurate reservation,
campaign, callback, and introductory-offer wording. Added the supplied logo as
a production asset plus a tightly framed browser icon. Published the public
Google Workspace app page, Privacy Policy, and Terms of Service with working
navigation and canonical metadata.

Published website-only commit `4eaf7fa` to `origin/phase1_03/backend`, built it
from a checksum-verified Git archive on the VM, preserved the prior image as
`tuvi-tuvi-website:rollback-07ebe32`, and recreated only the `tuvi-website`
Compose service. No migration ran and no API, worker, voice, database, Redis,
Caddy, DNS, catalog, or outreach service was changed. The website-specific
release marker is now `4eaf7fa`; the global application marker remains
`8d392da`.

**Checks Run:** Independent release reviews reported no remaining production
blockers. `git diff --check`, TypeScript validation, and the Node 22 production
Next.js build passed with all 13 pages generated. Local production smoke checks
and VM loopback/public checks returned HTTP 200 for `/`,
`/services/restaurants`, `/privacy`, `/terms`, `/google-workspace`, `/book`, and
both brand assets. Website, API, worker, voice, PostgreSQL, and Redis containers
are running with zero restarts; voice and demo public checks returned HTTP 200.

**Business Value / Plan Fit:** Presents one credible Tuvi identity across the
corporate and restaurant sales journeys, makes the restaurant offering easier
to evaluate, and supplies the public legal URLs required for the Google OAuth
consent-screen configuration while preserving Phase 1's human-review boundary.

**Risks / Follow-ups:** The container build reports two moderate npm dependency
advisories; review and upgrade them separately rather than applying a breaking
automatic fix during this release. The public legal pages do not by themselves
make unsolicited scraped-lead outreach compliant with Google's Gmail policy;
that verification blocker remains. Browser-based visual automation was
unavailable in this session, so release confidence came from production builds,
independent static review, media checks, and local/public HTTP smoke checks.

## 2026-07-14 — Google OAuth Public Information Pages

**Role:** Frontend, Security, and Documentation Agent

**Delivered:** Added public, statically rendered Tuvi pages for the Google OAuth
application homepage (`/google-workspace`), Privacy Policy (`/privacy`), and
Terms of Service (`/terms`). The app page identifies the integration as **Tuvi
Outreach**, describes the single `gmail.send` permission, links to support and
both legal documents, and explicitly states that the app does not read inboxes,
contacts, attachments, history, or Drive files. The privacy page documents
consultation, voice, outreach, OAuth-token, delivery-metadata, sharing,
retention, revocation, and Limited Use practices. The terms prohibit spam,
unsolicited bulk commercial email, and Gmail-limit circumvention. Footer links
now resolve to the legal routes, and canonical metadata uses
`https://tuvisolutions.com`.

**Checks Run:** `npx tsc --noEmit --pretty false` — passed. Production-equivalent
Node 22 `next build` — passed, including lint/type validation and static
generation of all three routes. A local production server returned HTTP 200 for
`/google-workspace`, `/privacy`, and `/terms`; the privacy HTML contained its
canonical URL, `gmail.send` disclosure, and Google API User Data Policy link.
The host's unsupported Node 23 build failed in Webpack; retrying with the
production Node 22 runtime passed. A Docker build could not be completed because
Docker Desktop's storage was full; no Docker pruning was performed.

**Business Value / Plan Fit:** Supplies accurate, same-domain public pages for
the OAuth consent-screen fields and makes Tuvi's narrow Gmail access visible to
mailbox owners. This supports provider verification while preserving Phase 1's
human-approved outreach boundary.

**Risks / Follow-ups:** The deployment status in this historical entry is
superseded by the website release documented above; these pages are now live.
External Gmail verification is not ready: the repo has no Tuvi OAuth
start/callback or pre-consent disclosure flow, refresh tokens are configured via
server environment rather than a verified encrypted-at-rest secret store, and
the current scraped-lead workflow has suppression but no prior-consent evidence.
Google's Workspace API policy prohibits unsolicited commercial email and using
multiple Gmail accounts to bypass limits. Prefer an organization-owned Internal
Workspace app for Tuvi-owned mailboxes, and do not submit a cold-outreach use
case as compliant. The owner should also confirm the legal entity name,
jurisdiction, and address with counsel before publication.

## 2026-07-14 — Durable Eight-Hour Outreach Pacing

**Role:** Backend, Security, Test, and Documentation Agent

**Delivered:** Replaced the fixed two-second in-process email loop with
PostgreSQL-backed pacing for the verified-lead outreach workflow. Each account's
40-attempt allowance is now divided into durable slots across a rolling
eight-hour window, with a persisted cryptographically sampled 2–5 minute jitter.
A singleton database gate enforces the same minimum velocity guard across
account transitions and concurrent callers. Delayed workers re-anchor remaining
slots instead of sending catch-up bursts. Slot 40 retains the existing 24-hour
per-account cooldown.

Each durable job activation now crosses the Gmail/Zoho HTTPS provider boundary
at most once, releases its worker lease, and requeues for the database
`available_at`. Provider/delivery errors remain terminal so a bad credential
cannot drain future slots; existing delivery leases still fail ambiguous
provider outcomes closed. The transactional
OCR `verified`, human profile review, published demo, approved campaign,
provenance, recipient, suppression, and prior-send checks were preserved. Local
and template config now use `OUTREACH_SEND_WINDOW=8h`,
`OUTREACH_SEND_JITTER_MIN=2m`, and `OUTREACH_SEND_JITTER_MAX=5m` with
`OUTREACH_EMAILS_PER_ACCOUNT=40`. Sending remains disabled.

**Checks Run:** `go test ./backend/...` — 145 passed across 42 packages;
`go vet ./backend/...` — passed; race checks for outreach, email providers, and
jobs — 35 passed; focused pacing/config/provider/job tests — passed; OpenAPI YAML
parse and `git diff --check` — passed. Migration 24 applied, rolled back, and
reapplied on isolated PostgreSQL 16; its partial-account selection and
account-rotation SQL were also exercised. No provider call or real email send
was made.

**Business Value / Plan Fit:** This makes Phase 1 approved outreach restart-safe
and deliberately low velocity without tying up the only worker. It preserves
the existing human approval boundary and auditable per-account send sequence.

**Risks / Follow-ups:** Forty is a maximum of reserved attempts, not a guarantee
of 40 Gmail-accepted or delivered messages when eligible leads are exhausted or
provider outcomes are skipped/unknown. Forty literal 2–5 minute gaps cannot span
eight hours, so the implemented policy uses ~12-minute slots with 2–5 minute
jitter (normally ~9–15 minute on-time gaps). This is a deliverability control,
not a promise of inbox placement. Migration 24 and the code are local only;
production deployment/migration requires explicit approval and sending must stay
disabled through rollout.

## 2026-07-14 — Google Workspace Outreach and Provider Environment Deployment

**Role:** Backend, DevOps, Security, Test, and Documentation Agent

**Delivered:** Added Google Workspace Gmail API delivery to the existing durable
outreach account pool without removing Zoho. Each mailbox uses a stable key,
primary mailbox identity, optional verified send-as address, OAuth client, and
per-mailbox offline refresh token. Delivery uses Google's fixed HTTPS OAuth and
`users.messages.send` endpoints with MIME/base64url messages, cached access
tokens, redirect rejection, header/address validation before quota claims, and
the existing PostgreSQL 40-attempt/24-hour rotation ledger. Invalid enabled
mailbox, sender-header, or redirect configuration now fails during loading or
provider construction instead of consuming quota or leaving outreach silently
unavailable. Updated the environment templates, OpenAPI, ADR, service notes,
deployment plan, detailed workflow runbook, and architecture brief/Word copy.

Stored the supplied Places, Apollo, optional SerpAPI, and Hugging Face values in
the ignored local `.env` and protected VM `/opt/tuvi/env/ingestion.env`; both the
local env and VM env files remain mode `0600`. No secret value was printed or
committed. Google Workspace account JSON remains empty because the required
mailbox OAuth details were not supplied. Email stays disabled, OCR stays
disabled, and no provider-validation call was made.

Published commit `59d8cd4` to `origin/phase1_03/backend` and deployed that exact
release to `/opt/tuvi/MonoRepo`. Took a custom-format PostgreSQL backup, ran the
migration command (all 23 migrations already current), rebuilt and recreated
the API, Go worker, and scrape worker, then recorded `59d8cd4` in
`/opt/tuvi/current-release`. The first post-migration Compose invocation left
those three services stopped; verification caught that before the release was
accepted, and an explicit recreate restored all three successfully.

**Checks Run:** `go test ./backend/...` — 138 passed across 42 packages;
`go vet ./backend/...` — passed; email-provider race test — passed; Go command
build, Python syntax checks, OpenAPI/Compose YAML parsing, Compose config, secret
scan, and git diff checks — passed. The Word artifact passed OOXML ZIP,
structure, heading, section, content, and accessibility checks with zero
accessibility findings; visual rendering remains unavailable because
LibreOffice/`soffice` is not installed. On the VM, API/worker/scrape-worker are
running with zero restarts, logs contain no fatal patterns, and the API,
website, demo, and voice readiness URLs return HTTP 200. PostgreSQL reports
23/23 migrations, zero scrape jobs/request usage, zero running OCR claims, zero
email attempts, and zero email-account quota rows.

**Backups / Rollback:** Source backup:
`/opt/tuvi/MonoRepo.prev-20260715T043742Z`. Database backup:
`/opt/tuvi/backups/postgres/monorepo-pre-59d8cd4-20260715T043742Z.dump`.

**Risks / Follow-ups:** Before enabling real outreach, enable the Gmail API in
the Google Cloud project, configure the OAuth consent/admin trust, and provide
the mailbox email, OAuth client ID/secret, and an offline `gmail.send` refresh
token for every mailbox. Keep stable account keys across credential rotation.
The supplied API keys were exposed in chat and should be rotated after this
deployment, with the replacements installed in both protected env locations.

## 2026-07-14 — Architecture Changes Word Brief

**Role:** Documentation Agent

**Delivered:** Created `docs/ARCHITECTURE_CHANGES.docx` from the concise
architecture summary. The Word document uses a compact business-brief layout
with a masthead, seven before/current architecture sections, a production safety
callout, consistent Word heading styles, and restrained header/footer furniture.

**Checks Run:** The canonical document renderer was attempted but could not run
because LibreOffice/`soffice` is not installed. Following the document workflow's
approved fallback, the OOXML package passed ZIP validation, a structural and
content audit confirmed Letter geometry, margins, styles, headings, all required
facts, and absence of secret-like values, and the packaged accessibility audit
reported zero findings. No code tests were required.

**Business Value / Plan Fit:** Provides a short, shareable Word handoff of the
most consequential Phase 1 workflow and production architecture changes without
requiring stakeholders to read the implementation log or operator runbook.

## 2026-07-14 — Concise Architecture Change Summary

**Role:** Documentation Agent

**Delivered:** Added `docs/ARCHITECTURE_CHANGES.md`, a short before/missing/now
summary of the seven most consequential changes: durable city scraping,
Places-first Apollo enrichment, explicit OCR states, provenance-bound review
gates, secure demo access, PostgreSQL-managed email quotas, and isolated
production persistence.
The note also states the current safety boundary so deployment is not mistaken
for enabled scraping, OCR, or outreach.

**Checks Run:** Documentation-only review and `git diff --check`; no code tests
were required.

**Business Value / Plan Fit:** Gives operators and stakeholders a fast,
non-technical handoff for the Phase 1 lead-to-demo-to-outreach architecture
without duplicating the full runbook.

## 2026-07-14 — Isolated Monorepo Database and VM Workflow Deployment

**Role:** Backend, DevOps, Security, and Documentation Agent

**Delivered:** Published the durable scrape/OCR/outreach release to
`origin/phase1_03/backend` and deployed application commit `ffdf1e6` to the
Tuvi VM. The VM's available managed-PostgreSQL credential belongs to the
restricted SustainabilityWise role and cannot administer roles or databases,
so that cluster was left untouched. Instead, the existing dedicated Tuvi
PostgreSQL 16 instance now contains a separate `monorepo` login role and
`monorepo` database. The role owns only that database and has no superuser,
database-creation, role-creation, replication, or row-security-bypass powers.
A generated 64-character hexadecimal password was never printed: it is stored
in the ignored local `.env` and `/opt/tuvi/env/monorepo.env`, both mode `0600`.

Took a fresh pre-cutover custom-format backup, restored the existing Tuvi data
into the isolated database, and applied all 23 migrations. The source held no
users, restaurants, profiles, demos, campaigns, or reservations and one
completed job record; those counts were preserved. Added a loopback-only VM
PostgreSQL binding on port `15432` and verified the local `.env` connection
through an SSH tunnel. The database port is closed publicly.

Built and deployed the updated API, Go worker, durable scrape worker, OCR image,
restaurant template, and voice agent. API, website, docs, demo template, voice
health, and browser-voice readiness all return HTTP 200. PostgreSQL and Redis
are healthy; every long-running Tuvi service is running. The OCR image is
available but no OCR container is scheduled. Email remains disabled, and the
new `/opt/tuvi/env/ingestion.env` deliberately has empty Places, Apollo, and OCR
provider keys. No Melbourne scrape, provider request, email attempt, or OCR
claim was triggered during deployment.

**Checks Run:** `git diff --check`; YAML parsing; local and VM Compose config
validation; local Go command build; VM Docker builds (including the Next.js
production build/type check); migration execution; database ownership/role
flags/migration-count checks; local SSH-tunnel login; container-state/log
inspection; and public HTTP health checks. Per the user's earlier instruction,
no automated test suites were run. The first VM build exposed one unused Go
import before cutover; commit `ffdf1e6` removed it, after which all images built.

**Backups / Rollback:** Pre-cutover database backup:
`/opt/tuvi/backups/postgres/tuvi-pre-monorepo-20260714T215600Z.dump`. Post-
migration baseline:
`/opt/tuvi/backups/postgres/monorepo-post-migrate-20260714T221326Z.dump`.
Previous source tree:
`/opt/tuvi/MonoRepo.prev-20260714T220742Z`. The VM records the active release in
`/opt/tuvi/current-release`.

**Risks / Follow-ups:** Configure workload-restricted Google Places and Apollo
keys plus one approved OCR vision provider before creating a city job. Keep OCR
and email disabled until their controlled enablement steps are followed. The VM
still has no automated PostgreSQL backup schedule; add one separately with a
retention and restore policy.

## 2026-07-14 — Durable City Scrape, OCR, Review, and Outreach Workflow

**Role:** Backend, Security, DevOps, and Documentation Agent

**Delivered:** Implemented the requested workflow as a durable, fail-closed
pipeline. Migrations `000015`–`000023` add city scrape jobs/grid cells/candidate
checkpoints, explicit OCR states and claims, automatic draft provenance, audited
human review gates, PostgreSQL email-account quota state, delivery sequences,
recoverable worker leases, and immutable per-send recipient attribution for
delivery and unsubscribe records. The private scrape-job API now creates and
reports city work. The worker discovers through Google Places first, dynamically
subdivides saturated grid cells, deduplicates Place IDs, then invokes Apollo only
for missing owner/work-email details; an Apollo no-match retains the valid Places
lead. Places and Apollo share one persisted
500-request window; the job resumes from its saved cell/page/candidate state
after 24 hours and starts recurring coverage cycles to find newly listed Places.
Provider-capped leaves subdivide to depth 12 by default and remain visibly
`coverage_incomplete` rather than falsely completing a city pass.

OCR now transitions through `pending`, `running`, `verified`, `no_images`, and
`failed`. It refreshes Google Places photo resources just in time, persists
neither expirable resource names nor short-lived photo URIs, accepts only trusted Google-hosted
fallback images in unattended operation, and never treats `no_images` as
verified. Failed unchanged inputs retry after a 24-hour cooldown up to three
attempts, including a hard cap on repeatedly abandoned stale claims. A verified
claim transactionally queues reusable idempotent `lead.prepare`,
which creates only a demo draft and outreach campaign draft. Real delivery
requires separate audited profile approval, demo publication, campaign
approval, and an internal administrator starting bulk outreach.

Bulk outreach now uses durable PostgreSQL account state: at most 40 reserved
attempts per account cycle, automatic account rotation, a 24-hour `available_at`
cooldown, automatic continuation, global and per-account sequences, and
delivery leases that fail ambiguous outcomes closed as `send_unknown`. External
email uses HTTP(S) provider APIs only. Production rendering validates the
canonical API, demo, marketing, and presentation links before any provider
call. Each confirmed send retains the delivery attempt's immutable normalized
recipient and global sequence; legacy unsubscribe tokens use an audited current-
address fallback so their opt-out links remain usable. A shared per-restaurant
transaction lock serializes import, preparation, review, campaign state, and
delivery finalization. The obsolete per-campaign `send-step` route and legacy one-shot ingestion
entrypoint are retired so they cannot bypass quota/request ledgers. Added ADRs,
OpenAPI coverage, Compose services/images, environment examples, and the
operator runbook at `docs/runbooks/lead-scrape-ocr-outreach.md`.

The public reservation flow now binds an idempotency key to the complete request,
uses each supported Australian city's IANA timezone, handles split and overnight
hours, returns typed conflicts, and continues to create only `pending` requests.

**Why / Business Value:** This supplies the Phase 1 lead-to-demo-to-reviewed
outreach loop while controlling provider rate limits, preventing duplicate
scrapes/sends, preserving work across restarts and cooldowns, and retaining the
required human approval before contacting a real lead.

**Checks Run:** Applied `gofmt` to changed Go files; `git diff --check`
passed; Ruby parsed `docs/openapi/openapi.yaml` and
both Docker Compose YAML files successfully. Independent read-only static
reviews of the Go workflow, Python workflow, migrations, Compose, OpenAPI, and
runbook were used to identify and close correctness/safety discrepancies.
Context7 supplied the current official Google Places Photo Media contract.
Per the user's instruction, no tests, builds, Python execution, migrations,
provider calls, containers, or smoke checks were run.

**Deployment / Risk Boundary:** This work exists only in the local worktree.
No SSH deployment, production migration, Melbourne scrape trigger, OCR run, or
real email send occurred. Before deployment, back up PostgreSQL, reconcile the
legacy states listed in the runbook, install migrations in order, configure
restricted provider credentials, keep email disabled, and re-record every human
review gate reset by migration `000018`. Deployment, migrations, and real
outreach remain approval-gated external actions.

The final static consistency pass also made administrator regeneration of an
automatic demo/campaign draft atomically refresh its verified OCR and current
profile provenance, preventing an expired-token refresh from retaining stale
review evidence.

## 2026-07-11 — Tuvi Restaurant Video Delivery Optimized

**Role:** Frontend / DevOps Agent

**Delivered:** Diagnosed the live restaurant-services page and found that it
eagerly downloaded two below-the-fold MP4s plus PNG posters, totaling about
40 MB, from the VM at roughly 0.5 MB/s. Transcoded the videos to web-ready
H.264 720p/24 fps MP4s with fast-start metadata and no unused audio, converted
the posters to compact JPEGs, and reduced the combined media payload to about
3.7 MB. Added immutable caching for the versioned media paths. Early iterations
used viewport-managed playback and then single-active-video coordination, but
the final user direction was to never pause either clip. Replaced all viewport
and active-video logic with immediate full-file fetches. The visible loading
message was subsequently removed and the four optimized assets were uploaded as
public, immutable objects under the dedicated
`tuvi/public/restaurant-services/v2/` prefix in the VM-configured DigitalOcean
Spaces bucket. The website now uses direct CDN URLs, native eager preload, and
an explicit post-hydration `play()` call; both videos autoplay and loop without
application-driven pauses. `Tap to start video` remains available only when
mobile autoplay policy actually rejects muted playback. No storage credentials
were added to the website or repository. Initial commit `a667c74` plus follow-up
commits `4e3431f`, `af96fd0`, `fb3d670`, `32a4b78`, `71a3cf1`, `1259950`,
`69467a7`, and final website release `07ebe32` were
pushed and deployed by rebuilding only the `tuvi-website` service; API, worker,
PostgreSQL, migrations, Caddy, and the pending managed-database cutover were not
restarted.

**Why / Business Value:** The two sales demos now become playable in a few
seconds on the same VM connection instead of competing over approximately
40 MB of eager downloads for 17–24 seconds. The page initially loads at most the
nearby preview, defers the second video until the visitor approaches it, and
still gives phone users a clear play action when browser policy blocks autoplay.
This improves the Phase 1 restaurant sales page without changing application or
database behavior.

**Tests / Checks Run:** TypeScript `tsc --noEmit` passed; both optimized MP4s
decoded end-to-end with FFmpeg without errors; `git diff --check` passed. The
clean Node 22 production Docker build compiled, type-checked, generated all
pages, and recreated the website container successfully. VM loopback returned
`200` for the page and `206` for byte-range video requests. Public Caddy checks
returned `200`/`206`, correct `video/mp4` and `image/jpeg` types, exact content
lengths, `Accept-Ranges`, and `Cache-Control: public, max-age=31536000,
immutable`. Full public downloads measured 3.76 seconds and 2.51 seconds.
In-app browser verification confirmed deferred media attachment, ready-state 4,
active playback after scrolling, no media errors, and no browser console errors.
The final live verification confirmed no loading label is rendered; both videos
use direct Spaces CDN URLs, become fully buffered, remain unpaused below the
fold, and advance together at real time without media or browser-console errors.
All four CDN objects return the correct MIME types, exact lengths, byte ranges,
public-read access, and one-year immutable cache headers. Warm-edge full-video
downloads measured about 0.15 seconds each versus roughly 2.5–3 seconds from the
VM in the comparison run.
Context7 was used to confirm Next.js public-asset cache behavior and current
FFmpeg encoding guidance.

**Build Notes / Rollback:** The workstation's pre-existing Node 23 cache still
fails Next's build with `Unexpected end of JSON input`, and the repo has no
ESLint 9 flat config; neither failure is caused by this change. The authoritative
clean Node 22 VM build passed. The pre-deploy source tree is preserved at
`/opt/tuvi/MonoRepo.prev-20260711184731`; rebuilding `tuvi-website` from that
tree restores the previous version. VM release records identify the website as
`07ebe32`, while the rest of the stack remains `f18ed02` and database cutover
commit `ed36120` remains pending.

**Risks / Follow-ups:** The assets use a dedicated Tuvi prefix but currently
share the existing Spaces bucket and upload credential configured on the VM;
provision a dedicated Tuvi bucket/key when stronger storage-level isolation is
needed. Runtime delivery is public and credential-free. A browser or operating
system may still suspend media in a hidden tab or power-saving mode; application
code no longer pauses it.

## 2026-07-11 — Managed PostgreSQL Cutover Prepared

**Role:** DevOps / Backend / Security Agent

**Delivered:** Audited the VM database topology without exposing credentials.
Confirmed Tuvi currently uses its dedicated PostgreSQL 16 container/database
`tuvi_api`, while SustainabilityWise uses managed PostgreSQL 18 database
`sustainability_wise` with role `sw_api`. Confirmed the managed cluster already
contains isolated database/role `monorepo`; that role is non-superuser and has
no database/role creation rights or SustainabilityWise role membership. Updated
VM Compose so migrate/API/worker honor the supplied `DATABASE_URL`, committed
and pushed as `ed36120`, synced the pending config to the VM, and created a
root-only placeholder for the managed connection URI.

**Data Safety:** No SustainabilityWise/EcoAudit/SolarSense schema or data was
read or changed. The live Tuvi API and worker still use the original source DB
and remain healthy. A fresh source-only custom-format dump was saved at
`/opt/tuvi/backups/postgres/tuvi_api-pre-managed-2026-07-11-171439.dump` with
mode `0600`.

**Checks Run:** Source identity/version/extensions/migrations were audited;
migrations 1–14 are present, PostgreSQL is 16.14, `pgcrypto` is installed, and
the database has no restaurant/user records plus one completed job. Managed
PostgreSQL 18.4 and private-network connectivity from the Tuvi API container
were verified. Compose validation passed with an external TLS URL. The public
restaurants API remains `200`; API and worker are running with zero restarts.

**Blocker / Next Step:** The VM does not contain the password for the existing
managed `monorepo` role, and the SustainabilityWise role correctly cannot create
or alter roles/databases. Complete `/opt/tuvi/env/managed-db.env` with the
DigitalOcean private URI, then quiesce writes, take a final dump, restore and
validate counts/migrations, update `stack.env`, restart, and run isolation/smoke
checks. Logical database isolation shares the managed cluster's resource and
failure domain; physical independence would mean retaining the current Tuvi
Postgres container.

## 2026-07-11 — Full Tuvi VM Redeploy And Corporate Site Cutover

**Role:** DevOps / Frontend Agent

**Delivered:** Fetched `origin/phase1_03/backend` at `95c776b`, created a clean
release that excluded unrelated local worktree changes, backed up PostgreSQL and
the previous VM source tree, rebuilt every service in
`infra/docker/docker-compose.vm.yml`, and deployed the redesigned
`tuvi-website/app` as the canonical site. Restored the corporate website service
to the committed VM Compose definition, tracked the Swagger UI entrypoint, and
mapped Caddy root/www traffic to `127.0.0.1:15174` while preserving API, voice,
demo, catalog, and unrelated host routes. The reproducible deployment support
was committed and pushed as `f18ed02`, which is the VM's recorded release.

**Why:** The latest remote release redesigned the corporate website, but the VM
Compose source no longer contained the service definition needed to rebuild it.
The deployment needed all MonoRepo services rebuilt and Caddy aligned with the
new canonical site.

**Business Value:** Tuvi's current corporate design and restaurant-services page
are now publicly served from the VM, with the API, worker, voice agent, demo
template, standalone catalog, database, Redis, and API documentation operating
behind the existing production routing.

**Plan Fit:** Advances Phase 1 production deployment readiness and keeps the
full VM stack reproducible from source instead of relying on an orphaned website
container.

**Tests / Checks Run:** `go test ./backend/...` — 115 tests passed;
`npm --prefix apps/restaurant-services-catalog run build` — passed; VM Docker
Compose config validation and full `up -d --build` — passed; migration container
exited `0`; all eight long-running containers are running; Caddy validation and
reload passed. Public root, www, restaurant services, API root, Swagger UI,
OpenAPI, public restaurants API, voice readiness, and demo endpoints all returned
`200`.

**Backups / Rollback:** PostgreSQL backup:
`/opt/tuvi/backups/postgres/restaurant_platform-2026-07-11-165111.sql`.
Previous source tree: `/opt/tuvi/MonoRepo.prev-20260711165431`. Previous Caddy
config: `/etc/caddy/Caddyfile.prev-20260711165727`.

**Risks / Follow-ups:** The workstation's Node 23 build path still produces a
webpack JSON parsing failure; the supported clean Node 22 Docker build passed on
the VM. Context7 was used to confirm clean `npm ci`/fresh `.next` Docker build
guidance. The root VM source directory must remain executable/readable by Caddy
so static API docs continue to work.

## 2026-07-08 — Tuvi VM Stack Deployed

**Role:** DevOps / Frontend / Backend Agent

**Delivered:** Implemented the VM deployment stack for `tuvisolutions.com` on
`root@170.64.154.143`. Added VM Docker Compose, catalog and template Dockerfiles,
catalog Nginx config, Caddy route example, Docker build excludes, and production
template route fixes. Synced the committed source to `/opt/tuvi/MonoRepo`,
created `/opt/tuvi/env` production env files with generated internal secrets,
started the Tuvi Compose project, repaired voice-agent data-volume ownership,
and appended validated Caddy routes for `tuvisolutions.com`,
`www.tuvisolutions.com`, `api.tuvisolutions.com`, `voice.tuvisolutions.com`, and
`demo.tuvisolutions.com`.

**Why:** Raj approved moving TuviSolutions.com off the older Vercel/presentation
site and onto the VM, with `apps/restaurant-services-catalog` as the canonical
public Tuvi website.

**Business Value:** The VM is now ready to serve the current restaurant services
catalog, API, worker, demo template, and voice service behind Caddy without
disturbing existing Tilnest, SustainabilityWise, or n8n services.

**Plan Fit:** Completes the Phase 1 deployment groundwork for the Tuvi marketing
site and related service endpoints. Public traffic still requires DNS cutover.

**Tests / Checks Run:**
- `rtk make restaurant-services-catalog-build` — passed locally
- `rtk make test` — backend Go tests passed locally
- `docker compose --env-file /opt/tuvi/env/stack.env -p tuvi -f infra/docker/docker-compose.vm.yml up -d --build`
  — built and started VM stack
- VM loopback `GET/HEAD http://127.0.0.1:15173` — catalog returned `200 OK`
- VM loopback `HEAD /media/qr-ordering-kitchen-v2.mp4` and
  `HEAD /media/rewards-reception-v3-pro.mp4` — both returned `200 OK` with
  `video/mp4`
- VM loopback `HEAD http://127.0.0.1:18080/api/public/v1/site/restaurants` —
  API returned `200 OK`
- VM loopback `HEAD http://127.0.0.1:13000` — template returned `200 OK`
- VM loopback `GET http://127.0.0.1:18000/readyz/browser` — voice service is
  running but reports missing voice provider keys
- `caddy validate --config /etc/caddy/Caddyfile` — passed after Tuvi route append

**Risks / Follow-ups:** DNS still points `tuvisolutions.com` and `www` at
Vercel, and the Tuvi subdomains are not yet configured. Voice readiness requires
real `DEEPGRAM_API_KEY`, `OPENAI_API_KEY`, `CARTESIA_API_KEY`, and
`CARTESIA_VOICE_ID` in `/opt/tuvi/env/voice.env`. GitHub HTTPS push failed with
an RPC 400 and SSH push was not authorized, so the VM was deployed by rsync from
the local committed branch.

## 2026-07-08 — VM Deployment Plan Draft

**Role:** DevOps / Documentation Agent

**Delivered:** Added `docs/runbooks/vm-deployment-plan.md`, a VM deployment
plan based on the repo's current runnable services and Compose setup. The plan
covers the main Go API, worker, migrations, PostgreSQL, voice agent, Redis,
restaurant services catalog, optional Tuvi Next site, restaurant demo template,
automation jobs, reverse proxy/TLS, env files, backups, rollback, smoke checks,
and open decisions.

**Why:** Raj asked for a complete plan to deploy all services on a VM and asked
that the existing setup be considered before planning.

**Business Value:** Creates a concrete deployment path from local Phase 1
services to one VM, while making clear which pieces are already containerized
and which pieces still need VM Compose/proxy work.

**Plan Fit:** Supports Phase 1 deployment/runbook work by turning the existing
Docker stack into a fuller VM production plan with no intentional service gaps.

**Tests / Checks Run:** Inspection only. Reviewed `AGENTS.md`, `git status`,
`README.md`, `Makefile`, `docs/SERVICES.md`, Phase 1 docs, Docker Compose,
backend Dockerfile/config, stack env example, startup scripts, Tuvi website
runtime docs, restaurant catalog docs, template env, and voice agent Docker/env
docs. Checked local SSH config for a VM alias; none was configured.

**Risks / Follow-ups:** Live VM audit is still pending because this workstation
does not have a usable VM SSH alias in `~/.ssh/config`. The plan includes the
exact audit commands to run once the VM host/deploy user is known.

## 2026-07-08 — Local Catalog Work Rebased on Phase Branch

**Role:** Frontend / Documentation Agent

**Delivered:** Preserved Raj's local `apps/restaurant-services-catalog` work by
committing it locally, creating a safety branch at `local/pre-pull-catalog-work`,
and rebasing the local `phase1_03/backend` branch on top of the latest fetched
`origin/phase1_03/backend`. Resolved the pull-blocking conflicts in the catalog
package files and session summary without replacing the video-enabled restaurant
services catalog.

**Why:** A direct `git pull` was blocked because local session docs and
untracked catalog files would have been overwritten by the incoming branch.

**Business Value:** The active phase branch now has the latest backend work plus
the restored restaurant services catalog assets, including the FAL-generated QR
ordering and rewards reception videos needed for the sales presentation flow.

**Plan Fit:** Keeps the Phase 1 restaurant services marketing surface aligned
with the backend branch while preserving local catalog iteration work for
outreach and demo conversations.

**Tests / Checks Run:**
- `rtk git fetch origin phase1_03/backend` — fetched latest remote branch state
- `rtk git rebase origin/phase1_03/backend` — completed after conflict
  resolution
- `rtk make restaurant-services-catalog-build` — Vite production build passed
- `rtk make test` — backend Go test suite passed
- `rtk rg -n 'qr-ordering-kitchen-v2|rewards-reception-v3-pro' apps/restaurant-services-catalog/src apps/restaurant-services-catalog/dist/index.html`
  — built catalog references both approved video files
- `rtk ls -lh apps/restaurant-services-catalog/public/media/*.mp4` — confirmed
  the current and fallback MP4 assets are present
- `rtk curl -I http://127.0.0.1:5174` — catalog dev server returned `200 OK`
- `rtk curl -s http://127.0.0.1:5174 | rtk rg -n 'qr-ordering-kitchen-v2.mp4|rewards-reception-v3-pro.mp4'`
  — served page references both approved video files
- `rtk curl -I http://127.0.0.1:5174/media/qr-ordering-kitchen-v2.mp4`
  — returned `200 OK` with `video/mp4`
- `rtk curl -I http://127.0.0.1:5174/media/rewards-reception-v3-pro.mp4`
  — returned `200 OK` with `video/mp4`

**Risks / Follow-ups:** The local commit is intentionally still ahead of
`origin/phase1_03/backend` by one commit and has not been pushed. Permanent
Cloudflare Pages deployment still requires Wrangler authentication.

## 2026-07-07 — Phase Branch Catalog Videos Preserved

**Role:** Frontend / Documentation Agent

**Delivered:** Pulled `origin/phase1_03/backend` into a local
`phase1_03/backend` branch and applied the safe local catalog changes on top:
root Makefile shortcuts for the restaurant services catalog, catalog README, and
public `.env.example`. Preserved the branch's FAL-generated restaurant feature
videos (`qr-ordering-kitchen-v2.mp4` and `rewards-reception-v3-pro.mp4`) and the
video-enabled catalog page.

**Why:** Raj asked to apply local changes on top of `phase1_03/backend` and
clarified that the target website is the catalog version with two FAL-generated
videos.

**Business Value:** The current branch now has the latest backend work and the
video-enabled restaurant services site, with simple root commands for local
startup and build verification.

**Plan Fit:** Keeps the Phase 1 restaurant services marketing surface aligned
with the active backend branch.

**Tests / Checks Run:**
- `rtk npm --prefix apps/restaurant-services-catalog install` — dependencies current; no vulnerabilities reported
- `rtk make restaurant-services-catalog-build` — initially failed on broken `lucide@0.547.1`; passed after pinning `lucide` to `0.547.0`
- `rtk make test` — backend tests passed
- `rtk curl -I http://127.0.0.1:5173` — catalog dev server returned `200 OK`
- `rtk curl -s http://127.0.0.1:5173 | rtk rg -n "qr-ordering-kitchen-v2.mp4|rewards-reception-v3-pro.mp4"` — served page references both FAL videos
- `rtk curl -I http://127.0.0.1:5173/media/qr-ordering-kitchen-v2.mp4` — returned `200 OK` with `video/mp4`
- `rtk curl -I http://127.0.0.1:5173/media/rewards-reception-v3-pro.mp4` — returned `200 OK` with `video/mp4`

**Risks / Follow-ups:** The local changes from the older restored catalog app
were not applied wholesale because they would replace the newer video-enabled
catalog implementation. `lucide` is pinned to `0.547.0` because `0.547.1`
installed locally without matching icon module files and broke the Vite build.

## 2026-07-06 — Tuvi Scheduler and Voice Agent Unified API Alignment

**Role:** Backend / Frontend / AI Workflow Agent

**Delivered:** Merged the remote unified Tuvi company consultation backend and
rewired the Tuvi corporate meeting scheduler and corporate voice assistant to use
the same main MonoRepo API:
`GET /api/v1/company/consultations/availability`,
`GET /api/v1/company/consultations/availability/check`, and
`POST /api/v1/company/consultations`. The scheduler now collects phone and sends
`web` source through a server-side Next proxy that keeps `TUVI_API_TOKEN` out of
the browser. The corporate voice assistant now asks for phone before booking and
sends `voice` source directly to the same unified API. The company consultation
success response now includes `prospect_phone`.

**Why:** Raj merged the Tuvi app into the main app and asked that all relevant
apps, including meeting scheduler and voice agent, call the new main endpoint.

**Business Value:** Website and voice-assisted consultation bookings now land in
one Phase 1 company consultation pipeline, reducing split-brain booking state and
making future dashboards, analytics, and follow-up workflows simpler.

**Plan Fit:** Advances P1-E05 reservation capture and P1-E08 voice assistant
integration by routing Tuvi web and voice consultation requests through one main
backend contract.

**Tests / Checks Run:**
- `rtk make test` — backend tests passed
- `rtk python3 -m py_compile voice-sales-agent/tuvi_api_client.py voice-sales-agent/bot.py` — passed
- `rtk npx tsc --noEmit` from `tuvi-website/app` — passed
- `rtk npx -p node@22 node ./node_modules/next/dist/bin/next build` from
  `tuvi-website/app` — passed after clearing stale `.next` output
- `rtk npm run build` from `tuvi-website/app` — passed

**Risks / Follow-ups:** Set the same `TUVI_API_TOKEN` in the main API, Tuvi
website server env, and corporate voice-agent env before trying bookings. The
unified API supports Google Calendar and SMTP through main backend configuration;
local defaults keep those providers disabled.

## 2026-06-30 — RTK Codex Initialization

**Role:** DevOps / Documentation Agent

**Delivered:** Installed and initialized RTK for Codex globally and locally.
Global Codex config now includes `/Users/rajchodisetti/.codex/RTK.md`, and this
repo now includes `RTK.md` with an `@RTK.md` reference from `AGENTS.md`.

**Why:** Raj requested RTK initialization, especially for Codex integrations, so
future Codex sessions get the token-optimized command instruction automatically.

**Business Value:** Reduces command-output context usage during future coding
sessions while preserving normal command behavior.

**Plan Fit:** Supports the local development tooling rules in `AGENTS.md` and
keeps agent command usage consistent across repo-local and global Codex contexts.

**Tests / Checks Run:**
- `rtk init -g --codex` — configured global Codex RTK files
- `rtk init --codex` — configured repo-local Codex RTK files
- `rtk init --codex --show` — verified global and local Codex RTK config
- `rtk read RTK.md` — verified repo-local RTK instructions

**Risks / Follow-ups:** Generic `rtk init --show` reports non-Codex hook status;
use `rtk init --codex --show` to verify Codex integration specifically.

## 2026-06-30 — Service and Workflow Inventory

**Role:** Planner / Documentation Agent

**Delivered:** Inspected the repo command surface, Phase 1/Phase 2 guides, Makefile,
Docker Compose, current backend packages, template frontend, and automation
README files to produce a business-prioritized service inventory and trigger
workflow map.

**Why:** Raj requested a clear view of which services exist or are planned, how
to start them locally, and which business workflows trigger downstream services.

**Business Value:** Clarifies the shortest path to running the sales MVP locally:
database, migrations, API, worker, seed data, demo template, and optional
automation. It also separates currently implemented services from planned Phase
1/Phase 2 services so execution can stay focused.

**Plan Fit:** Supports Phase 1 lead-to-demo-to-reservation sequencing and
identifies the Phase 2 orchestration services as future work after the Phase 1
sales loop is working.

**Tests / Checks Run:** Inspection only. Ran targeted reads of `AGENTS.md`,
`Makefile`, `README.md`, Phase 1/Phase 2 docs, Docker Compose, backend router,
app startup, worker startup, config, template README/package, and automation
README files. No code tests were run because no product code changed.

**Risks / Follow-ups:** `apps/web` is still a placeholder; the runnable demo
frontend currently lives under `template/`. Phase 2 docs are under `phase2/`
rather than `docs/phase2/`.

## 2026-06-24 — User Profile API (`GET /api/v1/user/me`)

**Role:** Backend Agent

**Delivered:** `GET /api/v1/user/me` for `restaurant_owner` and `developer` roles — returns full user profile (id, email, full_name, role, is_active, created_at, updated_at) plus linked restaurants with member_role. Parallel to existing `GET /api/v1/admin/me`. OpenAPI updated.

**Why:** Normal users need a dedicated profile endpoint with complete account details, not just JWT claims from `/api/v1/auth/me`.

**Tests / Checks Run:** `go test ./backend/...` — all passing

**Follow-ups:** Optional PATCH `/api/v1/user/me` for profile updates

## 2026-07-03 — Restaurant Services Catalog App

**Role:** Frontend Agent

**Delivered:** Added a new standalone Vite JavaScript app at
`apps/restaurant-services-catalog` for an interactive Tuvi restaurant services
catalog. The app uses the existing `presentation/index.html` service story as
content grounding, adds a premium visual system, scroll-driven motion sections,
interactive service selection, QR/rewards/voice/service modules, an ROI
estimator, Cloudflare Pages config, and locally stored fal-generated media
assets.

**Why:** Tuvi needs a polished services catalog that can be shared with
restaurant owners after demo outreach, showing the full platform package beyond
one personalized demo website.

**Business Value:** The catalog turns the service offering into a sales asset:
owners can see demo websites, QR ordering, rewards, AI receptionist,
reservations, outreach, content automation, and token-gated demos in one guided
experience.

**Plan Fit:** Supports Phase 1 sales positioning and demo/outreach workflows by
giving campaigns a richer Tuvi services destination after the restaurant-specific
demo link.

**Checks Run:**
- `npm install` in `apps/restaurant-services-catalog` — success, 0 vulnerabilities.
- `npm run build` — success.
- `curl` smoke checks for local and Cloudflare tunnel page/media assets — 200.
- Playwright screenshots for desktop/mobile hero and services sections.

**Risks / Follow-ups:** Cloudflare Pages direct deploy was blocked because
Wrangler is not authenticated in this shell. The app is currently hosted via a
Cloudflare quick tunnel backed by a local static server. Run
`npm run deploy` after `wrangler login` or with a Cloudflare API token to publish
to Pages.

## 2026-07-03 — Restaurant Services Catalog Polish

**Role:** Frontend Agent

**Delivered:** Polished the hosted restaurant services catalog by removing
visible AI-generation/provider labels from the UI, replacing the hero background
video with pointer/touch-reactive canvas motion graphics, and tightening the
desktop/mobile visual presentation after screenshot review.

**Why:** The catalog should read as a premium Tuvi-owned sales experience, not
as an asset-generation demo. The hero needed a lighter, interactive motion layer
without relying on a background video.

**Business Value:** Restaurant owners see a cleaner services story with fewer
technical labels and a more polished first impression, improving the catalog's
fit as an outreach and sales destination.

**Plan Fit:** Supports Phase 1 demo/outreach by improving the public-facing
services destination already hosted through the Cloudflare quick tunnel.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Source/dist scan confirmed the removed AI-generation/provider phrases are no
  longer visible in the app markup.
- Cloudflare tunnel smoke check for the hosted catalog — 200.
- Playwright screenshot review for desktop hero, mobile hero, and media section.

**Risks / Follow-ups:** The Cloudflare Pages permanent deploy remains blocked
until Wrangler is authenticated. The current hosted preview depends on the local
static server and Cloudflare quick tunnel staying up.

## 2026-07-03 — Restaurant Services Catalog Offer Revision

**Role:** Frontend Agent

**Delivered:** Updated the hosted catalog to focus only on restaurant-facing
Tuvi services. Renamed "Premium Demo Websites" to "Custom Premium Websites",
removed the scroll-driven "invisible lead to booked sales call" transformation
block, removed internal sales-pipeline/token-gated demo language from the
service catalog, and replaced the obsolete single generated loop with two new
fal-generated cinematic service-flow videos: QR table ordering routed to the
kitchen, and reception ordering with QR rewards. Follow-up cleanup removed
client-facing production/style language such as "cinematic product moments",
changed the section to "Guest Experience", changed the nav label from "Media" to
"Experience", neutralized public media filenames, and removed the public media
manifest so provider/generation details are not exposed in the hosted page. A
later client-facing cleanup removed the value estimator, removed the two extra
image tiles from the QR/rewards section, simplified the top-left Tuvi wordmark,
removed vague catalog/guest-experience phrasing, tightened hero/section spacing,
and upgraded service cards with a glass treatment plus color highlight on hover,
focus, touch, and selected state.

**Why:** Raj clarified that the catalog should sell what Tuvi can do for
restaurants, not describe Tuvi's internal lead-acquisition and sales process.

**Business Value:** Restaurant owners now see concrete guest-facing outcomes:
custom websites, QR ordering, rewards, voice agents, reservations, promotions,
menu/photo automation, and owner dashboards. The new videos make the QR ordering
and rewards flows easier to understand quickly.

**Plan Fit:** Improves the Phase 1 services destination that can be linked from
demo-site conversations and used as a Tuvi walkthrough for restaurant owners.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Source/dist phrase scan confirmed removed internal labels and old video
  references are gone.
- Hosted page-source scan confirmed provider/generation/prompt/style wording is
  not present in client-facing HTML.
- Source/dist scan confirmed estimator code, estimator markup, old vague labels,
  and unused image references are gone.
- Cloudflare tunnel smoke checks for the hosted page and both new MP4 assets —
  200; old generated/provider-named asset paths — 404.
- Playwright screenshots for desktop hero, mobile hero, services section, and
  QR/rewards media section.

**Risks / Follow-ups:** Permanent Cloudflare Pages deploy still needs Wrangler
authentication. The current hosted URL depends on the local static server and
Cloudflare quick tunnel staying online.

## 2026-07-03 — Restaurant Services Catalog Tuvi Solutions Update

**Role:** Frontend Agent

**Delivered:** Updated the hosted catalog to use "Tuvi Solutions" consistently
instead of standalone "Tuvi", replaced the services grid with a sticky parallax
service-card stack, and added animated icon/symbol treatments to every service
card using the existing Lucide icon system plus CSS motion. Cleaned the services
section copy so it sells restaurant outcomes instead of explaining the scroll
interaction. Tightened the landing hero first fold from mobile screenshot
feedback by making the hero height content-driven, brightening the restaurant
background layer, softening the dark overlay, adding topbar separation, hiding
the mobile scroll cue, and trimming the lower hero gap.
Added viewport-managed video loading/playback so large MP4 walkthroughs use
poster frames initially, load only when near view, pause when offscreen, and
resume only when visible to reduce mobile buffering and stutter through the
Cloudflare quick tunnel.

**Video Work:** After fal billing was recharged, generated and downloaded two
new service videos without deleting the previous files. The approved QR ordering
video remains `qr-ordering-kitchen-v2.mp4`. The rewards video was regenerated
again with Kling V3 Pro using a tighter script focused on scan, order
confirmation, points added, and redeem options, then wired as
`rewards-reception-v3-pro.mp4` with `rewards-reception-v3-pro-poster.png`. Older
rewards files remain in `public/media` for fallback/reference.

**Why:** Raj asked the catalog to present the company as Tuvi Solutions, make the
services area feel more premium and progressive, and produce clearer service-flow
videos without removing the current assets.

**Business Value:** Restaurant owners now see one premium service module at a
time with clearer visual emphasis, plus clearer service-flow videos for table QR
ordering to kitchen operations and counter QR rewards.

**Plan Fit:** Keeps the hosted Tuvi Solutions sales destination aligned with the
demo/outreach path while preserving reusable media assets for future iteration.

**Checks Run:**
- `npm run build` in `apps/restaurant-services-catalog` — success.
- Rebuilt the live-served catalog `dist` and confirmed the Cloudflare tunnel
  serves the updated CSS bundle with the brighter background, mobile topbar, and
  tighter hero spacing rules.
- Confirmed the live tunnel serves deferred video tags (`data-src`,
  `preload="none"`) and the JS bundle contains the managed video playback logic.
- Source scan confirmed no standalone "Tuvi" remains in the catalog source.
- Source scan confirmed removed internal labels such as AI-generated wording,
  prompts, estimator text, and scroll-driven/catalog labels remain absent.
- Local and Cloudflare tunnel smoke checks for page, `qr-ordering-kitchen-v2.mp4`,
  `rewards-reception-v3-pro.mp4`, and `rewards-reception-v3-pro-poster.png`
  returned 200.
- Playwright screenshot confirmed the latest walkthrough labels, V2 posters, and
  captions render cleanly in the media section.

**Risks / Follow-ups:** The first pre-recharge ordering request stayed queued at
position 0, so a fresh ordering retry was used for the live V2 file. Permanent
Cloudflare Pages deploy still requires Wrangler authentication.

## 2026-06-24 — Scraped Restaurant Data Schema and Import

**Role:** Backend Agent

**Delivered:** Migration `000008_restaurant_profiles_menus` with `restaurant_profiles`, `menus`, `menu_items`, `restaurant_reviews`, and `restaurant_data_imports` tables; `seed-restaurants-data` command and `make seed-restaurants-data` target; imported `data/restaurants_data.json` into PostgreSQL (8 unique restaurants, 3 duplicate `google_place_id` rows skipped).

**Why:** Scraped SerpAPI restaurant payloads (menus, reviews, contact, location, images) need durable storage before P1-011 profile APIs and demo payload builder can consume them.

**Business Value:** Sales and demo workflows can now query real scraped restaurant profiles, menus, and reviews locally instead of relying on the raw JSON file.

**Plan Fit:** Implements the P1-011 database shape (profiles, menus, menu items) plus review storage; unblocks profile CRUD APIs and demo payload builder polish.

**Tests / Checks Run:**
- `make migrate-up` — success
- `make seed-restaurants-data` — 8 imported, 3 skipped
- SQL verification — menu item and review counts per restaurant confirmed

**Risks / Follow-ups:**
- No profile/menu HTTP APIs yet; data is import-only
- Re-import is idempotent by `google_place_id` but replaces menu items and reviews
- Add profile API routes and link demo builder to imported data

## 2026-06-17 — P1-E01 Backend Foundation

**Role:** Backend Agent

**Delivered:** Initialized the repository as a Go module and added the Phase 1
backend foundation: API, worker, and migration commands; typed environment
configuration; structured logging; PostgreSQL pool wiring with `pgxpool`;
health/readiness endpoints; request ID, access logging, recovery, and CORS
middleware; initial SQL migration; internal migration runner; in-memory job
queue; tests; README; Makefile; `.env.example`; and a foundation-stack ADR.

**Why:** P1-E01 is the required platform base for all Phase 1 product work. The
project needed a runnable backend, local command surface, safe config behavior,
database readiness path, and worker abstraction before auth, restaurant CRUD,
demo generation, reservations, outreach, analytics, AI receptionist, or content
automation can be implemented.

**Business Value:** This turns the repo from planning-only into a runnable
engineering foundation. It reduces delivery risk by giving future work a
standard way to start services, validate configuration, run migrations, expose
health checks, and add background jobs. That directly supports the sales MVP by
making the lead-to-demo-to-reservation loop buildable in small, testable steps.

**Plan Fit:** This completes the first Phase 1 build-order item: Go backend
foundation, config, logging, migrations, and health check. It unblocks P1-E02
auth/roles/tenant safety and P1-E03 restaurant/profile/menu CRUD.

**Checks Run:** `go test ./backend/...`, `make test`, API smoke check for
`/healthz` and `/readyz`, and worker smoke check for sample job processing.

**Risks / Follow-ups:** `make migrate-up` was not run against a real PostgreSQL
database in-session. Durable queue, auth/session provider, and DB repository
pattern choices remain to be finalized as later epics need them.

## 2026-06-17 — Session Delivery Documentation Contract

**Role:** Documentation Agent

**Delivered:** Updated `AGENTS.md` so future sessions read the detailed delivery
log during orientation, overwrite the concise summary doc with 3-5 lines, and
update both session docs before the final response. Added this detailed
delivery log and the concise summary doc.

**Why:** Raj requested a persistent handoff mechanism so each session records
exactly what was delivered, why it matters, its business value, and how it fits
with the larger plan.

**Business Value:** This creates continuity between coding sessions. Future
agents can quickly understand what already exists, why it was built, and what
work it unlocks without re-reading every implementation detail or losing the
sales-MVP context.

**Plan Fit:** This strengthens the operating process around Phase 1 and Phase 2
delivery. It does not change product behavior, but it improves execution
discipline and handoff quality for every subsequent epic.

**Checks Run:** Documentation-only change; no code tests were required.

**Risks / Follow-ups:** Future agents must keep `docs/SESSION_SUMMARY.md`
concise and overwritten, while keeping this file as the detailed running log.

## 2026-06-22 — P1-E01 Foundation Completion + Auth Prep

**Role:** Backend Agent

**Delivered:** Completed P1-001 through P1-005 and early P1-E02 prep. Fiber API
with graceful shutdown; strict typed config (`env.go`); Docker Compose for
PostgreSQL 16; migrations 000001 (foundation) and 000002 (users); repository
pattern for metadata and users with Postgres + Mock implementations; `store`
package with startup verification; required DB connect before API/worker start;
`APP_ROLE` gating on health endpoints; aligned `.env.example` and `backend/.env`;
work log at `docs/work-log/2026-06-22-backend-foundation-session.md`; README
command updates.

**Why:** The sales MVP needs a reliable, testable backend base before auth,
restaurant CRUD, demos, reservations, outreach, and AI workflows can be built.
Each P1-E01 ticket unblocks the next layer of product work.

**Business Value:** Developers can run `make dev` and get a working API against
real PostgreSQL with migrations, health checks, structured logging, and a clean
path to add domain APIs. Reduces integration risk for the lead-to-demo loop.

**Plan Fit:** Finishes Phase 1 build-order items 1–5 (foundation through HTTP
layer). Unblocks P1-007 login, P1-008 auth middleware, and P1-E03 restaurant
CRUD.

**Checks Run:** `make test`, `go test ./backend/...`, `make migrate-up`,
`make api` smoke with `/healthz` and `/readyz`.

**Risks / Follow-ups:** No Postgres integration tests for repositories yet
(mocks only). P1-006 durable jobs not wired to `job_runs` table. Restrict
signup roles in staging/production. Health endpoints behind JWT may need
separate probe path for load balancers later.

## 2026-06-22 — JWT Auth + Protected Health Endpoints

**Role:** Backend Agent

**Delivered:** JWT-based signup/login (`POST /api/v1/auth/signup`, `POST /api/v1/auth/login`),
bcrypt password hashing, `TokenManager` with `JWT_ACCESS_TOKEN_TTL`, auth service layer,
`RequireAuth` and `RequireRole` middleware, developer-only `/healthz` and `/readyz`.
Removed `APP_ROLE` env gating and `health_access.go`. Migration `000003` drops
`users_role_check` DB constraint. Updated README, `.env.example`, and `backend/.env`.

**Why:** PM requested roles not be hardcoded in schema; authorization should come from
JWT issued at login with role stored in the database.

**Checks Run:** `go test ./backend/...`

**Follow-ups:** Restrict public signup role selection in production; add restaurant-scoped
RBAC (P1-008/P1-009).

## 2026-06-23 — Melbourne Food Image Scraper (automation/)

**Role:** Backend / Automation Agent

**Delivered:** Python pipeline under `automation/` that scrapes food images for
50 curated Melbourne restaurant dishes (Australian, Thai, Indian, Italian, and
other cuisines). Saves images to `automation/images/` by SHA-256 hash and writes
metadata to `automation/data/catalog.json`.

**Why:** Pre-builds demo menu image assets for P1-013 menus and P1-048 seed data
so sales demos look realistic without manual photo hunting per lead.

**Checks Run:** `python -m py_compile automation/scrape.py`; live scrape test
with `--limit 3` (3 images saved via Bing fallback).

**Business Value:** Reusable food catalog with hash-based image storage for
Melbourne-focused demo restaurants.

**Plan Fit:** Upstream content work for P1-E03 menu items (`image_url`) and
P1-048 seed/demo data.

**Risks / Follow-ups:** Google Images requires a browser and may CAPTCHA; Bing
fallback used when Google times out. Review image licensing before production.
Run full scrape with `python scrape.py --skip-google` for speed or default mode
for Google-first.

## 2026-06-22 — P1-009 Restaurant-Scoped Access Checks

**Role:** Backend Agent

**Delivered:** Hardened restaurant access layer with `RestaurantID` in request
context, extended `access.Service` (list/create/members), reusable
`protectRestaurantScoped` / `protectRestaurantAdmin` middleware, restaurant list
and member APIs, migration `000005_demo_sites`, public demo route
`GET /api/public/v1/demo/{slug}?token=...` with allowlisted `PublicDemoPayload`
DTO, admin `POST /api/v1/restaurants/{id}/demo-sites`, `make seed-demo-fixture`,
Postman collection updates, and table-driven isolation tests.

**Why:** P1-009 requires tenant isolation before restaurant CRUD, profiles, and
demo generator work can safely scale.

**Checks Run:** `go test ./backend/...`

**Business Value:** Restaurant owners cannot read another tenant's data; public
demo links expose only sales-safe payload fields.

**Plan Fit:** Completes P1-009 acceptance criteria; unblocks P1-010 full
restaurant CRUD and P1-015+ demo payload builder.

**Risks / Follow-ups:** Demo token in query string may appear in proxy logs;
P1-016 can move to header/path. Full restaurant lead fields deferred to P1-010.

## 2026-06-22 — Restaurant Lead Fields (sales scope)

**Role:** Backend Agent

**Delivered:** Migration `000006_restaurant_lead_fields` adding `email`,
`is_contacted`, and `shown_interest` to `restaurants`. API create/list/get now
return lead fields plus `created_at`/`updated_at`. Create requires `name` and
`email`. Repository `MarkShownInterest` added for future email click tracking.
No image or profile fields (deferred per scope).

**Why:** Aligns `restaurants` table with sales MVP lead tracking described in
the implementation guide and product scope: outreach email → link click →
`shown_interest = true`.

**Checks Run:** `go test ./backend/...`

**Follow-ups:** Wire `MarkShownInterest` from campaign/email click tracking;
admin PATCH to set `is_contacted` when outreach is sent.

## 2026-06-22 — P1-010 Restaurant CRUD + Query Filters

**Role:** Backend Agent

**Delivered:** Completed P1-010 with migration `000007_restaurant_status`, full
restaurant lifecycle `status`, `PATCH /api/v1/restaurants/{id}`,
`PATCH /api/v1/restaurants/{id}/status`, soft archive via
`DELETE /api/v1/restaurants/{id}`, and list filters via query params
(`?restaurant=`, `?status=`, `?is_contacted=`, `?shown_interest=`,
`?include_archived=`). Auto-status rules: `is_contacted` moves lead to
`emailed`; `shown_interest` moves to `interested`.

**Checks Run:** `go test ./backend/...`

**Plan Fit:** P1-010 acceptance criteria met; unblocks P1-011 profiles and demo
payload builder.

## 2026-06-22 — Repository Structure Alignment (Phase 1 Guide §5)

**Role:** Backend Agent

**Delivered:** Restructured `backend/internal` from layered `repositories/` +
`services/` into domain packages aligned with
`docs/phase1/PHASE1_IMPLEMENTATION_GUIDE.md`: `restaurants/` (repo + access
service), `demos/` (repo + demo service + token helpers), `auth/` (JWT, user
repo, signup/login service). Moved shared errors to `platform/errors/` and
metadata repo to `platform/metadata/`. Added placeholder `doc.go` packages for
future domains (`menus`, `reservations`, `campaigns`, `analytics`, `ai/*`,
`providers/*`, `platform/telemetry`, `backend/tests`). Moved
`PHASE1_IMPLEMENTATION_GUIDE.md` and `PHASE1_TECHNICAL_BACKLOG.md` to
`docs/phase1/`.

**Why:** The implementation guide defines domain-owned packages so Phase 1
modules can grow without cross-layer import sprawl and match the documented
monolith boundaries.

**Checks Run:** `go test ./backend/...` — pass

**Risks / Follow-ups:** No API behavior change intended; imports only. Add ADR
if we later split `auth` user persistence from JWT helpers.

## 2026-06-23 — Postman Full Flow Update

**Role:** Documentation Agent

**Delivered:** Rebuilt `postman/Restaurant-Platform.postman_collection.json`
with ordered **01 — PM Demo Flow** folder (15 steps), fixed role names
(`internal_admin`, `restaurant_owner`), added Auth Me, List Members, filter
examples, owner/admin token switching, and updated environment + README.

**Checks Run:** JSON validation on collection and environment files.

**Plan Fit:** Supports PM demo of P1-008 through P1-010 without manual curl.

## 2026-06-23 — OpenAPI / Swagger Documentation

**Role:** Documentation Agent

**Delivered:** OpenAPI 3.0 spec at `docs/openapi/openapi.yaml` covering all 17
endpoints (auth, admin, restaurants, members, demo sites, public demo, health).
Added `docs/openapi/README.md`, `make openapi` validation target, and README/Postman links.

**Checks Run:** `make openapi` — pass

## 2026-07-09 — VM Catalog Redeploy From Latest Local Source

**Role:** DevOps Agent

**Delivered:** Deployed the clean local source from
`/Users/rajchodisetti/MonoRepo` commit `c4a2146` to `/opt/tuvi/MonoRepo` on VM
`170.64.154.143`. Rebuilt and restarted the `restaurant-services-catalog`
container from that source. Updated host Caddy so `tuvisolutions.com` and
`www.tuvisolutions.com` route to the catalog service on `127.0.0.1:15173`
instead of the old Tuvi Next website on `127.0.0.1:15174`.

**Why:** Local `http://127.0.0.1:5174/` was serving the latest restaurant
services catalog, while the VM/public `/services/restaurants` route was still
coming from the older Next website. Public root and `/services/restaurants`
needed to use the same latest catalog code.

**Checks Run:** `npm --prefix apps/restaurant-services-catalog run build` —
pass. `npm --prefix tuvi-website/app run build` — failed locally with webpack
`Unexpected end of JSON input`, so the Next website was not used for this
deployment. VM `docker compose ... config` — pass. VM
`docker compose ... up -d --build --no-deps restaurant-services-catalog` —
pass. `caddy validate --config /etc/caddy/Caddyfile` — pass.

**Verification:** `https://tuvisolutions.com` and
`https://tuvisolutions.com/services/restaurants` both return
`<title>Tuvi Solutions Restaurant Services`. The FAL catalog video endpoint
returns `content-type: video/mp4`. `https://api.tuvisolutions.com/` and
`https://api.tuvisolutions.com/docs/` still respond.

**Operational Notes:** Previous VM source tree was preserved as
`/opt/tuvi/MonoRepo.prev-20260708191646`. The old `tuvi-tuvi-website-1`
container remains as an orphan but no public Caddy route points to it.

## 2026-07-14 — Places-First Ingestion, Apollo Contact Enrichment, and Outreach Safety Gates

**Role:** Backend, Security, and Documentation Agent

**Delivered:** Reordered scheduled ingestion to bounded, paginated Google Places
API (New) discovery and Place Details first, followed by targeted Apollo contact
enrichment only when the resulting lead lacks an owner or contact email. Apollo
People Search is constrained to an owned business domain and approved
decision-maker titles; People Match retrieves the selected person's full owner
details and work email with personal-email and phone reveal disabled. Ambiguous
free/social domains are skipped. The shared request budget accounts for both
providers, and both credentials are required while their default stages are
enabled.

Scheduled ingestion deduplicates by Place ID against PostgreSQL and local
canonical JSON, requires the database when import is enabled, and loads an
explicit host env file with correct precedence. Provider credentials are sent in
headers. Credential-bearing URLs are scrubbed from merged/imported payloads and
Google photo resource metadata is stored instead. Website email extraction now
rejects private-network targets, validates manual redirects, limits response
size, and stays on the business host.

Fixed the Apollo-era dedup bug that skipped newly registered leads before their
first Places scrape. Added the Places-first/Apollo-enrichment decision ADR and updated environment
templates, service inventory, runbook, OpenAPI, Phase 1 backlog, Makefile, and
operator README.

Hardened email automation so bulk jobs only consume individually approved
campaigns attached to approved, OCR-verified profiles and published demos. Bulk
sending now obeys the global disable switch, enforces a 150-send/50-per-account
ceiling, adds a configurable delay, and never creates or auto-approves drafts.
Disabled or redirected sends produce a `skipped` event, return the campaign to
`approved`, and do not mark the restaurant contacted. Provider errors and logs
redact recipient addresses, and provider sends use one job attempt to avoid
duplicate real-world delivery after a partial state-update failure.

**VM Status:** Inspection found no completed Melbourne ingestion and no usable
Places/Apollo/database ingestion credentials on the VM. The Python virtual
environment was prepared, but no Places or Apollo request, database import,
deployment, or outreach email was triggered. Production still has email sending
disabled and no configured bulk account pool.

**Checks Run:** `go test ./backend/...` — pass; outreach Python `py_compile` —
pass; `python -m unittest discover -p '*_test.py'` — 26 pass;
`python -m pip check` — pass; `make openapi` — valid with two pre-existing lint
warnings; `bash -n automation/outreach/cron_lead_ingestion.sh` — pass;
`git diff --check` — pass. `shellcheck` was unavailable locally.

**Business Value / Plan Fit:** Uses stable Place-ID restaurant discovery while
spending Apollo calls only on missing owner/work-email data, closes credential
and PII leaks, and restores Phase 1's required human approval boundary before
external outreach.

**Risks / Follow-ups:** A deployment still requires explicit approval. The VM
needs a protected, host-reachable `DATABASE_URL`, a workload-restricted Places
API (New) key, and an Apollo API key before a Melbourne run can start. Apollo
People Match can consume credits and its match accuracy/yield needs measurement.
Places does not provide menu items. Google photo rendering requires a future
server-side photo proxy if those images are exposed publicly.

## 2026-07-15 — Voice Booking Form and Services Menu Fix

**Role:** Frontend and AI Workflow Agent

**Delivered:** Replaced the browser voice assistant's model-dependent email-only
prompt with a required name, email, and phone booking form. The voice server now
opens the form after slot confirmation, receives its values over the existing
WebSocket, and falls back to opening it if the model tries to book without first
requesting the form. The agent thanks the visitor and says the booking is
confirmed only after the consultation API returns success. It no longer claims
that an email was sent when delivery may be disabled. The desktop Services menu
is now controlled only by component state, so selecting its first link closes it
instead of the hover/focus CSS reopening it.

**Production Inspection:** The VM contains zero rows in both
`company_consultations` and `reservations`, and the API logs contain no matching
confirmation-email delivery event. Production has `EMAIL_PROVIDER=disabled` and
`EMAIL_DISABLE_SENDING=true`, so no booking confirmation email was sent.

**Checks Run:** `npx tsc --noEmit --incremental false` — pass; Python
`py_compile` — pass; browser/phone prompt and tool contract assertions — pass;
`git diff --check` — pass. A local Next build remains blocked by the pre-existing
dependency/cache `Unexpected end of JSON input` failure. A clean Docker build
could not start because Docker Desktop's storage is full; no prune was run.

**Plan Fit / Follow-up:** Restores the Phase 1 consultation conversion path while
keeping confirmation tied to a successful database booking. No production
deployment or email-provider change was performed; both require explicit
approval.

## 2026-07-15 — PostgreSQL-Only Consultation Slot Locking

**Role:** Backend and DevOps Agent

**Delivered:** Made PostgreSQL the sole source of truth for Tuvi consultation
availability and confirmed bookings. The API no longer initializes or calls the
Google Calendar provider, checks only confirmed `company_consultations` rows,
and stores successful bookings before reporting confirmation. Migration 25
replaces the unconditional `slot_start` uniqueness rule with a partial unique
index for `status = 'confirmed'`, preventing concurrent double booking while
allowing cancelled slots to be reused. Added a focused availability regression
test and an accepted ADR for the deferred calendar integration.

**Checks Run:** `go test ./backend/internal/consultations ./backend/internal/http/...`
— 45 pass; `go test ./backend/...` — 146 pass. Cross-component and production
deployment verification are recorded with the release outcome below.

**Business Value / Plan Fit:** A confirmed booking immediately disappears from
future availability without relying on external calendar state. This provides
the deterministic Phase 1 booking boundary needed before Google Calendar
synchronization is designed and approved.

**Production Deployment:** Pushed functional commit `000cdd8` to
`agent/tuvi-oauth-homepage-verification`, created protected pre-migration backup
`/opt/tuvi/backups/monorepo-pre-000cdd8-20260715T192444Z.dump`, and staged the
exact commit at `/opt/tuvi/releases/monorepo-000cdd8`. Migration 25 applied
successfully. The API, voice agent, and Tuvi website were rebuilt and recreated;
all run with zero restarts. PostgreSQL reports the partial index predicate
`status = 'confirmed'`, the former unconditional constraint is absent, and the
consultation table remains empty. No fake production booking was created.

**Production Verification:** Protected and public availability both return
`status=success` with 16 current slots; the voice status returns `ready`; the
public website and voice host return HTTP 200; the rendered website contains the
voice widget; and new API logs contain no calendar-provider initialization.

## 2026-07-18 — Gmail Outreach Health and Restaurant Admin Completion

**Role:** Full-Stack, Backend, Frontend, Security, Test, and Documentation Agent

**GitHub / Local State:** Fetched and pruned all remotes. GitHub reports no open
pull requests; PRs 1–8 are merged. The current local HEAD `44b6d7c` is identical
to `origin/master`. Work stayed on `agent/tuvi-oauth-homepage-verification` to
preserve unrelated dirty-worktree files; no user-owned change was overwritten.

**Delivered:** Made the restaurant outreach account pool Gmail-only and kept it
configuration-driven through `OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON`, so a new
mailbox needs no code change. Added migration `000029`, a PostgreSQL sender-health
ledger, and a worker loop that sends one real check from every configured Gmail
mailbox to `rajchodisetti@gmail.com` when first registered and at least every 24
hours. The Outreach UI now shows each account's pending/checking/healthy/failed
state, last and next check, provider acceptance, and safe error. Existing
migration `000024` remains the strategic delivery schedule: 40 attempts per
account over eight hours, persisted 2–5 minute jitter/global gap, and 24-hour
cooldown/continuation.

The restaurant outreach job is now operationally controlled by a persisted
on/off control in the Outreach admin UI instead of `EMAIL_DISABLE_SENDING`.
Enabling queues the durable workflow; disabling prevents the next Gmail request
from starting while allowing an in-flight request to resolve safely. The job
continues to select only OCR-verified rows with an email plus the existing human
profile, demo, and campaign approval evidence. Gmail OAuth credentials remain in
secret environment configuration, and health checks remain separately controlled.

Fixed all Cinematic, Aurora, and Elysian admin preview URLs to carry the immutable
restaurant UUID and added a public UUID content resolver, preventing index changes
or fallback sample data from showing the wrong site. Outgoing email service cards
now include both the Tuvi restaurant-services presentation and three distinct
restaurant-specific signed template links. The restaurant demo UI no longer
exposes Publish/Unpublish or a public-demo link; its View personalized website
buttons create a protected session capability and begin tracking before opening
the chosen template.

Made Contacted and Shown interest read-only automation results in the restaurant
view. Confirmed sends already set Contacted/emailed; successful tracked link events
now set Shown interest/interested. Added signed-demo session capabilities,
heartbeats/end foreground-only duration, selected-template evidence, forwarded AI
receptionist transcript turns, a protected restaurant engagement API, and an
Engagement tab. Added phone, address, and OCR status to restaurant responses, an
OCR table column and verified-only filter, and per-row Apollo outcomes in profile
review.

Secured the voice service's existing call/transcript read routes and administrative
controls with the already-configured `CALL_API_SECRET` dependency and constant-time
credential comparison. Production currently contains 10 generic call rows and 308
transcript turns but zero restaurant-indexed calls, so there is no historical
restaurant-specific transcript set to migrate into the new tenant-checked ledger.

**Live Read-Only Findings:** The production environment and inspected backup have
zero Gmail and zero Zoho outreach account records, so no credential could be tested
and no real health email was sent. Production data has 708 restaurants: 589 lack
email and 15 lack phone. Apollo successfully enriched 119 rows and all 119 have an
email; 511 missing-email rows are `no_candidate` and 78 are `skipped_no_domain`.

**Checks Run:** `go test ./backend/...` — all backend packages pass, including
new admin-preview/template tests; `go vet ./backend/...` — pass; admin
`npx tsc --noEmit --incremental false` — pass; admin `npm run lint` — pass;
template `npx tsc --noEmit
--incremental false` — pass; Redocly OpenAPI lint — valid with three pre-existing
warnings plus the existing localhost-server warning when run without the Makefile
skip; voice `python3 -m py_compile voice-sales-agent/bot.py` — pass; `git diff
--check` — pass. Current Gmail and FastAPI documentation was checked with Context7
to verify OAuth refresh-token, `users.messages.send`, and protected dependency behavior.
Clean Node 22 production builds pass for both the admin and template apps. The
template app has no ESLint 9 configuration, so its standalone lint command remains
an existing tooling gap; its TypeScript and production-build checks are clean.

**Business Value / Plan Fit:** Completes the Phase 1 visibility loop from configured
sender → UI-authorized paced outreach → confirmed contact → tracked interest →
personalized template/foreground time/transcript, while retaining human profile,
demo, and campaign approval gates.

**Risks / Follow-ups:** Nothing was deployed, migrated, committed, pushed, or sent.
Deployment requires explicit approval, migrations `000024` and `000029`, and valid
per-mailbox Gmail OAuth records. The three pasted Google API keys were treated as
exposed, were not stored or tested, and cannot authorize mailbox sending; they must
be revoked/rotated. Each Gmail mailbox needs offline OAuth consent and a refresh
token with `gmail.send` permission. Daily health status proves Gmail API acceptance,
not inbox delivery.

## 2026-07-19 — Gmail Health, Admin Engagement, and Budgeted Background OCR Deployment

**Role:** Full-Stack, Backend, Frontend, AI Workflow, Security, Test, DevOps,
and Documentation Agent

**Delivered:** Pushed `e6e7950` and `caffcfb` directly to `master`, with
`caffcfb` as the deployed release. The Gmail-only outreach account pool, sender
health ledger/UI, persisted email-job switch, paced delivery workflow, three
restaurant-specific template links, Tuvi service presentation, UUID-backed
previews, engagement/time/transcript evidence, contact/Apollo/OCR visibility,
and secured voice reads from the prior local session are now live. The ignored
local Gmail JSON was normalized to one dotenv-safe line, validated as three
complete unique accounts, transferred through a mode-0600 temporary payload,
and written only to the protected VM stack environment. No credential value was
printed or committed.

Added migration `000030` and a long-running `ocr-worker`. It claims only
restaurants whose canonical email is nonblank and atomically reserves every
vision-provider request in `ocr_daily_request_usage`. The global limit is hard
capped at 200 requests per UTC day across worker restarts. OpenAI-compatible and
Google GenAI automatic retries are disabled, provider/image calls receive an
explicit timeout, and timeout/429/temporary failures release the current and
unstarted claims to `pending` without consuming an OCR attempt. Ambiguous timed-
out provider requests still consume one budget reservation. The old one-shot
`ocr-job` remains available for diagnosis, and no competing VM OCR cron exists.

Fixed the scrape-job trigger redirect to include the deployed `/admin` base path.
Direct local and public requests to `/admin/scrape-jobs/<id>` now reach the admin
application and return the expected authentication redirect instead of a 404.

**Production Deployment:** The live database was at migration `000014` when the
rollout began; pending migrations applied in order through `000029`, then
`000030` applied after the OCR changes. Restaurant outreach remained seeded off
throughout. A validated backup of the real `monorepo` database was written to
`/opt/tuvi/backups/monorepo-pre-caffcfb-20260719T092320Z.dump`; the OCR environment
backup is `/opt/tuvi/backups/ingestion.env.pre-caffcfb-20260719T092320Z`, and the
pre-Gmail stack configuration backup is
`/opt/tuvi/backups/stack.env.pre-e6e7950-20260719T090404Z`. Prior live images were
tagged `rollback-fd2cc94`. The exact release is
`/opt/tuvi/releases/monorepo-caffcfb`, now targeted by `/opt/tuvi/MonoRepo`.

API, Go worker, admin web, template, voice agent, and OCR worker images built and
started successfully. API auth gating returned 401 without a JWT as expected;
admin login, template, and voice returned HTTP 200. Containers reported zero
restarts at verification.

**Live Verification:** All three configured Gmail accounts registered, ran their
first real-message health check, and stored `healthy` with a provider message ID;
zero failed. This proves Gmail API acceptance, not inbox placement. The persisted
restaurant email job is `false`, active `outreach.bulk_send` jobs are zero, and
restaurant delivery attempts created in the deployment window are zero.

The OCR background worker started with 119 email-equipped candidates. At the
verification snapshot it had reserved 17/200 daily requests, moved one additional
profile to `verified`, held 49 email-equipped running claims, and had zero
email-less running claims. The worker and budget continue in the background.

**Checks Run:** `go test ./backend/...` — 159 pass in 43 packages; `go vet
./backend/...` — pass; admin `npm run lint` and `npx tsc --noEmit --incremental
false` — pass; focused OCR/unit suite — 20 pass; Python `py_compile` for all changed
OCR modules — pass; all three configured provider clients accept the bounded
timeout/no-retry construction; Compose VM config validation — pass; `git diff
--check` and staged secret scans — pass. The broader outreach discovery ran 40
tests with 38 pass and two pre-existing legacy-ingestion errors because the
workstation PostgreSQL tunnel at `127.0.0.1:15432` was not running. Production
Node 22/Docker builds passed for admin, template, voice, API, worker, migrate, and
OCR worker. Context7 current Google GenAI and OpenAI Python documentation was used
to verify timeout units and retry controls.

**Business Value / Plan Fit:** Sender readiness, controlled outreach, tracked
restaurant engagement, valid admin navigation, and continuously progressing OCR
are now operational while preserving the Phase 1 human approval gate. Scarce OCR
capacity is spent only on leads that can actually receive outreach and cannot
exceed the provider's daily allowance.

**Risks / Follow-ups:** Gmail health proves provider acceptance only; inspect the
configured internal recipient inbox for placement. OCR will stop reserving calls
at 200/200 and resume after the UTC date changes. Keep the Outreach UI email job
off until profiles, demos, campaigns, recipients, and opt-out content have been
reviewed for a real send window.

## 2026-07-19 — Hybrid Restaurant Media and OCR-Aware Website Galleries (Local)

**Role:** Full-Stack, Backend, Frontend, AI Workflow, Security, Test, and
Documentation Agent

**Delivered:** Implemented a hybrid media pipeline that keeps Google Places
photos live-only and uncached while providing durable S3-compatible storage for
restaurant-owned and separately licensed images. New migration `000031` adds
source, rights, approval, placement, accessibility, visual-quality, OCR, and
visibility metadata without storing Google photo URLs, resource names, or bytes.
The import path now replaces only OCR-owned image rows and never deletes existing
durable/manual rows when a Google-only OCR result has no persistent URL.

Public restaurant and signed-demo payloads now resolve fresh Google Places photos
at request time, pair them with their fresh attribution/report links, and expose
only photos whose one-way resource fingerprint exactly matches a website-safe
OCR classification. Unmatched, unclassified, low-confidence, text-heavy unknown,
and menu-document photos fail closed.
Public responses are non-cacheable, and live
Google media bypasses the Next.js image optimizer. Restaurant-owned/licensed
assets use immutable CDN URLs, remain draft until the shared OCR worker verifies
them, and are rejected if identified as a menu document. The upload and visibility
controls are available in the authenticated restaurant photo view with OCR state,
errors, source rights, captions, and placement metadata.

Expanded OCR classification to food, drink, interior, exterior, logo, team,
event, menu document, and other, plus factual captions, accessible alt text,
tags, orientation, subject placement, people/text flags, and quality/hero scores.
All three personalized templates now use this metadata for source-aware hero
selection, ambience/food gallery filters, lightboxes, and safe reusable placements.
Google photos carry visible attribution at every rendered placement. Only durable
owned/licensed media is reused across story, experience, footer, and SEO surfaces.
Menu documents are excluded from public handlers, demo snapshots, legacy gallery
fallbacks, menu-item image fallbacks, templates, uploads, and SEO.

**Checks Run:** `go test ./backend/...` — 165 pass in 44 packages; `go vet
./backend/...` and `go build ./backend/cmd/...` — pass; focused OCR/import suite —
13 pass; changed Python modules compile; admin `npm run lint` and both TypeScript
checks — pass; Node 22 optimized production builds for admin and template — pass;
Redocly OpenAPI lint — valid with four existing warnings; `git diff --check` —
pass; VM Compose configuration validation — pass. Context7 current AWS SDK v2 and Next.js documentation was used for the
S3-compatible client, endpoint configuration, multipart proxy limit, and public
image behavior. Docker Desktop was unavailable, so no container build or database
migration was run locally.

**Business Value / Plan Fit:** Existing personalized sites can display current,
properly attributed non-menu Google imagery without depending on expired database
URLs. Restaurants can progressively replace provider media with controlled,
stable, OCR-enriched assets that improve the templates and remain portable.

**Risks / Follow-ups:** This work is not committed, pushed, migrated, or deployed.
Production remains on release `caffcfb` with migrations through `000030`.
Deployment requires explicit approval plus an S3-compatible bucket/CDN configured
with `STORAGE_*` variables and the template's matching public media base URL. The
live Google path intentionally fails closed until each resource fingerprint has
a successful OCR classification; migration `000031` returns older positional-only
Google classifications to pending, and the existing background worker prioritizes
restaurants with demos while retaining the shared 200-request UTC-day ceiling.

## 2026-07-19 — Mobile Template Switch Clarity (Local)

**Role:** Frontend, Test, and Documentation Agent

**Delivered:** Reworked the shared template switcher used by Cinematic, Aurora,
and Elysian on mobile. The mobile control is now a full-width website-design card
that identifies the current design, shows its position in the three-template set,
names the next preview, and explicitly says that restaurant details and photos do
not change. Desktop copy now says “Preview” rather than implying a permanent
switch. Elysian's desktop and mobile controls have separate responsive visibility,
removing the duplicate control at smaller widths. Navigation now uses the current
pathname and a cloned `useSearchParams` value, so restaurant IDs, signed-demo
tokens, and other query parameters survive a template preview.

**Photo URL Clarification:** Personalized demos never try to reopen the removed
OCR URL. The API resolves current Google Places media URLs for each request and
returns a photo only when its one-way provider-resource fingerprint exactly
matches a website-safe OCR classification. If Google changes the underlying
resource identity, the old URL is not reused and the photo fails closed until an
OCR refresh. Durable owner/licensed media continues to use its stable configured
object-storage URL.

**Checks Run:** Template TypeScript check — pass; Node 22 optimized production
build — pass; local development server returned HTTP 200 for Cinematic, Aurora,
and Elysian sample routes; `git diff --check` — pass. Context7 current Next.js
App Router guidance was used for query-parameter preservation. The in-app browser
runtime had no available browser, so interactive screenshot verification could
not be run in this workspace.

**Risks / Follow-ups:** The change is local and unreleased. Production remains on
`caffcfb`/migration `000030`; the hybrid media and mobile-switch work still require
the previously documented storage configuration and explicit deployment approval.

## 2026-07-19 — Hybrid Media and Mobile Template Production Deployment

**Role:** Full-Stack, Backend, Frontend, AI Workflow, Security, Test, DevOps,
and Documentation Agent

**Delivered:** Pushed `b5f7299` and corrective packaging commit `6c21c15`
directly to `master`, then deployed the immutable `6c21c15` release at
`/opt/tuvi/releases/monorepo-6c21c15`. Migration `000031` is applied. Fresh,
attributed Google Places photo resolution, exact OCR resource-fingerprint
matching, public menu-document exclusion, non-destructive media imports,
OCR-aware template placements, authenticated admin media controls, and the
clear mobile template preview control are live.

The release was built before downtime. A compressed database backup was saved
as `/opt/tuvi/backups/monorepo-pre-b5f7299-20260719T174044Z.dump`, alongside
protected stack and ingestion environment backups. The first production smoke
found that `Dockerfile.ocr` omitted the new shared `media_asset_metadata.py`
module. The OCR container was isolated in a restart loop; the API, admin, and
templates remained healthy. Commit `6c21c15` added the missing module to the
image, focused OCR/import tests passed 13/13, and only the OCR image/container
was rebuilt. The corrected OCR worker is running with zero restarts.

**Live Verification:** API, Go worker, scrape worker, OCR worker, admin portal,
template, corporate website, and voice containers are running with zero
restarts and no current error/traceback/panic/fatal logs. Public corporate site,
restaurant API, admin login, voice readiness, and template variants 1/2/3 all
return HTTP 200. Restaurant detail responses return `Cache-Control: no-store,
max-age=0`. Schema version is 31 and `restaurant_media_assets` exists.

Migration `000031` safely returned the 21 old fingerprint-less `verified`
profiles to `pending`; production now has 940 pending and 4 failed profiles.
There are 154 email-equipped unverified candidates. Today's shared OCR allowance
was already 200/200, so the background worker is intentionally waiting for the
UTC reset and has not exceeded the provider limit. The persisted outreach email
job remains off and all three Gmail sender-health records remain healthy.

**Checks Run:** Pre-release checks remained green: 165 Go tests across 44
packages, Go vet/build, 13 OCR/import tests, Python compilation, admin lint and
TypeScript, template TypeScript, both Node 22 production builds, OpenAPI,
Compose validation, and diff checks. Production Docker builds succeeded for
migrate, API, worker, scrape worker, OCR worker, admin, and template. Post-release
database, container, logs, local HTTP, public HTTPS, cache-header, Gmail-health,
outreach-control, OCR-budget, and release-symlink checks passed.

**Risks / Follow-ups:** No S3-compatible production bucket/CDN is configured,
so owner/licensed uploads remain unavailable while `STORAGE_PROVIDER=disabled`.
This does not block live Google photo resolution. Personalized Google photos
will remain fail-closed until the post-reset OCR worker writes exact resource
fingerprints; menu documents remain excluded. The 200-request daily ceiling means
the 154 eligible profiles will complete progressively rather than all at once.

## 2026-07-19 — Restaurant List Filter Lifecycle Fix

**Role:** Backend, Test, and Documentation Agent

**Delivered:** Fixed the admin restaurant list's `demo_ready` lifecycle gap.
Automatic lead preparation now advances a lead to `demo_ready` when it creates,
refreshes, or reuses generated demo/outreach draft artifacts. Manual admin demo
draft creation now also advances only fresh `lead` records to `demo_ready`,
without downgrading emailed, interested, client, lost, or archived restaurants.
Migration `000032` backfills existing `lead` restaurants that already have draft
or published demo records, so the Status → `demo_ready` filter can return
already-generated demos after deployment.

Added focused regressions proving the admin list accepts and returns
`status=demo_ready` and `ocr_status=verified` results. The OCR filter code path
was already value-compatible; the new test protects it from regressing while the
demo-ready lifecycle fix removes the common combined-filter empty-list case.

**Checks Run:** Focused restaurant HTTP filter tests passed; focused demo-service
status-transition test passed; `go test ./backend/...` passed with 167 tests in
44 packages; `go vet ./backend/...` passed; `git diff --check` passed.

**Business Value / Plan Fit:** Admins can reliably find leads that have moved
from scraping/OCR into the personalized-demo review stage, which supports the
Phase 1 lead → demo → outreach workflow.

**Risks / Follow-ups:** The ignored local database tunnel at `127.0.0.1:15432`
was unavailable, so migration `000032` was not applied locally. Production or
staging deployment requires the normal migration approval gate.

## 2026-07-19 — Demo-Ready Filter Deployment and UIPro Codex Skill

**Role:** Backend, DevOps, Test, and Documentation Agent

**Delivered:** Installed UI/UX Pro Max for Codex in the monorepo with
`uipro init --ai codex`; the generated skill bundle lives under
`.codex/skills/ui-ux-pro-max`, with Python bytecode ignored and excluded.
Committed and pushed `cbc2eb8` to `master`, then deployed it to the VM as
`/opt/tuvi/releases/monorepo-cbc2eb8`. The `/opt/tuvi/MonoRepo` symlink now
points to `cbc2eb8`, and `/opt/tuvi/previous-release-path` points to the prior
`6c21c15` release.

**Production Migration:** A gzip-validated backup was saved at
`/opt/tuvi/backups/monorepo-pre-cbc2eb8-20260719T181834Z.dump.gz` before
migration. Migration `000032` applied successfully to the active `monorepo`
database. Before the migration, schema was 31 and 22 `lead` restaurants already
had draft/published demo records. After the migration, schema is 32,
`demo_ready=22`, `lead=922`, and zero `lead` rows still have a demo record.

**Live Verification:** Rebuilt backend/migrate images from `cbc2eb8` and
force-recreated only API and worker. All Tuvi containers are running with zero
restarts. API, admin login, corporate website, demo template, and voice readiness
all return HTTP 200. API, worker, OCR worker, and scrape-worker logs show no
panic/fatal/traceback/error lines in the checked tails. The UIPro search script
runs on the VM through `python3`.

**Safety State:** The Outreach UI email job remains disabled, active
`outreach.bulk_send` jobs are zero, and today's OCR budget remains 200/200.
Production currently has `restaurant_profiles` counts `pending=940`,
`failed=4`, and `verified=0`; the OCR verified-only list filter will remain
empty until new rows reach `ocr_status=verified`.

**Checks Run:** Local `go test ./backend/...` passed with 167 tests in 44
packages, `go vet ./backend/...` passed, staged whitespace checks passed, and
UIPro local/VM smoke checks passed.

## 2026-07-19 — Italian Villa Experimental Restaurant Template

**Role:** Frontend, Backend, Test, DevOps, and Documentation Agent

**Delivered:** Added a payload-driven `template=4` personalized website for
Italian-focused restaurant demos. The new Italian Villa template uses the
existing restaurant payload contract, safe source-aware media loader, Google
photo attribution, reservation availability and request flow, generated menu
sections, reviews, contact details, hours, template switching, and the existing
floating voice-agent widget. UI/UX Pro Max was used for the premium restaurant
design research pass, with emphasis on high-quality photography, restrained
luxury typography, mobile-first booking, trust cues, and a fast path from visual
appeal to reservation intent.

The admin restaurant Demo tab now separates stable generated-site templates
from an **Experimental templates** section and exposes “Italian Villa
experimental” for each restaurant-generated payload. Demo engagement analytics
now accepts `template_id=4`, and migration `000033` updates the production check
constraint accordingly.

**Production Deployment:** Commits `3b0c246` and `5ffddf2` were pushed to
`master`; the VM is running the app code from
`/opt/tuvi/releases/monorepo-5ffddf2` with `/opt/tuvi/previous-release-path`
pointing to `/opt/tuvi/releases/monorepo-3b0c246`. A pre-migration backup was
created at
`/opt/tuvi/backups/pre-italian-template-3b0c246-20260719-185708.sql.gz`
before applying migration `000033`; the active `monorepo` database now reports
schema version 33.

**Checks Run:** UIPro design-system search passed locally. Local checks passed:
template TypeScript, admin TypeScript, admin ESLint, `go test ./backend/...`
with 167 tests in 44 packages, and `git diff --check`. VM Docker builds passed
for migrate, API, template, and admin-web at `3b0c246`; the final template image
build passed at `5ffddf2`. Production smoke checks returned HTTP 200 for the
Italian demo URL, admin login, public restaurants API, and voice readiness. All
Tuvi containers were running with zero restarts after deployment.

**Business Value / Plan Fit:** Sales demos now have an Italian-cuisine-specific
premium website option that can be opened directly from restaurant leads without
hand-building content. This expands the Phase 1 personalized demo library while
preserving existing media safety, reservation, voice, and engagement contracts.

**Risks / Follow-ups:** The template intentionally uses the existing fail-closed
media pipeline. Restaurants without verified safe media still render the
premium layout and copy, but image-heavy sections appear only when the payload
contains safe media. Owner/licensed uploads remain dependent on the existing
future bucket/CDN configuration.

## 2026-07-19 — Italian Template Removal and Generated Preview Photo Repair

**Role:** Frontend, Backend, Test, DevOps, and Documentation Agent

**Delivered:** Removed the rejected Italian Villa experimental template from the
personalized demo app, admin restaurant Demo tab, engagement tracking, template
switcher, voice-widget variants, package scripts, and public template registry.
The generated-site template set is back to Cinematic, Aurora, and Elysian only.
Migration `000034` maps any existing `template_id='4'` demo-session analytics
rows to Elysian (`3`) and restores the database check constraint to
`template_id IN ('1', '2', '3')`.

Fixed generated lead preview photos. The immediate production issue was that
reviewed public media was empty: no durable media rows existed, and the older
Google OCR classifications lacked exact source fingerprints/public eligibility,
so the strict public path correctly returned no photos. Migration `000035`
returned those fingerprint-less Google classifications to pending, and the
rebuilt OCR worker will refresh them after the UTC daily budget resets.

To unblock admin-opened generated previews now, the by-restaurant public site
endpoint supports `preview_media=google_live`. That flag is used by the template
app's UUID preview adapter only, and only when reviewed media is empty. It
resolves fresh Google Places photos live, returns attribution/report metadata,
sets `unoptimized=true`, and keeps provider URLs uncached. All three production
templates now preserve the full media object for story, timeline, experience,
about, and footer placements so live Google images and attribution render
instead of being filtered out.

**Production Deployment:** Pushed and deployed removal commit `b2a2f83`, then
deployed photo-repair commit `88b58eb` as the active VM runtime release at
`/opt/tuvi/releases/monorepo-88b58eb`. The rollback pointer is
`/opt/tuvi/releases/monorepo-b2a2f83`. A gzip-validated pre-migration backup was
saved at `/opt/tuvi/backups/pre-remove-italian-b2a2f83-20260719-193900.sql.gz`.
The active database schema is `000035`.

**Live Verification:** Production `demo_sessions` now only contain template IDs
`1` and `2`, and the live constraint permits only `1`, `2`, and `3`. A generated
preview for restaurant `020ca56c-4f99-4d1a-b449-2b0d7aaab792` returned ten
`google_places_live` media objects with URLs and attribution, and the rendered
Cinematic preview HTML contained `lh3.googleusercontent.com` plus Google Maps
attribution. `?template=4` no longer renders Italian Villa content and falls
back to the default template. Public template variants 1/2/3, admin login, the
public restaurant API, and voice health all returned HTTP 200. API, admin,
template, OCR worker, worker, scrape worker, voice, Postgres, Redis, corporate
site, and catalog containers were running with zero restarts and no recent
panic/fatal/traceback/error log lines in the checked tails.

**Checks Run:** Template TypeScript passed; admin TypeScript and ESLint passed;
`go test ./backend/...` passed with 169 tests across 44 packages; focused Python
outreach/OCR unit tests passed with 10 tests; `git diff --check` passed. VM
Docker builds passed for migrate, API, admin-web, template, and OCR worker while
removing the template; the final API and template image rebuilds passed for
`88b58eb`.

**Business Value / Plan Fit:** The low-quality experimental design was removed
quickly from the sales surface, while generated previews from the lead table now
show real, attributed restaurant imagery immediately. Published/token-gated
demos still use the stricter reviewed-media path, preserving the Phase 1
media-safety policy.

**Risks / Follow-ups:** The preview fallback makes live Places media requests for
admin-opened generated previews before OCR classification is complete; it does
not copy, cache, or publish those URLs. The OCR worker remains at the
PostgreSQL-enforced 200/200 daily request ceiling until the UTC reset, after
which reviewed public media can be rebuilt through the normal fingerprinted
path. The production storage provider is still disabled, so owner/licensed
uploads remain pending bucket/CDN configuration.
