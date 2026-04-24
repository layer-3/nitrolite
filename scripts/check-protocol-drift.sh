#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/check-protocol-drift.sh [--static|--runtime]

  --static   Run deterministic protocol/SDK/compat drift checks.
  --runtime  Run runtime smoke checks against an ephemeral/local Clearnode.

Runtime smoke starts an isolated local Clearnode with a temporary config by
default. Set CLEARNODE_RUNTIME_SMOKE_EXTERNAL=1 and CLEARNODE_RUNTIME_SMOKE_WS_URL
to run the same lightweight compatibility smoke against an existing node. This is
not a load or stress test.
USAGE
}

run_package() {
  local package_path="$1"
  local command_name="$2"
  local full_path="$ROOT/$package_path"

  if [[ ! -d "$full_path" ]]; then
    echo "::error::drift check package path does not exist: $package_path" >&2
    return 1
  fi

  echo
  echo "==> $package_path: npm run $command_name"
  (
    cd "$full_path"
    npm run "$command_name"
  )
}

mode="${1:---static}"

case "$mode" in
  --static)
    echo "==> Running deterministic Nitrolite protocol drift checks"
    run_package "sdk/ts" "drift:check"
    run_package "sdk/ts-compat" "drift:check"
    ;;
  --runtime)
    echo "==> Running Nitrolite protocol runtime smoke checks"
    run_package "sdk/ts" "build:ci"
    run_package "sdk/ts-compat" "build:ci"
    node "$ROOT/scripts/drift/runtime-smoke.mjs"
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
