#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

. "$SCRIPT_DIR/lib.sh"

require_cmd git

"$SCRIPT_DIR/openapi-generate.sh"

status="$(
  cd "$ROOT_DIR"
  git status --porcelain -- backend/openapi/openapi.yaml frontend/packages/api-client/src/generated/schema.ts
)"

if [ -n "$status" ]; then
  printf '%s\n' "$status"
  fail "openapi artifacts are out of date"
fi
