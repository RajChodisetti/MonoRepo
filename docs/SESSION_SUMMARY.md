# Session Summary

Production runtime is `778c0fe` at `/opt/tuvi/releases/monorepo-778c0fe`; rollback points to `/opt/tuvi/releases/monorepo-e4d6801`, and schema remains at `000035` with no new migration.
Selective/manual outreach sends from the restaurant list or restaurant detail page now bypass the bulk email-job flag and generic `EMAIL_DISABLE_SENDING` flag, while keeping internal-admin, contact email, suppression, campaign draft, and configured-sender checks.
New outreach drafts now show two promotional links: tracked “Personalized demo websites” and direct “Services catalog” at `/services/restaurants`; legacy per-template placeholders still render for older drafts, and unsubscribe remains.
Backend tests/build, admin lint/TypeScript/build, OpenAPI validation, VM image builds, deployment, and safe HTTPS smokes passed. No real email was sent during verification.
