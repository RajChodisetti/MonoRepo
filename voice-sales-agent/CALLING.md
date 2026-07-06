# Twilio calling setup (Tuvi website)

Phone features need a **public HTTPS** base URL. Localhost alone cannot receive Twilio webhooks or Media Streams.

## 1. Tunnel the voice agent

```bash
# Example with ngrok (agent on :8000)
ngrok http 8000
```

Set in `voice-sales-agent/.env`:

```bash
PUBLIC_BASE_URL=https://<your-subdomain>.ngrok-free.app
CALL_API_SECRET=<same-secret-as-website>
```

`PUBLIC_BASE_URL` must match the URL Twilio hits (no trailing slash issues — trailing slash is stripped in code).

## 2. Twilio Console

1. Open the phone number matching `TWILIO_PHONE_NUMBER`.
2. Voice → A call comes in → Webhook: `POST https://<PUBLIC_BASE_URL>/twiml`
3. Save.

Outbound calls use inline TwiML from `POST /call` (Stream → `wss://…/stream` with `agent=corporate`).

## 3. Website env (`tuvi-website/app/.env.local`)

```bash
NEXT_PUBLIC_CALL_IN_NUMBER=+61XXXXXXXXX   # display / tel: only
NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000
CALL_API_SECRET=<same as voice-sales-agent>
# optional server-only override:
# VOICE_AGENT_URL=http://localhost:8000
```

Never put `TWILIO_AUTH_TOKEN` or `CALL_API_SECRET` in `NEXT_PUBLIC_*` vars.

## 4. Manual test matrix

| Path | Expect |
|------|--------|
| Call Twilio number | Corporate greeting (not restaurant sales) |
| Website “Call me now” form | Phone rings → corporate agent |
| Browser voice: say “call me” → give number | Phone rings |
| Invalid / opt-out / outside ACMA window | Clear message (or `skip_compliance` in development) |
| Wrong / missing HTTPS `PUBLIC_BASE_URL` | Twilio cannot connect Media Stream |

## 5. Common failures

- **503 / TTS errors**: Cartesia/Deepgram/OpenAI keys invalid.
- **No ring after form**: agent down, wrong `CALL_API_SECRET`, or Twilio credentials.
- **Inbound silence / wrong persona**: webhook not pointing at `/twiml`, or stream missing `agent=corporate`.
