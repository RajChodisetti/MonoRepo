#!/usr/bin/env bash
# Start PostgreSQL, run migrations, API, and worker with one command (local dev).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RUN_DIR="$ROOT/.run"
mkdir -p "$RUN_DIR"

GO="${GO:-go}"
COMPOSE_FILE="${COMPOSE_FILE:-infra/docker/docker-compose.yml}"

log() { printf '==> %s\n' "$*"; }

stop_pid() {
  local name="$1"
  local pid_file="$RUN_DIR/$name.pid"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file")"
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
    rm -f "$pid_file"
  fi
}

cleanup() {
  log "Stopping API and worker..."
  stop_pid api
  stop_pid worker
}

trap cleanup EXIT INT TERM

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required. Install Docker Desktop or use: make up" >&2
  exit 1
fi

log "Starting PostgreSQL..."
docker compose -f "$COMPOSE_FILE" up -d postgres --wait

log "Running migrations..."
"$GO" run ./backend/cmd/migrate up

stop_pid api
stop_pid worker

log "Starting API (http://localhost:8080)..."
nohup "$GO" run ./backend/cmd/api >"$RUN_DIR/api.log" 2>&1 &
echo $! >"$RUN_DIR/api.pid"

log "Starting worker (email jobs)..."
nohup "$GO" run ./backend/cmd/worker >"$RUN_DIR/worker.log" 2>&1 &
echo $! >"$RUN_DIR/worker.pid"

sleep 1

if ! kill -0 "$(cat "$RUN_DIR/api.pid")" 2>/dev/null; then
  echo "API failed to start. Log:" >&2
  tail -n 30 "$RUN_DIR/api.log" >&2 || true
  exit 1
fi

if ! kill -0 "$(cat "$RUN_DIR/worker.pid")" 2>/dev/null; then
  echo "Worker failed to start. Log:" >&2
  tail -n 30 "$RUN_DIR/worker.log" >&2 || true
  exit 1
fi

cat <<EOF

All services are running.

  API:      http://localhost:8080
  Postgres: localhost:5432
  Logs:     $RUN_DIR/api.log
            $RUN_DIR/worker.log

Press Ctrl+C to stop API and worker (Postgres stays up).
To stop everything including DB: make stop-all

EOF

wait "$(cat "$RUN_DIR/api.pid")" "$(cat "$RUN_DIR/worker.pid")"
