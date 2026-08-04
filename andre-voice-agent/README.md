# Andre Real Estate Voice Agent

Multilingual Indian **real-estate calling** agent that qualifies preferences one question at a time, searches a seeded inventory (Hyderabad + Bengaluru), and speaks matches aloud.

Stack:

- **Sarvam** Saaras STT + Bulbul TTS (Hindi, English, Telugu, Hinglish)
- **OpenAI** for dialogue + tools
- **Twilio** Programmable Voice + Media Streams
- **Pipecat** realtime pipeline
- Local property inventory: [`data/properties.json`](data/properties.json)

## Quick start

```bash
cd MonoRepo/andre-voice-agent
# Pipecat/numba need Python >=3.10,<3.14 (3.12 recommended)
/opt/homebrew/opt/python@3.12/bin/python3.12 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

uvicorn bot:app --host 0.0.0.0 --port 8001
```

Or from `MonoRepo/`:

```bash
make andre-voice-dev
```

- Live UI: http://127.0.0.1:8001/  
- Health: http://127.0.0.1:8001/healthz  

Click **Start talking**, pick a language, then speak. Andre asks type → city → budget (one at a time), then searches listings.

## Language on connect (API)

| Entry | How to set language |
|-------|---------------------|
| Browser WS | `/browser-stream?language=en\|hi\|te\|auto` |
| Outbound call | `POST /call` JSON `{ "to": "+91…", "language": "hi" }` |
| Inbound Twilio | `POST /twiml?language=te` (or form field `language`) |
| Mid-call | User asks to switch → `set_language` tool updates STT/TTS |

## Property APIs

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/properties/meta` | Cities, localities, types, budget bands |
| GET | `/api/properties/search?city=&locality=&property_type=&bhk=&min_price=&max_price=&status=&budget_band=&limit=` | Search |
| GET | `/api/properties/{id}` | Full listing |

Example:

```bash
curl 'http://127.0.0.1:8001/api/properties/search?city=Hyderabad&property_type=apartment&budget_band=50L_1Cr'
```

## Calling

See [CALLING.md](CALLING.md). Keep a live HTTPS tunnel in `PUBLIC_BASE_URL`.

```bash
python dial.py +919876543210 --language hi
```

## Main voice endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Keys present for phone |
| POST | `/call` | Outbound dial (`X-Call-Api-Key`) + `language` |
| POST | `/twiml` | Twilio inbound webhook |
| WS | `/stream` | Twilio Media Streams |
| WS | `/browser-stream` | Browser mic session |

## Docker

```bash
docker compose up --build
```

Agent listens on port **8001**.

## Smoke test

1. http://127.0.0.1:8001/ → Start talking (language Hindi/English/Telugu)
2. Say you want an apartment in Hyderabad under 1 crore → hear matching listings
3. Optional phone: set tunnel `PUBLIC_BASE_URL`, then **Call me now**
