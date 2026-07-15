# Session Summary

**Latest (2026-07-15):** PostgreSQL is now the sole authority for consultation availability and confirmed bookings; calendar API reads/writes are deferred.
**Booking safety:** Confirmed slots are excluded from later availability and protected by a confirmed-only unique index, while cancelled slots can be reused.
**UI/voice:** Browser bookings use the required name/email/phone form, confirm only after API success, and the Services dropdown closes on selection.
**Verification:** All 146 backend tests pass; production deployment and smoke results are recorded in the delivery log.
