# Twilio calling setup (Tuvi website + restaurant template)

Phone features need a **public HTTPS** base URL. Localhost alone cannot receive Twilio webhooks or Media Streams.

## 1. Tunnel the voice agent

```bash
# Example with ngrok (agent on :8000)
ngrok http 8000
```

Set in `voice-sales-agent/.env`:

```bash
PUBLIC_BASE_URL=https://<your-subdomain>.ngrok-free.app
CALL_API_SECRET=<same-secret-as-website-or-template>
```

`PUBLIC_BASE_URL` must match the URL Twilio hits (no trailing slash issues — trailing slash is stripped in code).

## 2. Twilio Console

1. Open the phone number matching `TWILIO_PHONE_NUMBER`.
2. Voice → A call comes in → Webhook: `POST https://<PUBLIC_BASE_URL>/twiml`
3. Save.

Outbound calls use inline TwiML from `POST /call` (Stream → `wss://…/stream` with `agent=corporate` or `agent=restaurant`).

### Restaurant caller ID (template)

Scraped restaurant business phones can appear as the outbound caller ID **only** if they are verified in Twilio:

1. Twilio Console → Phone Numbers → Verified Caller IDs → add each restaurant line.
2. Add the same numbers (E.164) to `TWILIO_VERIFIED_CALLER_IDS` in `voice-sales-agent/.env`:

```bash
TWILIO_VERIFIED_CALLER_IDS=+61293279713,+61212345678
```

If a restaurant phone is not verified, calls still work using `TWILIO_PHONE_NUMBER`; the template UI shows the restaurant name and listed number.

## 3. Website env (`tuvi-website/app/.env.local`)

```bash
NEXT_PUBLIC_CALL_IN_NUMBER=+61XXXXXXXXX   # display / tel: only
NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000
CALL_API_SECRET=<same as voice-sales-agent>
# optional server-only override:
# VOICE_AGENT_URL=http://localhost:8000
```

## 4. Template env (`template/.env.local`)

```bash
NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000
NEXT_PUBLIC_API_URL=http://localhost:8080
CALL_API_SECRET=<same as voice-sales-agent>
# optional:
# VOICE_AGENT_URL=http://localhost:8000
```

Never put `TWILIO_AUTH_TOKEN` or `CALL_API_SECRET` in `NEXT_PUBLIC_*` vars.

## 5. Manual test matrix

| Path | Expect |
|------|--------|
| Call Twilio number | Corporate greeting (not restaurant sales) |
| Tuvi website “Call me now” form | Phone rings → corporate agent |
| Template `/?id=0` → phone icon → Call me | Phone rings → restaurant receptionist for that index |
| Browser voice: say “call me” → give number (corporate) | Phone rings → corporate |
| Browser voice: say “call me” (restaurant template) | Phone rings → restaurant receptionist |
| Invalid / opt-out / outside ACMA window | Clear message (or `skip_compliance` in development) |
| Wrong / missing HTTPS `PUBLIC_BASE_URL` | Twilio cannot connect Media Stream |

## 6. Common failures

- **503 / TTS errors**: Cartesia/Deepgram/OpenAI keys invalid.
- **No ring after form**: agent down, wrong `CALL_API_SECRET`, or Twilio credentials.
- **Inbound silence / wrong persona**: webhook not pointing at `/twiml`, or stream missing `agent=corporate`.
- **Wrong restaurant on phone call**: check `restaurant_index` in template URL `?id=N` and MonoRepo API `GET /api/public/v1/site/restaurants/{index}`.
