#!/bin/bash
# Starts ALL Tuvi services: Postgres, API (:8080), worker, voice agent (:8000), website (:3001).
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"
[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh"

cd "$(dirname "$0")" || exit 1

# Docker Desktop must be running for Postgres + voice agent
if ! docker info >/dev/null 2>&1; then
  echo "Docker isn't running — starting Docker Desktop (this can take ~30s)..."
  open -a Docker 2>/dev/null || { echo "ERROR: Docker Desktop not found. Please install/open it, then re-run."; exit 1; }
  for i in $(seq 1 45); do
    docker info >/dev/null 2>&1 && break
    sleep 2
  done
  docker info >/dev/null 2>&1 || { echo "ERROR: Docker did not become ready. Open Docker Desktop and re-run."; exit 1; }
  echo "Docker is ready."
fi

./scripts/restart-tuvi.sh

echo
echo "================================================"
echo "  Website:     http://localhost:3001"
echo "  Restaurants: http://localhost:3001/services/restaurants"
echo "  API:         http://localhost:8080"
echo "  Voice agent: http://localhost:8000"
echo "================================================"
echo "You can close this window; services keep running."
