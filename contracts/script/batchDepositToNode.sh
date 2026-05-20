#!/usr/bin/env bash
# batchDepositToNode.sh — approve + depositToNode across multiple chains in parallel.
# See batchDepositToNode.md for config format and usage examples.
set -uo pipefail

# ── Usage ─────────────────────────────────────────────────────────────────────

usage() {
  echo "Usage: $0 <config.json> [--dry-run] [--account <name> | -i] [-- <forge-args>...]"
  echo ""
  echo "  config.json      Path to batch config (see batchDepositToNode.md)"
  echo "  --dry-run        Simulate without broadcasting; skips confirmation wait"
  echo "  --account <name> Keystore account — prompts for password once, reused across all invocations"
  echo "  -i               Interactive private key — prompts once, creates temp keystore"
  echo "  --               Forward remaining args to every forge script invocation"
  echo "                     e.g.: -- --ledger"
  echo "                     e.g.: -- -vvvv"
  exit 1
}

# ── Argument parsing ──────────────────────────────────────────────────────────

CONFIG=""
DRY_RUN=false
ACCOUNT_NAME=""
INTERACTIVE_KEY=false
FORGE_EXTRA=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage ;;
    --dry-run) DRY_RUN=true; shift ;;
    --account)
      [[ -z "${2:-}" ]] && { echo "Error: --account requires a name" >&2; usage; }
      ACCOUNT_NAME="$2"; shift 2 ;;
    -i) INTERACTIVE_KEY=true; shift ;;
    --)
      shift
      FORGE_EXTRA=("$@")
      break
      ;;
    -*)
      echo "Error: unknown flag: $1" >&2
      usage
      ;;
    *)
      if [[ -z "$CONFIG" ]]; then
        CONFIG="$1"
      else
        echo "Error: unexpected argument: $1" >&2
        usage
      fi
      shift
      ;;
  esac
done

[[ -z "$CONFIG"    ]] && { echo "Error: config file required" >&2; usage; }
[[ -f "$CONFIG"    ]] || { echo "Error: config not found: $CONFIG" >&2; exit 1; }
command -v jq  >/dev/null 2>&1 || { echo "Error: jq is required" >&2; exit 1; }
command -v cast >/dev/null 2>&1 || { echo "Error: cast (foundry) is required" >&2; exit 1; }
[[ -f foundry.toml ]] || { echo "Error: must be run from the contracts/ directory" >&2; exit 1; }

if [[ -n "$ACCOUNT_NAME" && "$INTERACTIVE_KEY" == true ]]; then
  echo "Error: --account and -i are mutually exclusive" >&2; exit 1
fi

# ── Signing setup (runs before logging so prompts appear on tty) ──────────────

TMPKS=""
TMPPW=""
SIGNING_ARGS=()

if [[ -n "$ACCOUNT_NAME" ]]; then
  TMPPW=$(mktemp)
  read -rsp "Keystore password for '$ACCOUNT_NAME': " KS_PASS; echo
  printf '%s' "$KS_PASS" > "$TMPPW"
  unset KS_PASS
  SIGNING_ARGS=(--account "$ACCOUNT_NAME" --password-file "$TMPPW")

elif $INTERACTIVE_KEY; then
  TMPKS=$(mktemp -d)
  TMPPW=$(mktemp)
  # Random password for the temp keystore — user never needs to see it
  LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom 2>/dev/null | head -c 32 > "$TMPPW" \
    || openssl rand -base64 24 > "$TMPPW"
  read -rsp "Private key (hex, with or without 0x): " PK; echo
  # Import to temp keystore — key and password briefly in process args for this one call only
  cast wallet import "batch-$$" \
    --keystore-dir "$TMPKS" \
    --private-key "$PK" \
    --unsafe-password "$(cat "$TMPPW")" \
    >/dev/null 2>&1
  unset PK
  SIGNING_ARGS=(--keystore "$TMPKS/batch-$$" --password-file "$TMPPW")
fi

# ── Setup ─────────────────────────────────────────────────────────────────────

LOGFILE="$(pwd)/batchDepositToNode-$(date +%Y%m%d-%H%M%S).log"
RESULTS=$(mktemp -d)
trap 'rm -rf "$RESULTS" "${TMPKS:-}" "${TMPPW:-}"' EXIT

exec > >(tee -a "$LOGFILE") 2>&1

echo "Log:     $LOGFILE"
echo "Config:  $CONFIG"
echo "Dry-run: $DRY_RUN"
echo "Started: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo

if $DRY_RUN; then
  BROADCAST_ARGS=()
  echo "*** DRY RUN — no transactions will be broadcast ***"
  echo
else
  BROADCAST_ARGS=(--broadcast)
fi

# ── Per-chain worker ──────────────────────────────────────────────────────────
# Each chain runs as an independent background subshell.
# Tokens within a chain run sequentially; on-chain confirmation is awaited
# after each token before starting the next.

