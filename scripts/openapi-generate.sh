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
    --dir cmd/api,internal/api/httpapi/auth,internal/api/httpapi/imports,internal/api/httpapi/recipes,internal/api/apicontract \
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

# Temporary workaround for a generator mismatch on media download endpoints.
# The backend handlers return binary content on 200 and JSON error envelopes on
# 401/404, but the current swag -> swagger2openapi pipeline cannot express that
# correctly. swag emits Swagger 2 first, and Swagger 2 only supports
# operation-level `produces`, not per-response content types. We patch the
# generated OpenAPI 3 document so the committed spec and generated TypeScript
# schema reflect actual runtime behavior.
#
# Until we move to another generator, this has to stay.
node - "$OPENAPI_FILE" <<'EOF'
const fs = require("fs");
const file = process.argv[2];
let content = fs.readFileSync(file, "utf8").replace(/\r\n/g, "\n");

const patchExactBlock = (from, to) => {
  if (content.includes(from)) {
    content = content.replace(from, to);
  }
};

patchExactBlock(
`  "/media/{mediaId}/original":
    get:
      description: Stream the original media bytes for the requested asset.
      parameters:
        - description: Media ID
          in: path
          name: mediaId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        "401":
          description: Unauthorized
          content:
            application/octet-stream:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
        "404":
          description: Not Found
          content:
            application/octet-stream:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
      summary: Fetch recipe media
      tags:
        - recipes`,
`  "/media/{mediaId}/original":
    get:
      description: Stream the original media bytes for the requested asset.
      parameters:
        - description: Media ID
          in: path
          name: mediaId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
        "404":
          description: Not Found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
      summary: Fetch recipe media
      tags:
        - recipes`,
);

patchExactBlock(
`  "/media/{mediaId}/thumbnail":
    get:
      description: Stream the stored thumbnail bytes for the requested asset.
      parameters:
        - description: Media ID
          in: path
          name: mediaId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        "401":
          description: Unauthorized
          content:
            application/octet-stream:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
        "404":
          description: Not Found
          content:
            application/octet-stream:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
      summary: Fetch recipe media thumbnail
      tags:
        - recipes`,
`  "/media/{mediaId}/thumbnail":
    get:
      description: Stream the stored thumbnail bytes for the requested asset.
      parameters:
        - description: Media ID
          in: path
          name: mediaId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
        "404":
          description: Not Found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/apicontract.ErrorResponse"
      summary: Fetch recipe media thumbnail
      tags:
        - recipes`,
);

fs.writeFileSync(file, content);
EOF

rm -f "$SWAGGER_FILE"

(
  cd "$ROOT_DIR"
  pnpm exec openapi-typescript backend/openapi/openapi.yaml -o frontend/packages/api-client/src/generated/schema.ts
)
