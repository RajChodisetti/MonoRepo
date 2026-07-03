#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RUN_DIR="$ROOT/.run"
COMPOSE_FILE="${COMPOSE_FILE:-infra/docker/docker-compose.yml}"
STOP_DB="${STOP_DB:-1}"

stop_pid() {
  local name="$1"
  local pid_file="$RUN_DIR/$name.pid"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file")"
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
      echo "Stopped $name (pid $pid)"
    fi
    rm -f "$pid_file"
  fi
}

stop_pid api
stop_pid worker

if [[ "$STOP_DB" == "1" ]] && command -v docker >/dev/null 2>&1; then
  docker compose -f "$COMPOSE_FILE" down
  echo "Stopped Docker services (postgres)."
else
  echo "Left Postgres running (STOP_DB=0)."
fi

echo "Done."
