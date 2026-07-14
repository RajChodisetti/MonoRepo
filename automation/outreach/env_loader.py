"""Load project ingestion settings without overriding the process environment."""

import os
from pathlib import Path

from dotenv import dotenv_values

_MONOREPO = Path(__file__).resolve().parents[2]
_OUTREACH = Path(__file__).resolve().parent


def load_project_env() -> None:
    paths = [
        _MONOREPO / ".env",
        _MONOREPO / "backend" / ".env",
        _OUTREACH / ".env",
    ]
    configured = os.getenv("INGESTION_ENV_FILE", "").strip()
    if configured:
        paths.append(Path(configured).expanduser())

    merged: dict[str, str] = {}
    for path in paths:
        if not path.is_file():
            continue
        for key, value in dotenv_values(path).items():
            if value is not None:
                merged[key] = value

    for key, value in merged.items():
        os.environ.setdefault(key, value)
