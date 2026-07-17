#!/usr/bin/env bash
# Trigger a production scrape job. Requires internal_admin Bearer token.
#
# Usage:
#   BEARER_TOKEN='eyJ...' ./scripts/trigger-scrape.sh
#   BEARER_TOKEN='eyJ...' CITY=Sydney NICHE=dentist ./scripts/trigger-scrape.sh
#
# Or paste token into Postman collection variable `accessToken` and skip Login.

set -euo pipefail

BASE_URL="${TUVI_API:-https://api.tuvisolutions.com}"
TOKEN="${BEARER_TOKEN:-${ADMIN_TOKEN:-}}"
CITY="${CITY:-Melbourne}"
NICHE="${NICHE:-restaurant}"

if [[ -z "$TOKEN" ]]; then
  echo "Set BEARER_TOKEN (internal_admin JWT from POST /api/v1/auth/login)." >&2
  exit 1
fi

auth=(-H "Authorization: Bearer ${TOKEN}")

echo "== Admin me =="
curl -fsS "${auth[@]}" "${BASE_URL}/api/v1/admin/me"
echo

echo "== Trigger scrape: city=${CITY} niche=${NICHE} =="
RESP=$(curl -fsS -X POST "${auth[@]}" \
  -H 'Content-Type: application/json' \
  --data "{\"city\":\"${CITY}\",\"niche\":\"${NICHE}\"}" \
  "${BASE_URL}/api/v1/scrape-jobs")
echo "$RESP"

JOB_ID=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); j=d.get('job') or d; print(j.get('id',''))" 2>/dev/null || true)
if [[ -z "$JOB_ID" ]]; then
  echo "Could not parse job id from response." >&2
  exit 0
fi

echo
echo "== Poll job ${JOB_ID} (5x, 3s apart) =="
for i in 1 2 3 4 5; do
  sleep 3
  echo "--- poll $i ---"
  curl -fsS "${auth[@]}" "${BASE_URL}/api/v1/scrape-jobs/${JOB_ID}" || true
  echo
done
