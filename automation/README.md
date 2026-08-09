# Automation

## Restaurant lead and outreach pipeline

See [outreach/README.md](outreach/README.md) for the durable Places-first city
worker, targeted Apollo contact enrichment, and legacy/manual lead tools.

Production city work starts through the private `POST /api/v1/scrape-jobs`
endpoint and is consumed by the long-running `scrape-worker`. It must not be
started through `city_pipeline.py` or a one-shot host cron. Follow the
[lead scrape and outreach runbook](../docs/runbooks/lead-scrape-outreach.md) for
setup, deployment, and operation.

OCR is retired and is not a lead-readiness dependency. An imported restaurant
with a name, valid business email, and recorded `inferred_business` source
evidence is enrolled in the active approved plain-text sequence. The email job
remains an explicit, persisted operator control.

## Media policy

Do not persist or publish media scraped from Google, review feeds, menus, or
other third-party listings. Google Places photos are resolved live by the Go API
with attribution. Only owner-provided or separately licensed files may use the
durable media path, and an internal admin must explicitly approve them before
public use.

The historical `automation/scrape.py` image downloader and its local catalog are
not part of the production lead pipeline and must not be scheduled or used to
populate customer-facing media.
