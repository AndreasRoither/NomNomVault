#!/usr/bin/env sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
ROOT_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

. "$SCRIPT_DIR/lib.sh"

require_cmd go
require_cmd pnpm

SWAGGER_FILE="$ROOT_DIR/backend/openapi/swagger.yaml"
OPENAPI_FILE="$ROOT_DIR/backend/openapi/openapi.yaml"
GENERATED_SCHEMA_FILE="$ROOT_DIR/frontend/packages/api-client/src/generated/schema.ts"

cleanup_swagger() {
  rm -f "$SWAGGER_FILE"
}

trap cleanup_swagger EXIT INT TERM

rm -f "$SWAGGER_FILE" "$OPENAPI_FILE"
mkdir -p "$ROOT_DIR/backend/openapi" "$(dirname "$GENERATED_SCHEMA_FILE")"

(
  cd "$ROOT_DIR/backend"
  go run github.com/swaggo/swag/cmd/swag@v1.16.6 init \
    --generalInfo main.go \
    --dir cmd/api,internal/api/httpapi/auth,internal/api/httpapi/recipes,internal/api/apicontract \
    --parseInternal \
    --output openapi \
    --outputTypes yaml
)

(
  cd "$ROOT_DIR"
  pnpm exec swagger2openapi backend/openapi/swagger.yaml \
    --yaml \
    --outfile backend/openapi/openapi.yaml
)

rm -f "$SWAGGER_FILE"

(
  cd "$ROOT_DIR"
  pnpm exec openapi-typescript backend/openapi/openapi.yaml -o frontend/packages/api-client/src/generated/schema.ts
)
