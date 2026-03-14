#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
FORMATTER_CONFIG="$ROOT_DIR/.sql-formatter.json"

. "$SCRIPT_DIR/lib.sh"

if [ "$#" -ne 1 ]; then
  fail "usage: ./scripts/db-diff.sh <migration_name>"
fi

case "$1" in
  *[!a-z0-9_]* | "" | _* )
    fail "invalid migration name '$1': use lowercase letters, numbers, and underscores, starting with a letter or number"
    ;;
esac

require_cmd docker
require_cmd migrate
require_cmd go
require_cmd pnpm
require_cmd date

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
container_name="nomnomvault-db-diff-$$"
migration_name="$1"

trap 'cleanup_container "$container_name"' EXIT INT TERM

printf 'starting temp postgres: image=%s host_port=%s\n' "$NNV_TEMP_PG_IMAGE" "$NNV_TEMP_PG_PORT"
if ! start_temp_postgres "$container_name" "$NNV_TEMP_PG_PORT" "$NNV_TEMP_PG_IMAGE"; then
  fail "failed to start temp postgres on host port $NNV_TEMP_PG_PORT; check Docker and whether NNV_TEMP_PG_PORT is already in use"
fi

wait_for_postgres "$container_name" "$NNV_TEMP_PG_DB" "$NNV_TEMP_PG_USER" 30
mkdir -p "$MIGRATIONS_DIR"

if ls "$MIGRATIONS_DIR"/*.up.sql >/dev/null 2>&1; then
  migrate -path "$MIGRATIONS_DIR" -database "$VERIFY_DATABASE_URL" up
fi

diff_sql="$(
  cd "$ROOT_DIR/backend"
  go run ./cmd/entschema --dsn "$VERIFY_DATABASE_URL"
)"

if [ -z "$diff_sql" ]; then
  printf '%s\n' "schema already matches the checked-in migrations"
  exit 0
fi

version="$(date -u +%Y%m%d%H%M%S)"
up_file="$MIGRATIONS_DIR/${version}_${migration_name}.up.sql"
down_file="$MIGRATIONS_DIR/${version}_${migration_name}.down.sql"
up_file_name="${up_file##*/}"

if [ -e "$up_file" ] || [ -e "$down_file" ]; then
  fail "migration files already exist for timestamp $version; retry once the collision is resolved"
fi

printf '%s\n' "$diff_sql" > "$up_file"
cat > "$down_file" <<EOF
-- Manual down migration required for ${migration_name}.
-- Review ${up_file_name} and replace this placeholder with the reverse migration.
EOF

pnpm exec -- sql-formatter --fix --config "$FORMATTER_CONFIG" "$up_file"
pnpm exec -- sql-formatter --fix --config "$FORMATTER_CONFIG" "$down_file"

printf 'created %s\n' "$up_file"
printf 'created %s\n' "$down_file"
