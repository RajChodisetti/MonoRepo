#!/usr/bin/env bash

# Structural checks for the repository's AI instruction hierarchy and router.
# Compatible with the Bash 3.2 shipped by macOS.

set -euo pipefail

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
if ! REPO_ROOT=$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null); then
  printf 'error: scripts/check-agent-context.sh must live inside a Git repository\n' >&2
  exit 1
fi

ROUTER="$REPO_ROOT/scripts/agent-context.sh"
ROOT_MAX_BYTES=$((16 * 1024))
SCOPED_MAX_BYTES=$((12 * 1024))
COMBINED_MAX_BYTES=$((32 * 1024))
FAILURES=0

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  FAILURES=$((FAILURES + 1))
}

pass() {
  printf 'PASS: %s\n' "$1"
}

warn() {
  printf 'WARN: %s\n' "$1" >&2
}

file_bytes() {
  wc -c < "$1" | tr -d '[:space:]'
}

check_claude_import() {
  claude_file=$1
  if [ ! -f "$claude_file" ]; then
    fail "missing ${claude_file#"$REPO_ROOT"/}"
    return 0
  fi

  import_count=$(grep -E -c '^[[:space:]]*@AGENTS\.md[[:space:]]*$' "$claude_file" 2>/dev/null || true)
  at_line_count=$(grep -E -c '^[[:space:]]*@' "$claude_file" 2>/dev/null || true)
  if [ "$import_count" -ne 1 ]; then
    fail "${claude_file#"$REPO_ROOT"/} must contain exactly one relative @AGENTS.md import"
  elif [ "$at_line_count" -ne 1 ]; then
    fail "${claude_file#"$REPO_ROOT"/} contains an unexpected additional @ import"
  else
    pass "${claude_file#"$REPO_ROOT"/} imports its sibling AGENTS.md"
  fi
}

require_fixed() {
  expected=$1
  target=$2
  description=$3
  if grep -Fq -- "$expected" "$target"; then
    pass "$description"
  else
    fail "$description (missing fixed string: $expected)"
  fi
}

if [ ! -f "$ROUTER" ]; then
  fail "missing scripts/agent-context.sh"
else
  if bash -n "$ROUTER"; then
    pass "scripts/agent-context.sh has valid Bash syntax"
  else
    fail "scripts/agent-context.sh has invalid Bash syntax"
  fi
fi

if bash -n "$REPO_ROOT/scripts/check-agent-context.sh"; then
  pass "scripts/check-agent-context.sh has valid Bash syntax"
else
  fail "scripts/check-agent-context.sh has invalid Bash syntax"
fi

if [ ! -f "$REPO_ROOT/AGENTS.md" ]; then
  fail "missing AGENTS.md"
  ROOT_BYTES=0
else
  ROOT_BYTES=$(file_bytes "$REPO_ROOT/AGENTS.md")
  if [ "$ROOT_BYTES" -le "$ROOT_MAX_BYTES" ]; then
    pass "AGENTS.md is ${ROOT_BYTES} bytes (limit ${ROOT_MAX_BYTES})"
  else
    fail "AGENTS.md is ${ROOT_BYTES} bytes (limit ${ROOT_MAX_BYTES})"
  fi
fi

check_claude_import "$REPO_ROOT/CLAUDE.md"

TMP_BASE=${TMPDIR:-/tmp}
SCOPE_FILE=$(mktemp "$TMP_BASE/tuvi-agent-scopes.XXXXXX")
ROUTE_OUTPUT=$(mktemp "$TMP_BASE/tuvi-agent-route.XXXXXX")
DELETION_TEST_ROOT=$(mktemp -d "$TMP_BASE/tuvi-agent-deletion.XXXXXX")
trap 'rm -f "$SCOPE_FILE" "$ROUTE_OUTPUT"; rm -rf "$DELETION_TEST_ROOT"' EXIT HUP INT TERM

if [ -f "$ROUTER" ]; then
  if bash "$ROUTER" --list-scopes > "$SCOPE_FILE"; then
    pass "agent context scope registry is readable"
  else
    fail "agent context scope registry is not readable"
    : > "$SCOPE_FILE"
  fi
else
  : > "$SCOPE_FILE"
fi

