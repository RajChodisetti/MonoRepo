# ADR: PostgreSQL-Only Consultation Slot Authority

Date: 2026-07-15
Status: Accepted
Amended: 2026-08-08

## Context

Tuvi consultation bookings must immediately remove confirmed times from future
availability. Google Calendar synchronization is intentionally deferred.

## Decision

PostgreSQL is the sole consultation availability and booking authority. The
configured timezone, weekday business hours, and slot duration define the
candidate grid. Per-slot rows in `company_consultation_slot_overrides` let an
internal administrator open or close every future candidate in a calendar
month. A missing override retains the configured-hours default so the migration
does not silently remove existing availability.

Confirmed rows in `company_consultations` are compared as half-open intervals,
`[slot_start, slot_end)`, so an older off-grid 09:17–09:47 call blocks both the
09:00 and 09:30 grid candidates it overlaps. A partial GiST exclusion constraint
on that same interval is the final protection against concurrent overlapping
confirmed bookings while allowing a cancelled interval to be booked again. A
companion check constraint rejects every row whose end is not strictly after
its start, preventing empty or reversed intervals from bypassing the overlap
rule.

Each calendar month has a revision row in
`company_consultation_calendar_months`. Admin GET returns the revision and PUT
must compare it before replacing the month in one transaction. A successful save
removes obsolete overrides, reinserts the complete current future grid, and
increments the revision. A stale save fails with `409` without changing slot
state. Canonical candidates that cross into the past after GET are validated and
ignored so a save can safely span a slot boundary. Booking and saving take
compatible locks on the month row, then the booking locks its exact override
row, so neither operation observes a partially replaced month. The API does not
query Google FreeBusy or create Google Calendar events.

## Options Considered

- PostgreSQL-only slot ledger — selected for deterministic behavior now.
- Persisted per-slot overrides with configured-hours fallback — selected to add
  administrator control without closing every slot during rollout.
- Fully explicit availability rows with no fallback — deferred because it would
  require populating availability before existing booking traffic can continue.
- Google Calendar plus PostgreSQL — deferred until credentials, reconciliation,
  and failure handling are ready.
- Google Calendar only — rejected because application bookings require a durable
  internal source of truth.

## Consequences

- Confirmed database bookings immediately disappear from availability.
- One confirmed call may block multiple candidate slots but is counted once in
  the admin month summary; its actual off-grid start remains visible.
- Internal administrators can save a complete month of slot overrides, while
  past and off-grid bookings remain read-only calendar evidence.
- Simultaneous admin sessions cannot silently overwrite one another; the stale
  editor must refresh and intentionally reapply changes.
- Public website and voice-agent availability/check/booking calls use the same
  persisted overrides without changing their response contract.
- Changing configured business hours can introduce newly valid candidates with
  no override; those candidates use the configured-hours fallback until that
  month is saved again, at which point obsolete grid overrides are removed.
- External calendar events do not affect availability during this phase.
- Calendar links and event IDs remain empty until the future integration.
- A future integration must treat PostgreSQL as authoritative and add explicit
  synchronization and reconciliation behavior.

## Rollback / Revisit Trigger

Revisit when Google Calendar credentials, operational ownership, reconciliation,
and failure recovery are approved for production.
