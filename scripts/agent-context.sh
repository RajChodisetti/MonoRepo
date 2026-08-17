#!/usr/bin/env bash

# Route changed or explicitly supplied repository paths to the smallest useful
# context set and the checks that protect their known dependency seams.
# Compatible with the Bash 3.2 shipped by macOS.

set -euo pipefail

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
if ! REPO_ROOT=$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null); then
  printf 'error: scripts/agent-context.sh must live inside a Git repository\n' >&2
  exit 1
fi

list_scopes() {
  cat <<'EOF'
backend
automation/outreach
apps/web
apps/andre-admin
template
web
apps/restaurant-services-catalog
voice-sales-agent
andre-voice-agent
ocr-electrical-poc
infra
docs
EOF
}

usage() {
  cat <<'EOF'
Usage:
  bash scripts/agent-context.sh [--] [PATH ...]
  bash scripts/agent-context.sh --changed
  bash scripts/agent-context.sh --list-scopes

With PATH arguments, routes those paths. Without arguments, routes unstaged,
staged, and untracked files in the current Git worktree. Paths may contain
spaces and may be repository-relative, current-directory-relative, or absolute
paths inside this repository.
EOF
}

CHANGED_MODE=0
case "${1-}" in
  -h|--help)
    usage
    exit 0
    ;;
  --list-scopes)
    list_scopes
    exit 0
    ;;
  --changed)
    CHANGED_MODE=1
    shift
    ;;
esac

TMP_BASE=${TMPDIR:-/tmp}
TMP_WORK=$(mktemp -d "$TMP_BASE/tuvi-agent-context.XXXXXX")
trap 'rm -rf "$TMP_WORK"' EXIT HUP INT TERM

PATHS_FILE="$TMP_WORK/paths"
DOCS_FILE="$TMP_WORK/docs"
ADAPTERS_FILE="$TMP_WORK/adapters"
IMPACTS_FILE="$TMP_WORK/impacts"
COMMANDS_FILE="$TMP_WORK/commands"
: > "$PATHS_FILE"
: > "$DOCS_FILE"
: > "$ADAPTERS_FILE"
: > "$IMPACTS_FILE"
: > "$COMMANDS_FILE"

add_unique() {
  destination=$1
  value=$2
  if ! grep -Fqx -- "$value" "$destination" 2>/dev/null; then
    printf '%s\n' "$value" >> "$destination"
  fi
}

add_doc() {
  add_unique "$DOCS_FILE" "$1"
}

add_adapter() {
  add_unique "$ADAPTERS_FILE" "$1"
}

add_scope() {
  scope=$1
  add_doc "$scope/AGENTS.md"
  add_adapter "$scope/CLAUDE.md -> @AGENTS.md"
}

add_impact() {
  add_unique "$IMPACTS_FILE" "$1"
}

add_command() {
  add_unique "$COMMANDS_FILE" "$1"
}

add_backend_checks() {
  add_command "rtk make test"
  add_command "rtk go vet ./backend/..."
  add_command "rtk go build ./backend/cmd/..."
}

add_automation_checks() {
  add_command "rtk automation/outreach/.venv/bin/python -m unittest discover -s automation/outreach -p '*_test.py'"
}

add_admin_checks() {
  add_command "rtk npm --prefix apps/web run lint"
  add_command "rtk npm exec --prefix apps/web -- tsc --noEmit --incremental false --pretty false -p apps/web/tsconfig.json"
  add_command "rtk npm --prefix apps/web run build"
}

add_andre_admin_checks() {
  add_command "rtk npm --prefix apps/andre-admin run lint"
  add_command "rtk npm exec --prefix apps/andre-admin -- tsc --noEmit --incremental false --pretty false -p apps/andre-admin/tsconfig.json"
  add_command "rtk npm --prefix apps/andre-admin run build"
}

