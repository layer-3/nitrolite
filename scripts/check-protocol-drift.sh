#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/check-protocol-drift.sh [--static|--runtime]

  --static   Run deterministic protocol/SDK/compat drift checks.
  --runtime  Run runtime smoke checks against an ephemeral/local Clearnode.

Runtime smoke is intentionally not implemented in the static runner yet; CI will
wire it once the ephemeral Clearnode harness lands.
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
    echo "::error::runtime protocol drift smoke is not implemented yet; use --static for deterministic checks" >&2
    exit 2
    ;;
  -h|--help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
