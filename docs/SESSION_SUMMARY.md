Release `2ffc134` is deployed to the QA API and production API/worker/admin at schema 54; `master` and the feature branch are reconciled to that commit.
The inbox is newest-received first, shows subject/text/from/by, opens complete stored text, and preserves page/detail state across manual and 15-second refreshes.
Template tests and same-mailbox replies bypass scheduled quota state while scheduled outreach still claims durable quota; the Sydney window remains 07:00-12:00.
Outreach remains disabled with zero active jobs/due health checks/send changes; production polling is 15 seconds and the `a07fffc` release plus explicit images/backups remain available for rollback.
