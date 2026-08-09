#!/usr/bin/env bash
# Keep template, website, API, worker, and voice agent running (macOS local dev).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT/.run"
COMPOSE_FILE="$ROOT/infra/docker/docker-compose.yml"
mkdir -p "$RUN_DIR/bin"

kill_port() {
  local port="$1"
  lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | xargs kill -9 2>/dev/null || true
}

ensure_api() {
  if lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then return 0; fi
  echo "[watchdog] starting API :8080"
  (cd "$ROOT" && go build -o "$RUN_DIR/bin/api" ./backend/cmd/api)
  nohup "$RUN_DIR/bin/api" >>"$RUN_DIR/api.log" 2>&1 &
  echo $! >"$RUN_DIR/api.pid"
}

ensure_worker() {
  if [[ -f "$RUN_DIR/worker.pid" ]] && kill -0 "$(cat "$RUN_DIR/worker.pid")" 2>/dev/null; then return 0; fi
  echo "[watchdog] starting worker"
  nohup go run "$ROOT/backend/cmd/worker" >>"$RUN_DIR/worker.log" 2>&1 &
  echo $! >"$RUN_DIR/worker.pid"
}

ensure_template() {
  if lsof -nP -iTCP:3000 -sTCP:LISTEN >/dev/null 2>&1; then return 0; fi
  echo "[watchdog] starting template :3000"
  nohup npm --prefix "$ROOT/template" run dev -- -p 3000 >>"$RUN_DIR/template.log" 2>&1 &
  echo $! >"$RUN_DIR/template.pid"
}

ensure_website() {
  if lsof -nP -iTCP:3001 -sTCP:LISTEN >/dev/null 2>&1; then return 0; fi
  echo "[watchdog] starting website :3001"
  nohup npm --prefix "$ROOT/web" run dev -- -p 3001 >>"$RUN_DIR/website.log" 2>&1 &
  echo $! >"$RUN_DIR/website.pid"
}

ensure_voice() {
  docker compose -f "$COMPOSE_FILE" --profile voice up -d voice-sales-agent voice-sales-redis >/dev/null 2>&1 || true
}

docker compose -f "$COMPOSE_FILE" up -d postgres --wait >/dev/null 2>&1 || true

echo "[watchdog] running — template :3000, website :3001, API :8080, voice :8000"

while true; do
  ensure_api
  ensure_worker
  ensure_template
  ensure_website
  ensure_voice
  sleep 15
done