add_template_checks() {
  add_command "rtk npm --prefix template run test:unit"
  add_command "rtk npm exec --prefix template -- tsc --noEmit --incremental false --pretty false -p template/tsconfig.json"
  add_command "rtk npm --prefix template run build"
}

add_corporate_checks() {
  add_command "rtk npm --prefix web run lint"
  add_command "rtk npm exec --prefix web -- tsc --noEmit --incremental false --pretty false -p web/tsconfig.json"
  add_command "rtk npm --prefix web run build"
}

add_catalog_checks() {
  add_command "rtk npm --prefix apps/restaurant-services-catalog run build"
}

add_compose_checks() {
  add_command "rtk proxy env DATABASE_URL=postgres://example:example@localhost/example POSTGRES_PASSWORD=example TUVI_API_TOKEN=example CALL_API_SECRET=example docker compose -f infra/docker/docker-compose.yml --profile stack --profile jobs --profile voice config --quiet"
  add_command "rtk proxy env DATABASE_URL=postgres://example:example@localhost/example POSTGRES_PASSWORD=example TUVI_API_TOKEN=example CALL_API_SECRET=example docker compose -f infra/docker/docker-compose.vm.yml config --quiet"
}

normalise_path() {
  input_path=$1

  case "$input_path" in
    "$REPO_ROOT")
      printf '.\n'
      return 0
      ;;
    "$REPO_ROOT"/*)
      relative_path=${input_path#"$REPO_ROOT"/}
      ;;
    /*)
      printf 'error: path is outside this repository: %s\n' "$input_path" >&2
      return 1
      ;;
    *)
      if ! current_root=$(git -C "$PWD" rev-parse --show-toplevel 2>/dev/null); then
        printf 'error: current directory is outside this repository: %s\n' "$PWD" >&2
        return 1
      fi
      if [ "$current_root" != "$REPO_ROOT" ]; then
        printf 'error: current directory belongs to another repository: %s\n' "$PWD" >&2
        return 1
      fi
      current_prefix=$(git -C "$PWD" rev-parse --show-prefix 2>/dev/null)
      relative_path=${current_prefix}${input_path}
      ;;
  esac

  while [ "${relative_path#./}" != "$relative_path" ]; do
    relative_path=${relative_path#./}
  done
  while [ "$relative_path" != "." ] && [ "${relative_path%/}" != "$relative_path" ]; do
    relative_path=${relative_path%/}
  done

  case "/$relative_path/" in
    */../*)
      printf 'error: parent-directory path segments are not supported: %s\n' "$input_path" >&2
      return 1
      ;;
  esac

  if [ -z "$relative_path" ]; then
    relative_path=.
  fi
  printf '%s\n' "$relative_path"
}

collect_changed_paths() {
  git -C "$REPO_ROOT" diff --name-only --diff-filter=ACMRTUXBD |
    while IFS= read -r changed_path; do
      [ -n "$changed_path" ] && add_unique "$PATHS_FILE" "$changed_path"
    done
  git -C "$REPO_ROOT" diff --cached --name-only --diff-filter=ACMRTUXBD |
    while IFS= read -r changed_path; do
      [ -n "$changed_path" ] && add_unique "$PATHS_FILE" "$changed_path"
    done
  git -C "$REPO_ROOT" ls-files --others --exclude-standard |
    while IFS= read -r changed_path; do
      [ -n "$changed_path" ] && add_unique "$PATHS_FILE" "$changed_path"
    done
}

if [ "$CHANGED_MODE" -eq 1 ] && [ "$#" -ne 0 ]; then
  printf 'error: --changed cannot be combined with explicit paths\n' >&2
  exit 2
fi

if [ "${1-}" = "--" ]; then
  shift
fi

if [ "$CHANGED_MODE" -eq 1 ]; then
  collect_changed_paths
elif [ "$#" -gt 0 ]; then
  for supplied_path in "$@"; do
    routed_path=$(normalise_path "$supplied_path")
    add_unique "$PATHS_FILE" "$routed_path"
  done
else
  collect_changed_paths
fi

add_doc "AGENTS.md"
add_doc "docs/SESSION_SUMMARY.md"
add_doc "docs/ai/DEPENDENCY_MAP.md"
if [ -f "$REPO_ROOT/docs/ai/CURRENT_STATE.md" ]; then
  add_doc "docs/ai/CURRENT_STATE.md"
fi
add_adapter "CLAUDE.md -> @AGENTS.md"
add_command "rtk git diff --check"

route_path() {
  path=$1
  quoted_path=$(printf '%q' "$path")

  case "$path" in
    backend|backend/*) add_scope "backend" ;;
    automation/outreach|automation/outreach/*) add_scope "automation/outreach" ;;
    apps/web|apps/web/*) add_scope "apps/web" ;;
    apps/andre-admin|apps/andre-admin/*) add_scope "apps/andre-admin" ;;
    template|template/*) add_scope "template" ;;
    web|web/*) add_scope "web" ;;
    apps/restaurant-services-catalog|apps/restaurant-services-catalog/*) add_scope "apps/restaurant-services-catalog" ;;
    voice-sales-agent|voice-sales-agent/*) add_scope "voice-sales-agent" ;;
    andre-voice-agent|andre-voice-agent/*) add_scope "andre-voice-agent" ;;
    ocr-electrical-poc|ocr-electrical-poc/*) add_scope "ocr-electrical-poc" ;;
    infra|infra/*) add_scope "infra" ;;
    docs|docs/*) add_scope "docs" ;;
  esac

  case "$path" in
    AGENTS.md|CLAUDE.md|*/AGENTS.md|*/CLAUDE.md|scripts/agent-context.sh|scripts/check-agent-context.sh|.github/workflows/agent-context.yml)
      add_impact "AI instruction hierarchy or routing changed; validate imports, byte budgets, and shell syntax."
      add_command "rtk bash scripts/check-agent-context.sh"
      return 0
      ;;
  esac

  case "$path" in
    *.sh)
      add_command "rtk bash -n $quoted_path"
      ;;
  esac

  case "$path" in
    .nvmrc)
      add_scope "apps/web"
      add_scope "apps/andre-admin"
      add_scope "template"
      add_scope "web"
      add_scope "apps/restaurant-services-catalog"
      add_scope "infra"
      add_impact "The pinned Node major applies to the two Next.js 16 apps, the Next.js 15 Andre admin, the restaurant template, and the Vite catalog."
      add_admin_checks
      add_andre_admin_checks
      add_template_checks
      add_corporate_checks
      add_catalog_checks
      ;;
    .python-version)
      add_scope "automation/outreach"
      add_scope "voice-sales-agent"
      add_scope "andre-voice-agent"
      add_scope "ocr-electrical-poc"
      add_scope "infra"
      add_impact "The pinned Python minor must match durable ingestion, voice, and standalone Python prototype runtimes."
      add_automation_checks
      add_command "rtk python3 -m compileall -q voice-sales-agent"
      add_command "rtk python3 -m compileall -q -x '(^|/)(\\.venv|__pycache__)(/|$)' andre-voice-agent"
      add_command "rtk python3 -m compileall -q -x '(^|/)(\\.venv|__pycache__)(/|$)' ocr-electrical-poc"
      ;;
  esac

  case "$path" in
    backend|backend/*|go.mod|go.sum|Makefile|infra/docker/Dockerfile.backend)
      add_scope "backend"
      add_impact "The root Go module backs the API, worker, migration binary, and shared domain contracts."
      add_backend_checks
      ;;
  esac

  case "$path" in
    automation/outreach|automation/outreach/*|infra/docker/Dockerfile.outreach)
      add_scope "automation/outreach"
      add_impact "Durable city ingestion/import uses a Python 3.12 container with an explicit Docker COPY allowlist; image OCR is retired."
      add_automation_checks
      ;;
  esac

  case "$path" in
    automation/outreach/*.py)
      add_command "rtk automation/outreach/.venv/bin/python -m py_compile $quoted_path"
      ;;
  esac

  case "$path" in
    automation/outreach/city_scrape_worker.py|automation/outreach/scrape_job_store.py|automation/outreach/env_loader.py|automation/outreach/apollo_enrichment.py|automation/outreach/google_places_scraper.py|automation/outreach/import_to_db.py|automation/outreach/niche_config.py|automation/outreach/request_budget.py|automation/outreach/scrape_safety.py|automation/outreach/known_leads.py|automation/outreach/ingestion_merge.py|automation/outreach/ingestion_state.py|automation/outreach/requirements.ingestion.txt|infra/docker/Dockerfile.outreach)
      add_command "rtk docker build -f infra/docker/Dockerfile.outreach ."
      ;;
  esac

  case "$path" in
    apps/web/package.json|apps/web/package-lock.json)
      add_command "rtk npm --prefix apps/web ci"
      ;;
  esac

  case "$path" in
    apps/web|apps/web/*|infra/docker/Dockerfile.web)
      add_scope "apps/web"
      add_impact "The admin portal is an independent Node 22 project and consumes main API contracts through its BFF."
      add_admin_checks
      ;;
  esac

  case "$path" in
    template/package.json|template/package-lock.json)
      add_command "rtk npm --prefix template ci"
      ;;
  esac

  case "$path" in
    template|template/*|data|data/*|infra/docker/Dockerfile.template)
      add_scope "template"
      add_impact "Restaurant templates are an independent Node 22 project; data and API payload changes affect all public variants."
      add_template_checks
      ;;
  esac

  case "$path" in
    template/src/lib/templateConfig.ts|.env.example|infra/docker/Dockerfile.template|infra/docker/docker-compose.vm.yml)
      add_scope "infra"
      add_impact "Template default selection is duplicated across code, root env defaults, Docker build args, and VM build/runtime defaults; all must remain template 3."
      add_template_checks
      add_compose_checks
      add_command "rtk bash scripts/check-agent-context.sh"
      ;;
  esac

  case "$path" in
    web/package.json|web/package-lock.json)
      add_command "rtk npm --prefix web ci"
      ;;
  esac

  case "$path" in
    web|web/*|infra/docker/Dockerfile.tuvi-website)
      add_scope "web"
      add_impact "The corporate website is an independent Next.js 16/Node 22 project and proxies search, report, and consultation contracts to the main API."
      add_corporate_checks
      ;;
  esac

  case "$path" in
    apps/restaurant-services-catalog/package.json|apps/restaurant-services-catalog/package-lock.json)
      add_command "rtk npm --prefix apps/restaurant-services-catalog ci"
      ;;
  esac

  case "$path" in
    apps/restaurant-services-catalog|apps/restaurant-services-catalog/*|infra/docker/Dockerfile.catalog|infra/docker/nginx-catalog.conf)
      add_scope "apps/restaurant-services-catalog"
      add_impact "The services catalog is an independent Vite/Node 22 static build with no separate lint or test script."
      add_catalog_checks
      ;;
  esac

  case "$path" in
    voice-sales-agent|voice-sales-agent/*|infra/docker/docker-compose.yml|infra/docker/docker-compose.vm.yml)
      add_scope "voice-sales-agent"
      add_impact "The voice agent is a Python 3.12 container service; the host Python may be unsupported, so run the inbound-only policy suite before relying on host checks."
      add_command "rtk python3 -m unittest discover -s voice-sales-agent/tests -p 'test_*.py'"
      ;;
  esac

  case "$path" in
    voice-sales-agent/*.py)
      add_command "rtk python3 -m py_compile $quoted_path"
      ;;
  esac

  case "$path" in
    voice-sales-agent/*|infra/docker/docker-compose.yml|infra/docker/docker-compose.vm.yml)
      add_command "rtk docker build -f voice-sales-agent/Dockerfile voice-sales-agent"
      ;;
  esac

  case "$path" in
    andre-voice-agent|andre-voice-agent/*)
      add_scope "andre-voice-agent"
      add_scope "apps/andre-admin"
      add_impact "Andre is a legacy real-estate prototype with active outbound Twilio code; repository work must not invoke, deploy, or integrate its dialing paths without separate explicit approval."
      add_command "rtk python3 -m compileall -q -x '(^|/)(\\.venv|__pycache__)(/|$)' andre-voice-agent"
      add_andre_admin_checks
      ;;
  esac

  case "$path" in
    andre-voice-agent/*.py)
      add_command "rtk python3 -m py_compile $quoted_path"
      ;;
    andre-voice-agent/Dockerfile|andre-voice-agent/requirements.txt|andre-voice-agent/docker-compose.yml)
      add_command "rtk docker build -f andre-voice-agent/Dockerfile andre-voice-agent"
      ;;
  esac

  case "$path" in
    apps/andre-admin/package.json|apps/andre-admin/package-lock.json)
      add_command "rtk npm --prefix apps/andre-admin ci"
      ;;
  esac

  case "$path" in
    apps/andre-admin|apps/andre-admin/*)
      add_scope "apps/andre-admin"
      add_scope "andre-voice-agent"
      add_impact "The Andre admin is a local-only Next.js operations UI whose server routes can invoke the legacy agent's property and outbound-call APIs."
      add_andre_admin_checks
      ;;
  esac

  case "$path" in
    ocr-electrical-poc|ocr-electrical-poc/*)
      add_scope "ocr-electrical-poc"
      add_impact "The electrical OCR proof of concept is independent from restaurant media ingestion and can invoke a billable external vision provider."
      add_command "rtk python3 -m compileall -q -x '(^|/)(\\.venv|__pycache__)(/|$)' ocr-electrical-poc"
      ;;
  esac

  case "$path" in
    ocr-electrical-poc/*.py|ocr-electrical-poc/app/*.py|ocr-electrical-poc/scripts/*.py)
      add_command "rtk python3 -m py_compile $quoted_path"
      ;;
  esac

  case "$path" in
    infra|infra/*|voice-sales-agent/docker-compose.yml)
      add_scope "infra"
      add_impact "Deployment definitions compose the Go, Node, Python, database, Redis, and proxy boundaries."
      add_compose_checks
      ;;
  esac

  case "$path" in
    docs/openapi|docs/openapi/*|postman|postman/*|backend/internal/http|backend/internal/http/*)
      add_scope "docs"
      add_impact "HTTP route changes can drift from OpenAPI, Postman, and frontend/voice consumers."
      add_command "rtk make openapi"
      ;;
  esac

  case "$path" in
    backend/internal/http/handlers/admin*|backend/internal/http/handlers/auth*|backend/internal/http/handlers/restaurant*|backend/internal/http/handlers/campaign*|backend/internal/http/handlers/outreach*|backend/internal/http/handlers/scrape*|backend/internal/auth|backend/internal/auth/*|backend/internal/restaurants|backend/internal/restaurants/*|backend/internal/campaigns|backend/internal/campaigns/*|backend/internal/outreach|backend/internal/outreach/*)
      add_scope "apps/web"
      add_impact "Admin-facing API response or authorization changes must remain compatible with apps/web BFF types and screens."
      add_admin_checks
      ;;
  esac

  case "$path" in
    backend/internal/http/handlers/demo*|backend/internal/http/handlers/restaurant_site*|backend/internal/http/handlers/tracking*|backend/internal/demos|backend/internal/demos/*|backend/internal/media|backend/internal/media/*|backend/internal/profiles|backend/internal/profiles/*)
      add_scope "template"
      add_impact "Public demo/site/media API changes must remain compatible with all restaurant template adapters and variants."
      add_template_checks
      ;;
  esac

  case "$path" in
    backend/internal/http/handlers/company_consultation*|backend/internal/consultations|backend/internal/consultations/*|web/src/app/api/consultations|web/src/app/api/consultations/*|voice-sales-agent/tuvi_api_client.py)
      add_scope "web"
      add_scope "voice-sales-agent"
      add_impact "Consultation availability and booking contracts are shared by the Go API, corporate Next.js BFF, and voice agent."
      add_backend_checks
      add_corporate_checks
      ;;
  esac

  case "$path" in
    backend/internal/http/handlers/seo_*|backend/internal/seoreport|backend/internal/seoreport/*|web/src/app/api/restaurants|web/src/app/api/restaurants/*|web/src/app/api/leads|web/src/app/api/leads/*|web/src/app/report|web/src/app/report/*|web/src/lib/report*)
      add_scope "backend"
      add_scope "web"
      add_impact "Corporate restaurant search, SEO report, unlock, and PDF contracts span the Go API and web same-origin routes/presentation helpers."
      add_backend_checks
      add_corporate_checks
      ;;
  esac

  case "$path" in
    backend/internal/restaurants|backend/internal/restaurants/*|backend/internal/outreach|backend/internal/outreach/*|backend/migrations/*outreach*|automation/outreach/import_to_db.py|automation/outreach/import_to_db_outreach_test.py)
      add_scope "backend"
      add_scope "automation/outreach"
      add_impact "Restaurant lifecycle, inferred-business consent, and sequence-enrollment predicates span Python import, Go policy, and SQL functions."
      add_backend_checks
      add_automation_checks
      ;;
  esac

  case "$path" in
    backend/internal/campaigns|backend/internal/campaigns/*|backend/internal/outreach|backend/internal/outreach/*|backend/migrations/*outreach*|apps/web/src/app/*outreach*|apps/web/src/lib/*outreach*)
      add_doc "docs/adr/2026-08-12-database-owned-outreach-unsubscribe-content.md"
      add_impact "The approved database sequence solely owns unsubscribe content; application unsubscribe tags/routes and the legacy suppression table are intentionally retired."
      ;;
  esac

  case "$path" in
    template/src/lib/templateConfig.ts|backend/internal/campaigns|backend/internal/campaigns/*|backend/internal/http/handlers/demo_engagement.go|backend/internal/http/handlers/restaurant_site_admin.go|backend/migrations/*template*|apps/web/src/lib/types.ts|apps/web/src/app/*restaurants*)
      add_scope "backend"
      add_scope "apps/web"
      add_scope "template"
      add_impact "Template IDs, names, ordering, tracking validation, database constraints, and admin controls form one cross-system contract."
      add_backend_checks
      add_admin_checks
      add_template_checks
      ;;
  esac
}

while IFS= read -r selected_path; do
  [ -n "$selected_path" ] && route_path "$selected_path"
done < "$PATHS_FILE"

if [ ! -s "$IMPACTS_FILE" ]; then
  add_impact "No known cross-system seam matched; use the root instructions and inspect the selected paths directly."
fi

print_file_list() {
  source_file=$1
  if [ ! -s "$source_file" ]; then
    printf '  - (none)\n'
    return 0
  fi
  while IFS= read -r line; do
    printf '  - %s\n' "$line"
  done < "$source_file"
}

printf 'Repository: %s\n\n' "$REPO_ROOT"
printf 'Selected paths:\n'
print_file_list "$PATHS_FILE"
printf '\nApplicable instructions and baseline context (read in this order):\n'
print_file_list "$DOCS_FILE"
printf '\nClaude compatibility adapters:\n'
print_file_list "$ADAPTERS_FILE"
printf '\nCross-system impacts to inspect:\n'
print_file_list "$IMPACTS_FILE"
printf '\nExact verification commands (run from repository root):\n'
print_file_list "$COMMANDS_FILE"
