"""
================================================================================
  CALL LOGGER — Saves transcripts, outcomes, and opt-outs to local SQLite.
  Swap DATABASE_URL with PostgreSQL + Redis in production.

  Tables:
    calls                    — one row per phone/browser session
    transcripts              — turns keyed by call_id (legacy)
    conversation_transcripts — every user/assistant line with phone and/or email
    opt_outs                 — do-not-call list
================================================================================
"""

from __future__ import annotations

import json
import os
import sqlite3
from datetime import datetime

DB_PATH = os.environ.get("CALL_LOG_DB", "calls.db")


def _connect() -> sqlite3.Connection:
    con = sqlite3.connect(DB_PATH)
    con.row_factory = sqlite3.Row
    return con


def _ensure_column(cur: sqlite3.Cursor, table: str, column: str, decl: str) -> None:
    cols = {row[1] for row in cur.execute(f"PRAGMA table_info({table})").fetchall()}
    if column not in cols:
        cur.execute(f"ALTER TABLE {table} ADD COLUMN {column} {decl}")


def init_db() -> None:
    """Create tables / migrate columns if they don't exist."""
    con = _connect()
    cur = con.cursor()
    cur.executescript(
        """
        CREATE TABLE IF NOT EXISTS calls (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            call_sid    TEXT,
            to_number   TEXT,
            started_at  TEXT,
            ended_at    TEXT,
            duration_s  REAL,
            outcome     TEXT,
            booking     TEXT,
            created_at  TEXT DEFAULT (datetime('now'))
        );

        CREATE TABLE IF NOT EXISTS transcripts (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            call_id     INTEGER REFERENCES calls(id),
            role        TEXT,
            content     TEXT,
            ts          TEXT DEFAULT (datetime('now'))
        );

        CREATE TABLE IF NOT EXISTS conversation_transcripts (
            id            INTEGER PRIMARY KEY AUTOINCREMENT,
            call_id       INTEGER REFERENCES calls(id),
            call_sid      TEXT,
            channel       TEXT,
            phone_number  TEXT,
            email         TEXT,
            role          TEXT,
            content       TEXT,
            agent_mode    TEXT,
            ts            TEXT DEFAULT (datetime('now'))
        );

        CREATE TABLE IF NOT EXISTS opt_outs (
            id           INTEGER PRIMARY KEY AUTOINCREMENT,
            phone        TEXT NOT NULL,
            call_id      INTEGER REFERENCES calls(id),
            opted_out_at TEXT DEFAULT (datetime('now'))
        );

        CREATE INDEX IF NOT EXISTS idx_opt_outs_phone ON opt_outs (phone);
        CREATE INDEX IF NOT EXISTS idx_conv_phone ON conversation_transcripts (phone_number);
        CREATE INDEX IF NOT EXISTS idx_conv_email ON conversation_transcripts (email);
        CREATE INDEX IF NOT EXISTS idx_conv_call ON conversation_transcripts (call_id);
        CREATE INDEX IF NOT EXISTS idx_conv_ts ON conversation_transcripts (ts);
        """
    )
    for col, decl in (
        ("email", "TEXT"),
        ("channel", "TEXT"),
        ("agent_mode", "TEXT"),
        ("contact_name", "TEXT"),
    ):
        _ensure_column(cur, "calls", col, decl)
    for col, decl in (
        ("phone_number", "TEXT"),
        ("email", "TEXT"),
    ):
        _ensure_column(cur, "transcripts", col, decl)
    con.commit()
    con.close()


