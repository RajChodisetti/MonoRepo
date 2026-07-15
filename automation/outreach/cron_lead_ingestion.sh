#!/usr/bin/env bash
# RETIRED daily Google Places ingestion cron.
# After migration 000015, daily_ingestion.py refuses to make provider calls;
# use the durable city-scrape API and scrape-worker instead.
#
# Install (daily 02:00 local time):
#   crontab -e
#   0 2 * * * LEAD_INGESTION_ENABLED=true /path/to/MonoRepo/automation/outreach/cron_lead_ingestion.sh
#
# Or run once:
#   LEAD_INGESTION_ENABLED=true ./cron_lead_ingestion.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

LOG_DIR="${ROOT}/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="${LOG_DIR}/lead_ingestion_$(date +%Y%m%d).log"

if [[ -x "${ROOT}/.venv/bin/python" ]]; then
  PYTHON="${ROOT}/.venv/bin/python"
elif command -v python3 >/dev/null 2>&1; then
  PYTHON="python3"
else
  PYTHON="python"
fi

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

# An explicit host/secret file wins over local development defaults.
if [[ -n "${INGESTION_ENV_FILE:-}" ]]; then
  if [[ ! -f "${INGESTION_ENV_FILE}" ]]; then
    echo "INGESTION_ENV_FILE does not exist" >&2
    exit 1
  fi
  set -a
  # shellcheck disable=SC1090
  source "${INGESTION_ENV_FILE}"
  set +a
fi

{
  echo "===== $(date -Iseconds) Lead ingestion cron start ====="
  if [[ "${LEAD_INGESTION_ENABLED:-false}" != "true" ]]; then
    echo "LEAD_INGESTION_ENABLED is not true — skipping"
  else
    export INGESTION_TYPE="${INGESTION_TYPE:-restaurant}"
    export LEAD_INGESTION_MAX_REQUESTS="${LEAD_INGESTION_MAX_REQUESTS:-500}"
    "$PYTHON" daily_ingestion.py
  fi
  echo "===== $(date -Iseconds) Lead ingestion cron done ====="
} >>"$LOG_FILE" 2>&1
