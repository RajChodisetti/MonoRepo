# Session Summary

Production runtime is `91fc93a` at `/opt/tuvi/releases/monorepo-91fc93a`; rollback points to `/opt/tuvi/releases/monorepo-778c0fe`, and schema migration `000036` is applied in the `monorepo` DB.
Outreach preview/send now canonicalizes stale campaign HTML to the current three-anchor template: personalized demo website, Services catalog, and unsubscribe. Bulk and ad hoc sends render fresh current copy before tracking injection.
Migration `000036` rewrote unsent old-template campaign rows; production checks show zero draft/approved rows with legacy template labels/placeholders or any HTML link count other than three. Hotel520's token-gated demo preview returned 200 without printing the token.
Backend tests/build, admin lint/TypeScript/build, OpenAPI validation, VM image builds, deployment, and safe HTTPS smokes passed. No real email was sent during verification.
