#!/usr/bin/env bash
# VM / production: build and start postgres + migrate + api + worker in Docker.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/infra/docker"

STACK_ENV="stack.env"

if [[ -f "$ROOT/backend/.env" ]]; then
  cp "$ROOT/backend/.env" "$STACK_ENV"
  echo "Using backend/.env as Docker stack env"
elif [[ ! -f "$STACK_ENV" ]]; then
  cp stack.env.example "$STACK_ENV"
  echo "Created $STACK_ENV from stack.env.example — set SMTP/secrets before sending email"
fi

docker compose --profile stack up -d --build --wait

cat <<EOF

Stack is up.

  API:      http://localhost:8080
  Postgres: localhost:5432

  Logs:     docker compose --profile stack logs -f
  Stop:     docker compose --profile stack down

On a VM, set PUBLIC_BASE_URL and CORS_ALLOWED_ORIGINS to your server URL in stack.env.

EOF
