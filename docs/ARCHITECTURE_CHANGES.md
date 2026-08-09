# Crucial Architecture Changes

Date: 2026-08-08

This is the short current contract for restaurant lead outreach, restaurant
media, and the public digital-footprint review.

| Area | Current contract |
| --- | --- |
| Lead ingestion | Google Places remains the discovery source. Apollo runs afterward only when owner or work-email details are missing. A no-match never discards a Places lead. Imports persist a nonempty inferred-business source record and enroll only restaurants with a name and valid business email. |
| Outreach eligibility | `lead` and `emailed` restaurants with recorded `inferred_business` evidence are eligible. Expressed interest pauses automation. Suppressed, lost, archived, onboarding, and active-client restaurants are excluded. OCR, profile approval, demo publication, and legacy campaign readiness are not gates. |
| Email content | Outreach is a versioned, administrator-approved, plain-text sequence. It starts with three concise approved messages but supports adding, removing, disabling, and reordering steps. Each enabled message renders exactly two direct URLs: `https://tuvisolutions.com` and its recipient-specific unsubscribe URL. No HTML body, redirect link, open pixel, ABN, or postal address is added. |
| Addressing | The renderer greets a known owner by first name. If owner details are absent, it uses `Hi {restaurant name} team,`. |
| Sequence progress | Each enrollment stores integer `current_step` and `next_step`, plus last-send and next-send timestamps. Only confirmed provider acceptance advances a step. Failed, skipped, or ambiguous delivery does not advance it. The next enabled step defaults to a 72-hour delay. |
| Send ordering | Any unfinished recipient follow-up phase blocks first messages to new recipients, including while the follow-up is waiting for its due time. This completes the existing list before starting new restaurants. |
| Runtime control | A persisted admin switch is the authoritative outreach gate. Disabling it cancels deferred work and prevents another provider boundary; enabling it creates or safely resumes one fenced bulk workflow. Deployment verification never enables it or sends to real leads. |
| Sequence versions | Editing creates a draft version. Approval archives the previous active version, moves only untouched enrollments to the new version, and leaves in-progress recipients pinned to the immutable version they already received. |
| Unsubscribe | The email URL opens a non-mutating confirmation page with an opt-out button and a Tuvi Solutions website link. Only the POST confirmation suppresses the exact recipient and stops its campaign; repeated confirmation is safe. |
| Admin portal | The outreach page provides sequence draft/version editing, add/remove/reorder/enable controls, delays, preview and approval, recipient progress, sender health, and the persisted email-job switch. Restaurant media is approved or rejected manually. |
| OCR | OCR workers, cron wrappers, provider code, image-classification jobs, configuration, and provider credentials are retired. Historical database columns and old migrations remain temporarily for audit and rollback compatibility but have no runtime role. |
| Restaurant media | Scrapers persist text/menu facts without third-party image URLs or bytes. Public Google photos are resolved live with attribution and are not stored. Durable public media must be owner-granted or licensed and manually approved. Legacy scraped images fail closed on public API, report, and template boundaries. |
| Public AI review | The digital-footprint report runs independent sources concurrently under a 15-second server budget. Same-place requests coalesce, global report/browser work is bounded, and slow providers produce a clearly labeled conservative partial result. Chromium runs as an unprivileged sandboxed user behind DNS-rebinding-resistant public-network enforcement. |
| Mobile report | Restaurant identity, score/status, live attributed photos, and map are placed near the top of the mobile report instead of being hidden below long analysis sections. |
| AI provider | The API reads the vision-capable model and key only from protected host configuration. The production rollout reuses the existing protected OpenAI key in place; no key is copied into source or logs. |

Operational details and rollback checks are in
[lead-scrape-outreach.md](runbooks/lead-scrape-outreach.md) and
[vm-deployment-plan.md](runbooks/vm-deployment-plan.md).
