# Session Summary

Production runtime is `5bfe1dc` at `/opt/tuvi/releases/monorepo-5bfe1dc`; the active tree uses canonical `web/`, has no legacy `tuvi-website/`, and the application database is at migration `000041`.
The internal monthly consultation calendar shows confirmed calls, persists slot availability with revision checks, and feeds the same PostgreSQL availability/booking path used by the public site and inbound voice agent.
The latest QA website is live; public callback/outbound endpoints remain HTTP 403. A Pipecat 0.0.108 `ToolsSchema` adapter restored live voice calendar queries while preserving inbound-only behavior.
Builds, migrations, UI/API/constraint checks, HTTPS and browser smokes, five voice safety tests, and a non-PII production voice availability lookup passed. No real booking, outbound call, email, or SMS was created.
Migrations `000039`–`000041` are forward-only; retain the validated pre-deploy backup and prefer a forward fix or approved restore over a pre-calendar API rollback once slot overrides are used.
