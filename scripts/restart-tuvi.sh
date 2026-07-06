#!/usr/bin/env bash
# Stop and start every service the Tuvi corporate website needs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
RUN_DIR="$ROOT/.run"
COMPOSE_FILE="${COMPOSE_FILE:-infra/docker/docker-compose.yml}"
mkdir -p "$RUN_DIR" voice-sales-agent/data

stop_pid() {
  local name="$1"
  local pid_file="$RUN_DIR/$name.pid"
  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file")"
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      sleep 0.4
      kill -9 "$pid" 2>/dev/null || true
      echo "Stopped $name (pid $pid)"
    fi
    rm -f "$pid_file"
  fi
}

kill_port() {
  local port="$1"
  local pids
  pids="$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    echo "Killing listeners on :$port -> $pids"
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
    sleep 0.4
    # shellcheck disable=SC2086
    kill -9 $pids 2>/dev/null || true
  fi
}

echo "==> Stopping Tuvi-related services"
stop_pid api
stop_pid worker
stop_pid website
pkill -f 'backend/cmd/api' 2>/dev/null || true
pkill -f 'backend/cmd/worker' 2>/dev/null || true
pkill -f 'next dev -p 3001' 2>/dev/null || true
kill_port 3001
kill_port 8080

docker compose -f "$COMPOSE_FILE" --profile voice stop voice-sales-agent voice-sales-redis || true
docker compose -f "$COMPOSE_FILE" --profile voice rm -f voice-sales-agent voice-sales-redis || true

echo "==> Starting Postgres"
docker compose -f "$COMPOSE_FILE" up -d postgres --wait

echo "==> Migrations"
go run ./backend/cmd/migrate up

if [[ -f tuvi-website/app/.env.local ]] && grep -q '^CONSULTATION_API_URL=http://localhost:8090' tuvi-website/app/.env.local; then
  sed -i '' 's|^CONSULTATION_API_URL=http://localhost:8090|CONSULTATION_API_URL=http://localhost:8080|' tuvi-website/app/.env.local
  echo "Updated CONSULTATION_API_URL -> http://localhost:8080"
fi

echo "==> Building API"
mkdir -p "$RUN_DIR/bin"
go build -o "$RUN_DIR/bin/api" ./backend/cmd/api

echo "==> Starting API :8080"
# Run the binary directly (not `go run`) so SMTP/network works reliably.
nohup "$RUN_DIR/bin/api" >"$RUN_DIR/api.log" 2>&1 &
echo $! >"$RUN_DIR/api.pid"

echo "==> Starting worker"
nohup go run ./backend/cmd/worker >"$RUN_DIR/worker.log" 2>&1 &
echo $! >"$RUN_DIR/worker.pid"

echo "==> Starting voice agent :8000"
docker compose -f "$COMPOSE_FILE" --profile voice up -d --force-recreate voice-sales-agent voice-sales-redis

echo "==> Starting Tuvi website :3001"
nohup npm --prefix tuvi-website/app run dev >"$RUN_DIR/website.log" 2>&1 &
echo $! >"$RUN_DIR/website.pid"

echo "==> Waiting for health"
for i in $(seq 1 30); do
  voice_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:8000/readyz/browser 2>/dev/null || echo 000)
  web_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:3001/ 2>/dev/null || echo 000)
  api_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:8080/healthz 2>/dev/null || echo 000)
  if [[ "$api_ok" != "200" ]]; then
    api_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:8080/health 2>/dev/null || echo 000)
  fi
  # API may answer only on API routes; treat any non-000+refused as alive once port open.
  if lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then
    api_listen=1
  else
    api_listen=0
  fi
  echo "try $i: api_http=$api_ok api_listen=$api_listen voice=$voice_ok web=$web_ok"
  if [[ "$api_listen" == "1" && "$voice_ok" == "200" && ("$web_ok" == "200" || "$web_ok" == "304") ]]; then
    echo "All services are up."
    break
  fi
  sleep 2
done

echo
echo "Website:  http://localhost:3001"
echo "API:      http://localhost:8080"
echo "Voice:    http://localhost:8000"
echo "Logs:     $RUN_DIR/api.log"
echo "          $RUN_DIR/worker.log"
echo "          $RUN_DIR/website.log"
lsof -nP -iTCP:3001,8000,8080 -sTCP:LISTEN || true
