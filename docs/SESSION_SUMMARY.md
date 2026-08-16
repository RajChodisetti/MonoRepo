Release `2ffc134` remains deployed to QA API and production API/worker/admin at schema 54; outreach remains disabled and no provider send was performed.
Local branch `codex/fix-outreach-signature-test-send-greeting-20260816` now rehydrates saved signatures, distinguishes selected versus active test versions, and verifies the persisted signature in preview.
Definitive Gmail credential/auth failures skip to other configured accounts for direct tests and quarantine/defer safely for scheduled work; ambiguous outcomes still fail closed.
Template 1 surfaces now show only the canonical `greeting01`, with preview/test/live parity tests; the change is unreleased and awaits approved deployment plus a test-recipient smoke.
