# Voice Agent Instructions

These instructions specialize the repository-root `AGENTS.md` for everything
under `voice-sales-agent/` and must not weaken its global safety or approval
rules.

## Scope

This FastAPI/Pipecat service provides browser voice and Twilio-capable voice on
port `8000`. Docker is the supported runtime because its Python and native media
dependencies are version-sensitive. Call logs and transcripts are sensitive.

## Existing Patterns

- Keep corporate and restaurant personas separate under `prompts/`; select them
  through the existing `agent`/`restaurant_index` protocol rather than mixing
  prompt data.
- Backend data access belongs in `api_client.py` (restaurant public APIs) and
  `tuvi_api_client.py` (token-protected company consultations).
- Keep browser WebSocket event shapes aligned with both Next.js voice hooks.
- The restaurant assistant must identify itself as the restaurant's AI
  assistant, use approved restaurant data only, ask one question at a time, and
  create `pending` reservation requests only. It must never say a table is
  booked or confirmed.
- Escalate unknown, angry, payment-related, sensitive, or explicit human
  requests. Never pretend to be human or collect payment credentials.

## Outbound Call Prohibition

The runtime accepts inbound Twilio/browser sessions and retains human-transfer
support. `POST /call` is deliberately retained as a fail-closed `403`; the
public dialer, `dial.py`, callback tools, and SMS tools are retired. Historical
design docs are not authorization to restore them.

- Do not invoke, test with a real number, expose, enable, schedule, or expand
  any outbound call/SMS flow without explicit human approval for that action.
- Do not use `skip_compliance`, even in development, to place an external call.
- Customer-facing prompt/model/tool-route changes also require explicit
  approval under the root contract.
- Safe local browser-WebSocket work is allowed when it cannot contact a phone,
  send a message, or mutate production data.

## Dependency Impact

- Browser event/readiness/query changes affect `template/` and `web/`; update
  both hooks, widgets, and server proxies together.
- Restaurant reservation tools depend on the Go public site, availability, and
  pending-reservation APIs. Consultation tools depend on the main Go company
  consultation endpoints and `TUVI_API_TOKEN`.
- Prompt or tool-schema changes need regression cases for AI disclosure, hours,
  known/unknown menu items, missing booking details, outside-hours requests,
  pending-only wording, human escalation, angry callers, and unknown fallback.
- Preserve Twilio signature validation, call-secret checks, opt-out state,
  calling-window controls, and transcript access protection.

## Safety

- Never print, commit, or return provider keys, call secrets, phone/email data,
  call recordings, or transcripts. Keep `.env`, SQLite data, and logs local.
- Do not infer production readiness from `/healthz`; browser and provider
  readiness have separate checks.

## Checks

Run from the repository root without provider calls:

```bash
rtk python3 -m compileall -q voice-sales-agent
rtk python3 -m unittest discover -s voice-sales-agent/tests -p 'test_*.py'
rtk docker compose -f infra/docker/docker-compose.yml --profile voice config --quiet
rtk docker compose -f infra/docker/docker-compose.yml --profile voice build voice-sales-agent
```

The inbound-only suite statically verifies fail-closed outbound behavior and
the absence of the public dialer/SMS tools. For browser-protocol changes, also
run a local browser-only smoke with test restaurant data and confirm no `/call`,
SMS, or production booking request is made.
