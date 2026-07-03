"""Load MonoRepo + backend + outreach .env files (later files override)."""

from pathlib import Path

from dotenv import load_dotenv

_MONOREPO = Path(__file__).resolve().parents[2]
_OUTREACH = Path(__file__).resolve().parent


def load_project_env() -> None:
    load_dotenv(_MONOREPO / ".env")
    load_dotenv(_MONOREPO / "backend" / ".env")
    load_dotenv(_OUTREACH / ".env")
