#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUN_DIR="$ROOT/.run"
COMPOSE_FILE="$ROOT/infra/docker/docker-compose.yml"
mkdir -p "$RUN_DIR/bin"

kill_port() {
  local port="$1"
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    lsof -nP -tiTCP:"$port" -sTCP:LISTEN | xargs kill -9 2>/dev/null || true
  fi
}

echo "==> Stopping Tuvi website services"
pkill -f "$ROOT/web" 2>/dev/null || true
pkill -f 'next dev -p 3001' 2>/dev/null || true
pkill -f 'next-server' 2>/dev/null || true
kill_port 3001

echo "==> Stopping leftover template / API ports we will restart"
pkill -f 'tuvi-restaurant-template' 2>/dev/null || true
pkill -f 'next dev -p 3000' 2>/dev/null || true
kill_port 3000

# Keep / restart platform services template needs (API + voice agent)
if ! lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then
  echo "==> Building & starting API :8080"
  (cd "$ROOT" && go build -o "$RUN_DIR/bin/api" ./backend/cmd/api)
  nohup "$RUN_DIR/bin/api" >"$RUN_DIR/api.log" 2>&1 &
  echo $! >"$RUN_DIR/api.pid"
fi

echo "==> Starting voice agent :8000"
docker compose -f "$COMPOSE_FILE" --profile voice up -d voice-sales-agent voice-sales-redis

echo "==> Installing template deps if needed"
if [[ ! -d "$ROOT/template/node_modules" ]]; then
  npm --prefix "$ROOT/template" install
fi

echo "==> Starting restaurant template :3000"
nohup npm --prefix "$ROOT/template" run dev -- -p 3000 >"$RUN_DIR/template.log" 2>&1 &
echo $! >"$RUN_DIR/template.pid"

echo "==> Waiting for health"
for i in $(seq 1 40); do
  api_listen=0
  lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1 && api_listen=1
  voice_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:8000/readyz/browser 2>/dev/null || echo 000)
  web_ok=$(curl -sS -m 2 -o /dev/null -w '%{http_code}' http://localhost:3000/ 2>/dev/null || echo 000)
  echo "try $i: api_listen=$api_listen voice=$voice_ok template=$web_ok"
  if [[ "$api_listen" == "1" && "$voice_ok" == "200" && ("$web_ok" == "200" || "$web_ok" == "304") ]]; then
    echo "Template stack is up."
    break
  fi
  sleep 2
done

echo
echo "Template: http://localhost:3000"
echo "API:      http://localhost:8080"
echo "Voice:    http://localhost:8000"
lsof -nP -iTCP:3000,8000,8080 -sTCP:LISTEN || true
