# Session Summary

**Latest (2026-07-15):** Deployed functional commit `371c062`; the Google Workspace disclosure has been removed from the homepage while the dedicated app and legal pages remain public.
**UI:** The homepage is now entirely focused on practical AI and custom software, with the small Workspace app link retained only in the footer.
**Verification:** TypeScript, local and VM production builds, rendered-content checks, and loopback/public HTTP checks passed; the website has zero restarts.
**Operations:** Only `tuvi-website` changed, `7651213` is the immediate rollback, and removing the homepage purpose statement may cause Google branding re-verification to fail again.
