set shell := ["sh", "-cu"]
set windows-shell := ["C:/Program Files/Git/bin/bash.exe", "-uc"]

default:
    @just --list

# Install all workspace dependencies.
install:
    @pnpm install

# Start both frontend development servers locally.
dev:
    @exec pnpm -r --parallel --stream --filter @nomnomvault/recipes-web --filter @nomnomvault/grocery-web run dev

# Start only the recipes frontend development server.
recipes-dev:
    @just --justfile frontend/apps/recipes-web/justfile dev

# Start only the grocery frontend development server.
grocery-dev:
    @just --justfile frontend/apps/grocery-web/justfile dev

# Start only the local infrastructure needed for development.
infra-up:
    @docker compose up -d postgres

# Stop the local infrastructure started via infra-up.
infra-down:
    @docker compose stop postgres

# Build all backend and frontend targets.
build:
    @just backend-build
    @just frontend-build

# Run all backend and frontend tests.
test:
    @just backend-test
    @just frontend-test

# Run all lint and formatting checks.
lint:
    @just backend-lint
    @just frontend-lint

# Run the aggregate validation gate used locally and in CI.
check:
    @just lint
    @just test
    @just build
    @just openapi-check

# Build the backend packages.
backend-build:
    @cd backend && go build ./...

# Run backend tests.
backend-test:
    @cd backend && go test ./...

# Check backend formatting.
backend-lint:
    @files="$(find backend -name '*.go' -type f)"; \
    output="$(gofmt -l $files)"; \
    if [ -n "$output" ]; then \
      echo "$output"; \
      exit 1; \
    fi

# Build all frontend apps.
frontend-build:
    @just recipes-build
    @just grocery-build

# Remove frontend build output and local framework caches for all apps.
frontend-clean:
    @just recipes-clean
    @just grocery-clean

# Run all frontend tests.
frontend-test:
    @just recipes-test
    @just grocery-test

# Run all frontend linters.
frontend-lint:
    @just recipes-lint
    @just grocery-lint

# Run all frontend checks.
frontend-check:
    @just recipes-check
    @just grocery-check

# Build the recipes frontend.
recipes-build:
    @just --justfile frontend/apps/recipes-web/justfile build

# Remove recipes frontend build output and local caches.
recipes-clean:
    @just --justfile frontend/apps/recipes-web/justfile clean

# Build the grocery frontend.
grocery-build:
    @just --justfile frontend/apps/grocery-web/justfile build

# Remove grocery frontend build output and local caches.
grocery-clean:
    @just --justfile frontend/apps/grocery-web/justfile clean

# Test the recipes frontend.
recipes-test:
    @just --justfile frontend/apps/recipes-web/justfile test

# Test the grocery frontend.
grocery-test:
    @just --justfile frontend/apps/grocery-web/justfile test

# Lint the recipes frontend.
recipes-lint:
    @just --justfile frontend/apps/recipes-web/justfile lint

# Lint the grocery frontend.
grocery-lint:
    @just --justfile frontend/apps/grocery-web/justfile lint

# Run formatter and lint checks for the recipes frontend.
recipes-check:
    @just --justfile frontend/apps/recipes-web/justfile check

# Run formatter and lint checks for the grocery frontend.
grocery-check:
    @just --justfile frontend/apps/grocery-web/justfile check

# Check the committed OpenAPI artifacts when present.
openapi-check:
    @if [ ! -f backend/openapi/openapi.yaml ]; then \
      echo "OpenAPI artifacts not present yet; skipping drift check."; \
      exit 0; \
    elif [ ! -f frontend/packages/api-client/src/generated/schema.ts ]; then \
      echo "backend/openapi/openapi.yaml exists but frontend/packages/api-client/src/generated/schema.ts is missing."; \
      exit 1; \
    else \
      echo "OpenAPI artifacts are present. Full drift enforcement lands with the OpenAPI pipeline ticket."; \
    fi

# Start the full local Docker Compose stack.
compose-up:
    @docker compose up --build -d postgres api worker recipes-web grocery-web caddy

# Stop the full local Docker Compose stack.
compose-down:
    @docker compose down --remove-orphans

# Follow logs from the compose stack.
compose-logs:
    @docker compose logs -f