WORST_COMBINED=0
WORST_SCOPE=
while IFS= read -r scope; do
  [ -z "$scope" ] && continue
  scoped_agents="$REPO_ROOT/$scope/AGENTS.md"
  scoped_claude="$REPO_ROOT/$scope/CLAUDE.md"

  if [ ! -f "$scoped_agents" ]; then
    fail "missing $scope/AGENTS.md"
    scoped_bytes=0
  else
    scoped_bytes=$(file_bytes "$scoped_agents")
    if [ "$scoped_bytes" -le "$SCOPED_MAX_BYTES" ]; then
      pass "$scope/AGENTS.md is ${scoped_bytes} bytes (limit ${SCOPED_MAX_BYTES})"
    else
      fail "$scope/AGENTS.md is ${scoped_bytes} bytes (limit ${SCOPED_MAX_BYTES})"
    fi
  fi

  combined_bytes=$((ROOT_BYTES + scoped_bytes))
  if [ "$combined_bytes" -le "$COMBINED_MAX_BYTES" ]; then
    pass "AGENTS.md + $scope/AGENTS.md is ${combined_bytes} bytes (limit ${COMBINED_MAX_BYTES})"
  else
    fail "AGENTS.md + $scope/AGENTS.md is ${combined_bytes} bytes (limit ${COMBINED_MAX_BYTES})"
  fi
  if [ "$combined_bytes" -gt "$WORST_COMBINED" ]; then
    WORST_COMBINED=$combined_bytes
    WORST_SCOPE=$scope
  fi

  check_claude_import "$scoped_claude"

  if [ -f "$ROUTER" ]; then
    if bash "$ROUTER" "$scope/example file" > "$ROUTE_OUTPUT" &&
      grep -Fq -- "  - $scope/AGENTS.md" "$ROUTE_OUTPUT" &&
      grep -Fq -- "  - $scope/example file" "$ROUTE_OUTPUT"; then
      pass "router maps a space-containing path in $scope"
    else
      fail "router does not map $scope paths safely"
    fi
    if bash "$ROUTER" "$scope/AGENTS.md" > "$ROUTE_OUTPUT" &&
      grep -Fq -- "  - $scope/AGENTS.md" "$ROUTE_OUTPUT"; then
      pass "router retains $scope instructions on instruction-file changes"
    else
      fail "router drops $scope instructions on instruction-file changes"
    fi
  fi
done < "$SCOPE_FILE"

if [ -n "$WORST_SCOPE" ]; then
  printf 'INFO: largest root+nested instruction set is %s bytes at %s\n' "$WORST_COMBINED" "$WORST_SCOPE"
fi

if [ -f "$ROUTER" ]; then
  mkdir -p "$DELETION_TEST_ROOT/scripts" "$DELETION_TEST_ROOT/backend"
  cp "$ROUTER" "$DELETION_TEST_ROOT/scripts/agent-context.sh"
  printf 'package deleted\n' > "$DELETION_TEST_ROOT/backend/deleted.go"
  git -C "$DELETION_TEST_ROOT" init -q
  git -C "$DELETION_TEST_ROOT" add scripts/agent-context.sh backend/deleted.go
  git -C "$DELETION_TEST_ROOT" -c user.name='Tuvi Context Check' -c user.email='context-check@invalid' commit -qm 'fixture'
  rm "$DELETION_TEST_ROOT/backend/deleted.go"
  if bash "$DELETION_TEST_ROOT/scripts/agent-context.sh" --changed > "$ROUTE_OUTPUT" &&
    grep -Fq -- '  - backend/deleted.go' "$ROUTE_OUTPUT" &&
    grep -Fq -- '  - backend/AGENTS.md' "$ROUTE_OUTPUT"; then
    pass "router includes deleted paths and their scoped instructions"
  else
    fail "router omitted a deleted path or its scoped instructions"
  fi
fi

tracked_noise=$(git -C "$REPO_ROOT" ls-files -- api '*.tsbuildinfo' 'automation/outreach/logs/*.log' || true)
if [ -n "$tracked_noise" ]; then
  warn "tracked generated/runtime artifacts can keep the worktree dirty even though new copies are ignored:"
  printf '%s\n' "$tracked_noise" >&2
fi

if [ ! -f "$REPO_ROOT/.nvmrc" ]; then
  fail "missing .nvmrc"
