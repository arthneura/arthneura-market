#!/bin/bash
set -euo pipefail
BIN="${BIN:-$HOME/.arthneura/bin/decide-me}"
POINTER="$HOME/.arthneura/owner.dir"
[ -x "$BIN" ] || exit 0
[ -f "$POINTER" ] || exit 0
DIR="$(tr -d '[:space:]' < "$POINTER")"
[ -n "$DIR" ] || exit 0
[ -f "$DIR/rules.json" ] || exit 0
[ -f "$DIR/owner.did" ] || exit 0
[ -f "$DIR/controller.seed" ] || exit 0
DID="$(tr -d '[:space:]' < "$DIR/owner.did")"
SEED="$(tr -d '[:space:]' < "$DIR/controller.seed")"
MARKET_URL="${MARKET_URL:-https://api.arthneura.com}" \
  OWNER_DIR="$DIR" OWNER_DID="$DID" CONTROLLER_SEED="$SEED" \
  "$BIN"
