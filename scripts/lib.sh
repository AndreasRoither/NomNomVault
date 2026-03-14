#!/usr/bin/env sh
set -eu

fail() {
  printf '%s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

cleanup_container() {
  container_name="${1:-}"
  if [ -z "$container_name" ]; then
    return 0
  fi

  docker rm -f "$container_name" >/dev/null 2>&1 || true
}

wait_for_postgres() {
  container_name="$1"
  db_name="$2"
  db_user="$3"
  max_attempts="$4"
  attempt=0

  until docker exec "$container_name" pg_isready -U "$db_user" -d "$db_name" >/dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ "$attempt" -ge "$max_attempts" ]; then
      fail "postgres did not become ready in time"
    fi
    sleep 1
  done
}

start_temp_postgres() {
  container_name="$1"
  host_port="$2"
  image="$3"

  docker run -d --rm \
    --name "$container_name" \
    -e "POSTGRES_DB=$NNV_TEMP_PG_DB" \
    -e "POSTGRES_USER=$NNV_TEMP_PG_USER" \
    -e "POSTGRES_PASSWORD=$NNV_TEMP_PG_PASSWORD" \
    -p "$host_port:5432" \
    "$image" >/dev/null
}
