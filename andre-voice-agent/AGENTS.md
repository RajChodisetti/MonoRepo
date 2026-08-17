# Andre Voice Prototype Instructions

These instructions specialize the repository-root `AGENTS.md` for
`andre-voice-agent/` and must not weaken its global safety or approval rules.

## Scope and Status

Andre is a standalone real-estate browser/Twilio prototype on port `8001`. It
is not part of the production restaurant platform or the supported
`voice-sales-agent` service. It contains active outbound dialing, callback,
Twilio, transcript, and property-administration code.

## High-Risk Boundaries

- Never invoke `dial.py`, `POST /call`, `place_callback_call`, the Gradio
  outbound-call control, or any Twilio/provider path unless the user separately
  and explicitly approves the exact external call action.
- Never set or rely on `skip_compliance`, even in development. The current
  prototype's development bypass is legacy code, not an approved workflow.
- Do not deploy, expose, schedule, or integrate Andre into restaurant services
  without an accepted architecture decision and the root production/provider
  approvals.
- Preserve call API authentication, Twilio signature validation, calling
  windows, do-not-call state, caller-ID restrictions, duration limits, and
  manual hangup controls. A health response does not prove provider readiness.
- Phone numbers, transcripts, call logs, property leads, recordings, and local
  SQLite files are sensitive. Do not print, commit, copy into fixtures, or
  include them in summaries.

## Existing Patterns

- `bot.py` owns FastAPI/browser/Twilio orchestration; `dial.py` and `ui.py`
  contain explicit outbound entrypoints.
- `property_store.py` and `listings_api.py` own local inventory behavior;
  `prompts/` owns the Andre persona and tool schemas.
- Docker/Python 3.12 is the supported dependency environment. Avoid importing
  `bot.py` merely to inspect it because import-time initialization can create
  local state and load provider dependencies.
- Keep `.env`, `.venv`, runtime databases, logs, and generated datasets local.

## Safe Checks

Run only non-calling checks unless the user explicitly authorizes more:

```bash
rtk python3 -m compileall -q -x '(^|/)(\.venv|__pycache__)(/|$)' andre-voice-agent
rtk docker build -f andre-voice-agent/Dockerfile andre-voice-agent
```

The subtree has no automated policy suite. Use static inspection or isolated
fixtures; do not use a real phone number, provider credential, or external API
as a smoke test.
