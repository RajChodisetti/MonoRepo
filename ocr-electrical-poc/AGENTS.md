# Electrical OCR Proof-of-Concept Instructions

These instructions specialize the repository-root `AGENTS.md` for
`ocr-electrical-poc/`.

## Scope and Status

This is a standalone FastAPI/Gemini experiment that converts electrical meter
and utility-panel images into structured JSON. It is unrelated to the retired
restaurant-menu OCR pipeline and is not part of the production restaurant
deployment topology.

## Safety and Data Handling

- Do not call Gemini or another billable provider unless the user explicitly
  authorizes that external action. Prefer the checked-in synthetic sample and
  fixture-only validation.
- Real utility images and extracted identifiers such as NMI values can be
  sensitive. Never upload, log, commit, or use real customer images as fixtures.
- Keep `.env`, API keys, `.venv`, uploads, exports, and generated fine-tuning
  data local. Never print credential values.
- Keep provider behavior behind `app/gemini_client.py` and structured output
  aligned with `app/schema.py`; prompt changes must preserve fail-closed schema
  validation.

## Safe Checks

Run from the repository root without provider calls:

```bash
rtk python3 -m compileall -q -x '(^|/)(\.venv|__pycache__)(/|$)' ocr-electrical-poc
```

The existing CLI smoke invokes the provider and therefore requires separate
approval. There is no automated offline test suite yet.
