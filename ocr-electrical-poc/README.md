# Tuvi Electrical Circuit OCR POC

Standalone Gemini vision POC that turns utility / meter-board photos into **structured JSON** (NMI, fuses, brands, unlabeled objects like gas pipes).

## Quick start

```bash
cd /Users/praveenmaurya/Desktop/Tuvi/MonoRepo/ocr-electrical-poc
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# Put your Gemini API key in .env → GEMINI_API_KEY=...
# Default model: gemini-3.1-flash-lite (override with GEMINI_MODEL)
uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload
```

Open **http://localhost:8090**

## CLI smoke test

```bash
source .venv/bin/activate
python scripts/smoke_test.py examples/meter_panel_sample.png
```

## API

- `GET /health` — key configured?
- `POST /analyze` — multipart field `file` (JPEG/PNG/WebP)

## Fine-tune later

See [`dataset/README.md`](dataset/README.md) and:

```bash
python scripts/prepare_finetune_jsonl.py --out exports/finetune.jsonl
```

## Security

Never commit `.env`. If a key was pasted in chat, rotate it in Google AI Studio.