def start_call(
    call_sid: str,
    phone_number: str = "",
    *,
    email: str = "",
    channel: str = "phone",
    agent_mode: str = "",
    contact_name: str = "",
) -> int:
    """Insert a new call/session row and return its DB id."""
    init_db()
    phone = (phone_number or "").strip()
    mail = (email or "").strip().lower()
    if channel == "browser" and not phone:
        phone = "browser"
    con = _connect()
    cur = con.cursor()
    cur.execute(
        """
        INSERT INTO calls (
            call_sid, to_number, started_at, outcome, email, channel, agent_mode, contact_name
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            call_sid,
            phone,
            datetime.utcnow().isoformat(),
            "unknown",
            mail or None,
            channel,
            (agent_mode or "").strip() or None,
            (contact_name or "").strip() or None,
        ),
    )
    call_id = int(cur.lastrowid)
    con.commit()
    con.close()
    return call_id


def _call_meta(con: sqlite3.Connection, call_id: int) -> dict:
    row = con.execute(
        "SELECT call_sid, to_number, email, channel, agent_mode FROM calls WHERE id = ?",
        (call_id,),
    ).fetchone()
    if not row:
        return {
            "call_sid": "",
            "phone_number": "",
            "email": "",
            "channel": "",
            "agent_mode": "",
        }
    phone = (row["to_number"] or "").strip()
    if phone.lower() in {"browser", "unknown"}:
        phone = ""
    return {
        "call_sid": row["call_sid"] or "",
        "phone_number": phone,
        "email": (row["email"] or "").strip().lower(),
        "channel": row["channel"] or "",
        "agent_mode": row["agent_mode"] or "",
    }


def log_turn(call_id: int, role: str, content: str) -> None:
    """
    Append one transcript line for a session.
    Writes legacy `transcripts` and denormalized `conversation_transcripts`
    (with phone_number / email from the parent call row).
    """
    text = (content or "").strip()
    if call_id is None or call_id < 0 or not text:
        return

    init_db()
    con = _connect()
    meta = _call_meta(con, call_id)
    phone = meta["phone_number"] or None
    email = meta["email"] or None
    ts = datetime.utcnow().isoformat()

    con.execute(
        """
        INSERT INTO transcripts (call_id, role, content, phone_number, email, ts)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        (call_id, role, text, phone, email, ts),
    )
    con.execute(
        """
        INSERT INTO conversation_transcripts (
            call_id, call_sid, channel, phone_number, email, role, content, agent_mode, ts
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
        (
            call_id,
            meta["call_sid"] or None,
            meta["channel"] or None,
            phone,
            email,
            role,
            text,
            meta["agent_mode"] or None,
            ts,
        ),
    )
    con.commit()
    con.close()


def update_call_contact(
    call_id: int,
    *,
    phone: str | None = None,
    email: str | None = None,
    contact_name: str | None = None,
) -> None:
    """
    Attach/update phone and email on a session and backfill prior transcript rows.
    Call this when booking collects details (browser) or identity improves.
    """
    if call_id is None or call_id < 0:
        return

    init_db()
    con = _connect()
    fields: list[str] = []
    values: list[object] = []

    phone_norm = (phone or "").strip() if phone is not None else None
    email_norm = (email or "").strip().lower() if email is not None else None
    name_norm = (contact_name or "").strip() if contact_name is not None else None

    if phone_norm is not None and phone_norm and phone_norm.lower() not in {"browser", "unknown"}:
        fields.append("to_number = ?")
        values.append(phone_norm)
    if email_norm is not None:
        fields.append("email = ?")
        values.append(email_norm or None)
    if name_norm is not None:
        fields.append("contact_name = ?")
        values.append(name_norm or None)

    if fields:
        values.append(call_id)
        con.execute(f"UPDATE calls SET {', '.join(fields)} WHERE id = ?", values)

    if phone_norm and phone_norm.lower() not in {"browser", "unknown"}:
        con.execute(
            """
            UPDATE conversation_transcripts
            SET phone_number = ?
            WHERE call_id = ? AND (phone_number IS NULL OR phone_number = '')
            """,
            (phone_norm, call_id),
        )
        con.execute(
            """
            UPDATE transcripts
            SET phone_number = ?
            WHERE call_id = ? AND (phone_number IS NULL OR phone_number = '')
            """,
            (phone_norm, call_id),
        )
    if email_norm:
        con.execute(
            """
            UPDATE conversation_transcripts
            SET email = ?
            WHERE call_id = ? AND (email IS NULL OR email = '')
            """,
            (email_norm, call_id),
        )
        con.execute(
            """
            UPDATE transcripts
            SET email = ?
            WHERE call_id = ? AND (email IS NULL OR email = '')
            """,
            (email_norm, call_id),
        )

    con.commit()
    con.close()


def end_call(
    call_id: int,
    outcome: str,
    booking: dict | None = None,
    duration_s: float | None = None,
) -> None:
    """Mark the call as ended. Safe to call with the same call_id more than once."""
    if call_id is None or call_id < 0:
        return
    init_db()
    if booking:
        update_call_contact(
            call_id,
            phone=booking.get("prospect_phone") or booking.get("guest_phone"),
            email=booking.get("prospect_email") or booking.get("guest_email"),
            contact_name=booking.get("prospect_name") or booking.get("guest_name"),
        )
    con = _connect()
    con.execute(
        """
        UPDATE calls
           SET ended_at   = ?,
               outcome    = ?,
               booking    = ?,
               duration_s = ?
         WHERE id = ?
        """,
        (
            datetime.utcnow().isoformat(),
            outcome,
            json.dumps(booking) if booking else None,
            duration_s,
            call_id,
        ),
    )
    con.commit()
    con.close()


def get_call_summary(call_id: int) -> dict:
    """Return the full call record + transcript."""
    init_db()
    con = _connect()
    call = con.execute("SELECT * FROM calls WHERE id = ?", (call_id,)).fetchone()
    turns = con.execute(
        """
        SELECT role, content, ts, phone_number, email
          FROM conversation_transcripts
         WHERE call_id = ?
         ORDER BY id
        """,
        (call_id,),
    ).fetchall()
    if not turns:
        turns = con.execute(
            "SELECT role, content, ts, phone_number, email FROM transcripts WHERE call_id = ? ORDER BY id",
            (call_id,),
        ).fetchall()
    con.close()
    if not call:
        return {}
    return {
        **dict(call),
        "transcript": [dict(t) for t in turns],
    }


def list_calls(limit: int = 20) -> list[dict]:
    """Return recent calls without transcripts."""
    init_db()
    con = _connect()
    rows = con.execute(
        "SELECT * FROM calls ORDER BY id DESC LIMIT ?", (limit,)
    ).fetchall()
    con.close()
    return [dict(r) for r in rows]


def list_conversation_transcripts(
    *,
    phone: str | None = None,
    email: str | None = None,
    call_id: int | None = None,
    limit: int = 500,
) -> list[dict]:
    """Query conversation_transcripts by phone, email, and/or call_id."""
    init_db()
    con = _connect()
    clauses: list[str] = []
    args: list[object] = []
    if phone:
        clauses.append("phone_number = ?")
        args.append(phone.strip())
    if email:
        clauses.append("email = ?")
        args.append(email.strip().lower())
    if call_id is not None:
        clauses.append("call_id = ?")
        args.append(call_id)
    where = f"WHERE {' AND '.join(clauses)}" if clauses else ""
    args.append(max(1, min(int(limit), 5000)))
    rows = con.execute(
        f"""
        SELECT id, call_id, call_sid, channel, phone_number, email,
               role, content, agent_mode, ts
          FROM conversation_transcripts
          {where}
         ORDER BY id DESC
         LIMIT ?
        """,
        args,
    ).fetchall()
    con.close()
    data = [dict(r) for r in rows]
    # Chronological when scoped to one contact/call
    if call_id is not None or phone or email:
        data.reverse()
    return data


def get_call_number(call_id: int) -> str | None:
    """Return the contact phone stored for a given call_id."""
    init_db()
    con = _connect()
    row = con.execute("SELECT to_number FROM calls WHERE id = ?", (call_id,)).fetchone()
    con.close()
    if not row:
        return None
    phone = (row["to_number"] or "").strip()
    if phone.lower() in {"browser", "unknown", ""}:
        return None
    return phone


def is_opted_out(phone: str) -> bool:
    """Return True if the number is in the internal do-not-call list."""
    init_db()
    con = _connect()
    row = con.execute(
        "SELECT 1 FROM opt_outs WHERE phone = ? LIMIT 1", (phone,)
    ).fetchone()
    con.close()
    return row is not None


def record_opt_out(phone: str, call_id: int) -> None:
    """
    Add phone to the internal do-not-call list (idempotent).
    Also marks the call outcome as 'opted_out'.
    """
    init_db()
    con = _connect()
    existing = con.execute(
        "SELECT 1 FROM opt_outs WHERE phone = ? LIMIT 1", (phone,)
    ).fetchone()
    if not existing:
        con.execute(
            "INSERT INTO opt_outs (phone, call_id) VALUES (?, ?)",
            (phone, call_id),
        )
    con.execute(
        "UPDATE calls SET outcome = 'opted_out' WHERE id = ? AND outcome NOT IN ('opted_out')",
        (call_id,),
    )
    con.commit()
    con.close()
