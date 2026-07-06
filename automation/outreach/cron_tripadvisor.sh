#!/usr/bin/env bash
# System crontab entry for TripAdvisor menu-photo scrape.
#
# Install (daily 02:15 local time):
#   crontab -e
#   15 2 * * * /path/to/MonoRepo/automation/outreach/cron_tripadvisor.sh
#
# Or run once:
#   ./cron_tripadvisor.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

LOG_DIR="${ROOT}/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="${LOG_DIR}/tripadvisor_cron_$(date +%Y%m%d).log"

# Prefer project venv if present
if [[ -x "${ROOT}/.venv/bin/python" ]]; then
  PYTHON="${ROOT}/.venv/bin/python"
elif command -v python3 >/dev/null 2>&1; then
  PYTHON="python3"
else
  PYTHON="python"
fi

{
  echo "===== $(date -Iseconds) TripAdvisor cron start ====="
  # Same geo defaults as Google / Places: Sydney Melbourne Perth Adelaide Brisbane
  # --merge writes images.menu_photos from TripAdvisor only into restaurants_data_<city>.json
  "$PYTHON" scrape_tripadvisor.py --all-cities --limit 100 --merge
  echo "===== $(date -Iseconds) TripAdvisor cron done ====="
} >>"$LOG_FILE" 2>&1
