# Session Summary

**Latest (2026-07-15):** PostgreSQL is now the sole authority for consultation availability and confirmed bookings; calendar API reads/writes are deferred.
**Booking safety:** Confirmed slots are excluded from later availability and protected by a confirmed-only unique index, while cancelled slots can be reused.
**UI/voice:** Browser bookings use the required name/email/phone form, confirm only after API success, and the Services dropdown closes on selection.
**Production:** Migration 25 and commit `000cdd8` are deployed; API, voice agent, and website are healthy with zero restarts and public availability returns 16 current slots.