elif grep -Fxq '22' "$REPO_ROOT/.nvmrc"; then
  pass ".nvmrc pins Node 22"
else
  fail ".nvmrc must pin Node 22"
fi

if [ ! -f "$REPO_ROOT/.python-version" ]; then
  fail "missing .python-version"
elif grep -Fxq '3.12' "$REPO_ROOT/.python-version"; then
  pass ".python-version pins Python 3.12"
else
  fail ".python-version must pin Python 3.12"
fi

if [ -f "$REPO_ROOT/.nvmrc" ] && command -v node >/dev/null 2>&1; then
  expected_node=$(tr -d '[:space:]' < "$REPO_ROOT/.nvmrc")
  expected_node=${expected_node#v}
  expected_node_major=${expected_node%%.*}
  active_node_major=$(node -p 'process.versions.node.split(".")[0]')
  if [ "$active_node_major" != "$expected_node_major" ]; then
    warn "active Node major is $active_node_major but .nvmrc requires $expected_node_major; use Node $expected_node_major for production-equivalent checks"
  else
    pass "active Node major matches .nvmrc ($expected_node_major)"
  fi
fi

if [ -f "$REPO_ROOT/.python-version" ] && command -v python3 >/dev/null 2>&1; then
  expected_python=$(tr -d '[:space:]' < "$REPO_ROOT/.python-version")
  expected_python=${expected_python#python-}
  active_python=$(python3 -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
  if [ "$active_python" != "$expected_python" ]; then
    warn "active Python is $active_python but .python-version requires $expected_python; use Python $expected_python for runtime checks"
  else
    pass "active Python matches .python-version ($expected_python)"
  fi
fi

if command -v go >/dev/null 2>&1 && [ -f "$REPO_ROOT/go.mod" ]; then
  active_go=$(go version | awk '{print $3}')
  module_go=$(awk '$1 == "go" { print $2; exit }' "$REPO_ROOT/go.mod")
  printf 'INFO: active Go is %s; root go.mod requires %s\n' "$active_go" "$module_go"
fi

if [ ! -f "$REPO_ROOT/.github/workflows/agent-context.yml" ]; then
  fail "missing .github/workflows/agent-context.yml"
else
  pass ".github/workflows/agent-context.yml exists"
fi

require_fixed '<!-- BEGIN:nextjs-agent-rules -->' "$REPO_ROOT/apps/web/AGENTS.md" "admin Next.js generated instruction block is preserved"
require_fixed '<!-- BEGIN:nextjs-agent-rules -->' "$REPO_ROOT/apps/andre-admin/AGENTS.md" "Andre admin Next.js generated instruction block is preserved"
require_fixed '<!-- BEGIN:nextjs-agent-rules -->' "$REPO_ROOT/web/AGENTS.md" "corporate Next.js generated instruction block is preserved"

if grep -Fq -- 'tuvi-website/app' "$ROUTER"; then
  fail "agent router contains the retired tuvi-website/app path"
else
  pass "agent router uses the canonical web/ corporate path"
fi

require_fixed 'process.env.TEMPLATE ?? "3"' "$REPO_ROOT/template/src/lib/templateConfig.ts" "template code fallback defaults to template 3"
if grep -Fxq 'TEMPLATE=3' "$REPO_ROOT/.env.example"; then
  pass "root .env.example defaults to TEMPLATE=3"
else
  fail "root .env.example must contain an exact TEMPLATE=3 line"
fi
if grep -Fxq 'ARG TEMPLATE=3' "$REPO_ROOT/infra/docker/Dockerfile.template"; then
  pass "template Docker build defaults to template 3"
else
  fail "infra/docker/Dockerfile.template must contain an exact ARG TEMPLATE=3 line"
fi
compose_default_count=$(grep -F -c 'TEMPLATE: ${TEMPLATE:-3}' "$REPO_ROOT/infra/docker/docker-compose.vm.yml" 2>/dev/null || true)
if [ "$compose_default_count" -eq 2 ]; then
  pass "VM template build and runtime both default to template 3"
else
  fail "infra/docker/docker-compose.vm.yml must contain exactly two TEMPLATE: \${TEMPLATE:-3} defaults; found $compose_default_count"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '\nAgent context checks failed: %s problem(s).\n' "$FAILURES" >&2
  exit 1
fi

printf '\nAgent context checks passed.\n'
