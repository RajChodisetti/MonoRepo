# Inbound voice setup (Tuvi website + restaurant template)

Phase 1 supports inbound voice only. Browser WebSocket sessions and calls made
by a person to the configured Twilio number are supported. The service never
dials a visitor: `POST /call` returns `403`, callback tools are absent, and the
central dial function fails closed.

## 1. Public HTTPS endpoint

Twilio webhooks and Media Streams require a public HTTPS base URL. For local
inbound testing, expose port 8000 with a tunnel and set:

```bash
PUBLIC_BASE_URL=https://<your-subdomain>.ngrok-free.app
```

`PUBLIC_BASE_URL` must match the URL Twilio reaches. The runtime strips a
trailing slash.

## 2. Twilio Console

1. Open the phone number matching `TWILIO_PHONE_NUMBER`.
2. Set Voice → A call comes in to
   `POST https://<PUBLIC_BASE_URL>/twiml`.
3. Save.

Do not configure an outbound callback webhook or expose provider credentials to
the browser.

## 3. Website configuration

Root corporate site:

```bash
NEXT_PUBLIC_CALL_IN_NUMBER=+61XXXXXXXXX
NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000
# Optional server-only override:
VOICE_AGENT_URL=http://localhost:8000
```

Restaurant template:

```bash
NEXT_PUBLIC_VOICE_AGENT_URL=http://localhost:8000
NEXT_PUBLIC_API_URL=http://localhost:8080
# Optional server-only override:
VOICE_AGENT_URL=http://localhost:8000
```

Never put `TWILIO_AUTH_TOKEN` or another provider credential in a
`NEXT_PUBLIC_*` variable.

## 4. Read-only test matrix

| Path | Expected result |
| --- | --- |
| Call the Twilio number | Inbound corporate greeting |
| Tuvi website voice widget | Browser connects to the corporate agent |
| Restaurant template voice widget | Browser connects to the selected restaurant receptionist |
| Browser asks for consultation times | Agent reads only database-backed availability |
| `POST /call` with or without credentials | `403` disabled; no provider call |
| Wrong or missing HTTPS `PUBLIC_BASE_URL` | Twilio cannot connect the inbound Media Stream |

Deployment smoke tests must not create a fake booking, place a phone call, or
submit personal data. A real inbound Twilio call requires separate approval.

## 5. Common failures

- **503 / TTS errors:** Cartesia, Deepgram, or OpenAI credentials are invalid.
- **Inbound silence / wrong persona:** the webhook is not pointing at `/twiml`,
  or the stream is missing the expected agent mode.
- **Wrong restaurant in browser voice:** check `restaurant_index` in the
  template URL and the MonoRepo restaurant endpoint.
