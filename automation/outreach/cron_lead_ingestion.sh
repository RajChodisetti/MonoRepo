#!/usr/bin/env bash
# Daily lead ingestion cron — fetch, scrape, import with 500-request budget.
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
