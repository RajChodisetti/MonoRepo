# Twilio calling setup (Andre)

Phone features need a **public HTTPS** base URL. Localhost alone cannot receive Twilio webhooks or Media Streams.

## 1. Tunnel the voice agent

```bash
# Agent on :8001
ngrok http 8001
```

Set in `.env`:

```bash
PUBLIC_BASE_URL=https://<your-subdomain>.ngrok-free.app
CALL_API_SECRET=<shared-secret>
PORT=8001
```

No trailing slash on `PUBLIC_BASE_URL`.

## 2. Twilio Console

1. Open the phone number matching `TWILIO_PHONE_NUMBER`.
2. Voice → A call comes in → Webhook: `POST https://<PUBLIC_BASE_URL>/twiml`
3. Save.

Outbound calls use inline TwiML from `POST /call` (Stream → `wss://…/stream` with `agent=andre` and `language=…`).

## 3. Outbound test

```bash
# With agent running locally (dial hits localhost; Twilio still needs PUBLIC_BASE_URL HTTPS)
python dial.py +919876543210 --language hi

# Or Gradio → Call me tab
python ui.py
```

`POST /call` requires header `X-Call-Api-Key: $CALL_API_SECRET` when the secret is set.

Body example:

```json
{
  "to": "+919876543210",
  "language": "te",
  "campaign_id": "demo",
  "skip_compliance": true
}
```

## 4. Common failures

| Symptom | Likely cause |
|---------|----------------|
| 503 / TTS errors | Invalid `SARVAM_API_KEY` or `OPENAI_API_KEY` |
| No ring after dial | Agent down, wrong `CALL_API_SECRET`, bad Twilio creds |
| Inbound silence | Webhook not pointing at `/twiml`, or `PUBLIC_BASE_URL` not HTTPS |
| Stream fails | ngrok URL mismatch vs `PUBLIC_BASE_URL` |

## 5. Language on the call

- Inbound: uses `DEFAULT_LANGUAGE` from `.env` (default `auto`)
- Outbound: pass `language` in `/call`, `dial.py --language`, or Gradio

## Language on dial

Pass language when placing a call:

```bash
python dial.py +919876543210 --language hi
```

Or `POST /call` with `{ "to": "+91…", "language": "te" }`.
Inbound: point Twilio to `/twiml?language=hi`.
