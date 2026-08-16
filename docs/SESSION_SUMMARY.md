Release `a07fffc` remains deployed to QA and production at schema 54; scheduled outreach is disabled and the saved Australia/Sydney window remains 07:00-12:00.
Local/unreleased inbox hardening orders by the latest received email, shows subject/text/from/by, opens the complete stored text, and adds page/detail-preserving refresh plus 15-second UI and inbound-worker polling.
Local/unreleased manual-send hardening keeps template tests and same-mailbox replies entirely outside scheduled quota reconciliation, synchronization, claims, and pacing; provider limits still apply.
Backend, race, web, template, OpenAPI, Compose, and agent-context checks pass; no deployment, protected configuration mutation, provider call, or email send was performed.
