from __future__ import annotations

import mimetypes
import os
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
from PIL import Image
import io

from app.gemini_client import analyze_image
from app.schema import ElectricalAnalysis

ROOT = Path(__file__).resolve().parent.parent
load_dotenv(ROOT / ".env")

app = FastAPI(
    title="Tuvi Electrical OCR POC",
    description="Gemini vision → structured electrical / utility site extraction",
    version="0.1.0",
)

static_dir = ROOT / "static"
static_dir.mkdir(exist_ok=True)
app.mount("/static", StaticFiles(directory=str(static_dir)), name="static")

ALLOWED = {"image/jpeg", "image/png", "image/webp", "image/gif"}


@app.get("/")
def index() -> FileResponse:
    return FileResponse(static_dir / "index.html")


@app.get("/health")
def health() -> dict:
    key = (os.getenv("GEMINI_API_KEY") or "").strip()
    return {
        "ok": True,
        "model": os.getenv("GEMINI_MODEL") or "gemini-3.1-flash-lite",
        "api_key_configured": bool(key) and not key.startswith("your_"),
    }


@app.post("/analyze", response_model=ElectricalAnalysis)
async def analyze(file: UploadFile = File(...)) -> ElectricalAnalysis:
    content_type = (file.content_type or "").split(";")[0].strip().lower()
    if content_type not in ALLOWED:
        # Guess from filename
        guess, _ = mimetypes.guess_type(file.filename or "")
        content_type = (guess or "").lower()
        if content_type not in ALLOWED:
            raise HTTPException(
                status_code=400,
                detail=f"Unsupported file type: {file.content_type}. Use JPEG/PNG/WebP.",
            )

    raw = await file.read()
    if not raw:
        raise HTTPException(status_code=400, detail="Empty file")
    if len(raw) > 15 * 1024 * 1024:
        raise HTTPException(status_code=400, detail="File too large (max 15MB)")

    # Normalize to JPEG for consistent Gemini input when needed
    mime = content_type
    image_bytes = raw
    if content_type != "image/jpeg":
        try:
            img = Image.open(io.BytesIO(raw))
            if img.mode not in ("RGB", "L"):
                img = img.convert("RGB")
            elif img.mode == "L":
                img = img.convert("RGB")
            buf = io.BytesIO()
            img.save(buf, format="JPEG", quality=92)
            image_bytes = buf.getvalue()
            mime = "image/jpeg"
        except Exception as exc:
            raise HTTPException(status_code=400, detail=f"Could not read image: {exc}") from exc

    try:
        return analyze_image(image_bytes, mime_type=mime)
    except RuntimeError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc
    except Exception as exc:
        raise HTTPException(status_code=500, detail=f"Analysis failed: {exc}") from exc
