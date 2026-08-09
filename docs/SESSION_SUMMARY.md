# Session Summary

The production corporate website is deployed from `64c2dc2` at `/opt/tuvi/releases/monorepo-64c2dc2`; only `tuvi-website` was recreated, and every API, worker, database, voice, QA, and infrastructure container remained unchanged.
Restaurant growth stories now auto-scroll horizontally with a visible pause control, pause on hover/focus, and a reduced-motion manual-scroll fallback; duplicate marquee links are excluded from the accessibility tree.
“More Repeat Orders” now uses the latest horizontal five-stop Jules flow and no longer renders the stale Celine/Ciara portrait; mobile production QA confirms the complete flow fits without clipping.
The latest feature-branch commit `1116cb9` was audited and selectively integrated because its full diff regressed the report experience and included unrelated public contact code.
The comprehensive digital-footprint review redesign remains unreleased pending approval of a Google-compliant data/AI architecture; no provider, billing, migration, API, worker, or AI-review production change was made in this release.
