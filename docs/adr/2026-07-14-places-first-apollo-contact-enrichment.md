# ADR: Use Places-First Discovery With Targeted Apollo Contact Enrichment

Date: 2026-07-14
Status: Accepted

## Context

The previous scheduled workflow used Apollo as the broad discovery source and
Google Places for restaurant enrichment. That flow added provider cost before a
restaurant was identified and had a deduplication bug that could register a new
Apollo lead before its first Places scrape and then skip it.

Google Places is the stronger source for restaurant identity, location, website,
hours, reviews, and stable Place-ID deduplication. It usually does not provide an
owner name or direct work email, so Apollo remains useful after the restaurant is
known.

## Decision

Use Google Places API (New) Text Search for bounded city discovery and Place
Details for public business enrichment. After Places and best-effort website
email extraction, invoke Apollo only when an owner or contact email is missing.

- Match Apollo People Search by the restaurant's owned website domain and
  approved decision-maker titles.
- Use People Match on the selected Apollo person ID to retrieve full owner details
  and a work email.
- Keep personal-email and phone reveal disabled.
- Skip Apollo when both fields are already present or no unambiguous business
  domain is available.
- Deduplicate by Google Place ID against PostgreSQL and canonical JSON.
- Require both provider keys while their default-enabled stages are active.
- Send credentials in headers and never persist them in URLs or payloads.

## Options Considered

- Apollo-first discovery plus Places enrichment: rejected for the scheduled path
  because it spends Apollo calls before confirming a restaurant and complicates
  deduplication.
- Places only: rejected because owner and contact-email coverage is insufficient.
- Places-first plus targeted Apollo enrichment: selected because it keeps stable
  restaurant discovery while using Apollo credits only for missing contact data.
- General web/SerpAPI discovery: retained as legacy/manual tooling due to its less
  predictable structure and wider operational surface.

## Consequences

- Apollo remains in the scheduled pipeline, but it is no longer the broad lead
  discovery source.
- Each contact candidate can use one Apollo People Search call and one credit-
  consuming People Match call, subject to plan behavior.
- Businesses without an owned domain are not sent to Apollo to avoid ambiguous
  owner matches.
- Places still does not supply menu items; menu enrichment remains separate.
- OCR refreshes Google photo resources server-side just before resolving them
  to short-lived media URLs; neither resource names nor media URLs are cached.
  Public rendering still needs its own compliant media flow.

## Rollback / Revisit Trigger

Revisit if Apollo match accuracy, work-email yield, or credit cost misses the
approved targets. Any broader Apollo search must retain Place-ID deduplication,
data-minimization controls, tests, cost review, and explicit approval.
