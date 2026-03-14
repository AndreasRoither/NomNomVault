#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

. "$SCRIPT_DIR/lib.sh"

require_cmd docker
require_cmd migrate
require_cmd go

if ! docker info >/dev/null 2>&1; then
  fail "docker daemon is not reachable"
fi

NNV_TEMP_PG_IMAGE="${NNV_TEMP_PG_IMAGE:-postgres:18-alpine}"
NNV_TEMP_PG_PORT="${NNV_TEMP_PG_PORT:-15432}"
NNV_TEMP_PG_DB="${NNV_TEMP_PG_DB:-nomnomvault}"
NNV_TEMP_PG_USER="${NNV_TEMP_PG_USER:-nomnomvault}"
NNV_TEMP_PG_PASSWORD="${NNV_TEMP_PG_PASSWORD:-nomnomvault}"

case "$NNV_TEMP_PG_PORT" in
  '' | *[!0-9]*)
    fail "invalid NNV_TEMP_PG_PORT '$NNV_TEMP_PG_PORT': expected a numeric TCP port"
    ;;
esac

if [ "$NNV_TEMP_PG_PORT" -lt 1 ] || [ "$NNV_TEMP_PG_PORT" -gt 65535 ]; then
  fail "invalid NNV_TEMP_PG_PORT '$NNV_TEMP_PG_PORT': expected a value between 1 and 65535"
fi

VERIFY_DATABASE_URL="${NNV_VERIFY_DATABASE_URL:-postgres://$NNV_TEMP_PG_USER:$NNV_TEMP_PG_PASSWORD@127.0.0.1:$NNV_TEMP_PG_PORT/$NNV_TEMP_PG_DB?sslmode=disable}"
MIGRATIONS_DIR="$ROOT_DIR/backend/db/migrations"
container_name="nomnomvault-db-verify-$$"

trap 'cleanup_container "$container_name"' EXIT INT TERM

printf 'starting temp postgres: image=%s host_port=%s\n' "$NNV_TEMP_PG_IMAGE" "$NNV_TEMP_PG_PORT"
if ! start_temp_postgres "$container_name" "$NNV_TEMP_PG_PORT" "$NNV_TEMP_PG_IMAGE"; then
  fail "failed to start temp postgres on host port $NNV_TEMP_PG_PORT; check Docker and whether NNV_TEMP_PG_PORT is already in use"
fi

wait_for_postgres "$container_name" "$NNV_TEMP_PG_DB" "$NNV_TEMP_PG_USER" 30

if [ -d "$MIGRATIONS_DIR" ] && ls "$MIGRATIONS_DIR"/*.up.sql >/dev/null 2>&1; then
  migrate -path "$MIGRATIONS_DIR" -database "$VERIFY_DATABASE_URL" up
else
  printf '%s\n' "no checked-in migrations found; verifying ent schema against an empty database"
fi

(
  cd "$ROOT_DIR/backend"
  go run ./cmd/entschema --dsn "$VERIFY_DATABASE_URL" --require-clean
)
