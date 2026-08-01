# ADR: PostgreSQL-Only Consultation Slot Authority

Date: 2026-07-15
Status: Accepted

## Context

Tuvi consultation bookings must immediately remove confirmed times from future
availability. Google Calendar synchronization is intentionally deferred.

## Decision

PostgreSQL is the sole consultation availability and booking authority. Slots
are generated from configured business hours and confirmed rows in
`company_consultations` are removed from availability. A partial unique index
on `slot_start` for `status = 'confirmed'` prevents concurrent double booking
while allowing a cancelled slot to be booked again. The API does not query
Google FreeBusy or create Google Calendar events.

## Options Considered

- PostgreSQL-only slot ledger — selected for deterministic behavior now.
- Google Calendar plus PostgreSQL — deferred until credentials, reconciliation,
  and failure handling are ready.
- Google Calendar only — rejected because application bookings require a durable
  internal source of truth.

## Consequences

- Confirmed database bookings immediately disappear from availability.
- External calendar events do not affect availability during this phase.
- Calendar links and event IDs remain empty until the future integration.
- A future integration must treat PostgreSQL as authoritative and add explicit
  synchronization and reconciliation behavior.

## Rollback / Revisit Trigger

Revisit when Google Calendar credentials, operational ownership, reconciliation,
and failure recovery are approved for production.