pids=()
chain_idx=0

while read -r chain; do
  rpc=$(echo "$chain" | jq -r '.rpcUrl')
  hub=$(echo "$chain" | jq -r '.hubAddress')
  idx=$chain_idx
  chain_idx=$((chain_idx + 1))

  (
    ok=0
    fail=0
    fail_tokens=""
    chain_id=$(cast chain-id --rpc-url "$rpc" 2>/dev/null)
    if [[ -z "$chain_id" ]]; then
      echo "!!! FAILED: cannot reach RPC $rpc — skipping chain"
      jq -n --arg rpc "$rpc" --arg chain_id "unknown" \
        --argjson ok 0 --argjson fail 1 \
        --arg fail_tokens "ALL (RPC unreachable)" \
        '{rpc:$rpc, chain_id:$chain_id, ok:$ok, fail:$fail, fail_tokens:$fail_tokens}' \
        > "$RESULTS/$idx.json"
      exit 1
    fi

    while read -r entry; do
      token=$(echo "$entry" | jq -r '.address')
      human_amount=$(echo "$entry" | jq -r '.amount')

      # Fetch token decimals from chain and compute raw on-chain amount
      decimals=$(cast call "$token" "decimals()(uint8)" --rpc-url "$rpc" 2>/dev/null)
      if [[ -z "$decimals" ]]; then
        echo "!!! FAILED: [chain=$chain_id] could not fetch decimals() for $token"
        fail=$((fail + 1))
        fail_tokens="$fail_tokens $token"
        continue
      fi
      raw_amount=$(awk "BEGIN { printf \"%.0f\", $human_amount * (10 ^ $decimals) }")

      echo ">>> [chain=$chain_id] token=$token amount=$human_amount (decimals=$decimals raw=$raw_amount)"

      if forge script script/DepositToNode.s.sol \
          --sig "run(address,address,uint256)" "$hub" "$token" "$raw_amount" \
          --rpc-url "$rpc" \
          ${BROADCAST_ARGS[@]+"${BROADCAST_ARGS[@]}"} \
          ${SIGNING_ARGS[@]+"${SIGNING_ARGS[@]}"} \
          ${FORGE_EXTRA[@]+"${FORGE_EXTRA[@]}"}; then

        if ! $DRY_RUN; then
          broadcast="broadcast/DepositToNode.s.sol/$chain_id/run-latest.json"
          if [[ -f "$broadcast" ]]; then
            while read -r hash; do
              echo "    awaiting $hash ..."
              cast receipt --confirmations 1 "$hash" --rpc-url "$rpc" >/dev/null \
                || echo "    WARNING: could not confirm $hash — tx may still land"
            done < <(jq -r '.transactions[].hash' "$broadcast")
          fi
        fi

        ok=$((ok + 1))
        echo "    confirmed: [chain=$chain_id] token=$token amount=$human_amount"
      else
        fail=$((fail + 1))
        fail_tokens="$fail_tokens $token"
        echo "!!! FAILED: [chain=$chain_id] token=$token"
      fi
    done < <(echo "$chain" | jq -c '.tokens[]')

    jq -n \
      --arg rpc         "$rpc"        \
      --arg chain_id    "$chain_id"   \
      --argjson ok      "$ok"         \
      --argjson fail    "$fail"       \
      --arg fail_tokens "$fail_tokens" \
      '{rpc:$rpc, chain_id:$chain_id, ok:$ok, fail:$fail, fail_tokens:$fail_tokens}' \
      > "$RESULTS/$idx.json"
  ) &

  pids+=($!)
done < <(jq -c '.[]' "$CONFIG")

# ── Collect results ───────────────────────────────────────────────────────────

for pid in "${pids[@]}"; do
  wait "$pid" || true  # failures reported via result files, not exit codes
done

# ── Summary ───────────────────────────────────────────────────────────────────

echo
echo "=== Summary ==="

total_ok=0
total_chains=0
fail_reports=()

for f in "$RESULTS"/*.json; do
  [[ -f "$f" ]] || continue
  total_chains=$((total_chains + 1))
  ok=$(jq -r '.ok' "$f")
  fail=$(jq -r '.fail' "$f")
  chain_id=$(jq -r '.chain_id' "$f")
  fail_tokens=$(jq -r '.fail_tokens' "$f")
  total_ok=$((total_ok + ok))
  [[ "$fail" -gt 0 ]] && fail_reports+=("chain=$chain_id: $fail_tokens")
done

echo "Deposited $total_ok tokens across $total_chains chains"

if [[ "${#fail_reports[@]}" -gt 0 ]]; then
  echo
  echo "FAILED:"
  printf '  %s\n' "${fail_reports[@]}"
  exit 1
fi
