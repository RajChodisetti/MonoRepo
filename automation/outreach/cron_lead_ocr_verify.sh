#!/usr/bin/env bash
# Daily lead OCR verification cron.
#
# Install (daily 03:00 local time):
#   crontab -e
#   0 3 * * * /path/to/MonoRepo/automation/outreach/cron_lead_ocr_verify.sh
#
# Or run once:
#   LEAD_OCR_VERIFICATION_ENABLED=true ./cron_lead_ocr_verify.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

LOG_DIR="${ROOT}/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="${LOG_DIR}/lead_ocr_verify_$(date +%Y%m%d).log"

if [[ -x "${ROOT}/.venv/bin/python" ]]; then
  PYTHON="${ROOT}/.venv/bin/python"
elif command -v python3 >/dev/null 2>&1; then
  PYTHON="python3"
else
  PYTHON="python"
fi

# Load outreach .env only; verify_leads_from_db.py loads backend/.env via env_loader
if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${ROOT}/.env"
  set +a
fi

{
  echo "===== $(date -Iseconds) Lead OCR verify cron start ====="
  if [[ "${LEAD_OCR_VERIFICATION_ENABLED:-false}" != "true" ]]; then
    echo "LEAD_OCR_VERIFICATION_ENABLED is not true — skipping"
  else
    "$PYTHON" verify_leads_from_db.py --force
  fi
  echo "===== $(date -Iseconds) Lead OCR verify cron done ====="
} >>"$LOG_FILE" 2>&1
