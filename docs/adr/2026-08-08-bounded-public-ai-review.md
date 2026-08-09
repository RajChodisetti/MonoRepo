# ADR: Bounded public AI digital-footprint review

Date: 2026-08-08
Status: Accepted

## Context

The public Tuvi website offers a restaurant digital-footprint review. Its
sequential Places, website, screenshot, and AI stages could hold the report page
for more than a minute. In production, the API image also lacked a browser
runtime, so screenshot-backed analysis could silently degrade to deterministic
fallbacks. On mobile, the most persuasive evidence—restaurant identity, photos,
and map—sat below secondary report content.

The desired experience is approximately 15 seconds without fabricating results
when a restaurant website or upstream provider is slow.

## Decision

- Run independent discovery and website-audit work concurrently under one
  bounded report deadline.
- Return useful, source-labelled partial results when optional work misses its
  budget. A slow or blocked website must not extend the whole request to the old
  minute-scale timeout.
- Package Chromium in the production API image so browser-backed website capture
  uses the same runtime locally and in production.
- Route the optional visual summary through OpenAI `gpt-4.1-nano`, using low-detail
  image input and a small output budget. Preserve deterministic scoring and
  explicit partial/fallback status when the model is unavailable.
- Reuse the existing protected production OpenAI credential by mapping it to the
  API's `LLM_API_KEY` in an API-only mode-`0600` environment file. Never copy it
  into source control, build arguments, logs, browser-visible configuration, or
  unrelated worker/frontend environments.
- Move restaurant identity, progress, live photos, and map into the mobile
  report's first viewport; keep detailed findings below them.

## Options Considered

- Keep the previous sequential pipeline: rejected because one slow dependency
  consumed the whole interaction budget.
- Remove website evidence entirely: rejected because the report is specifically a
  digital-footprint review and should use direct evidence when available.
- Use a larger reasoning model: rejected for this bounded summarization task
  because latency matters more than deep multi-step reasoning.
- Guarantee completion in 15 seconds: rejected because upstream networks are not
  deterministic. The product targets roughly 15 seconds and labels partial data
  rather than claiming evidence it did not receive.

## Consequences

- Fast providers can produce the complete report near the target; slow providers
  produce a useful partial report within the bounded request window.
- The API image is larger because it includes Chromium.
- OpenAI usage remains optional and bounded. Losing its credential or quota
  degrades the narrative summary, not the underlying report route.
- Mobile visitors see the restaurant and visual proof before long-form findings.

## Rollback / Revisit Trigger

Set `LLM_PROVIDER=disabled` to remove model calls without rolling back the report.
Restore the previous release if browser packaging or concurrency causes runtime
instability. Revisit the model or time budget when measured production latency and
completion telemetry justify a different trade-off.
