set shell := ["sh", "-cu"]
set windows-shell := ["C:/Program Files/Git/bin/bash.exe", "-uc"]

default:
    @just --list

# Install all workspace dependencies.
install:
    @pnpm install

# Start both frontend development servers.
dev:
    @pnpm -r --parallel --stream --filter @nomnomvault/recipes-web --filter @nomnomvault/grocery-web run dev

# Start the recipes frontend development server.
recipes-dev:
    @just --justfile frontend/apps/recipes-web/justfile dev

# Start the grocery frontend development server.
grocery-dev:
    @just --justfile frontend/apps/grocery-web/justfile dev

# Open the recipes frontend in the default browser.
recipes-open:
    @just --justfile frontend/apps/recipes-web/justfile open

# Open the grocery frontend in the default browser.
grocery-open:
    @just --justfile frontend/apps/grocery-web/justfile open

# Run the current frontend test suites.
test:
    @just frontend-test

# Run the current frontend linters.
lint:
    @just frontend-lint

# Run the current frontend formatter checks.
check:
    @just frontend-check

# Build all frontend apps.
frontend-build:
    @just recipes-build
    @just grocery-build

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

# Build the grocery frontend.
grocery-build:
    @just --justfile frontend/apps/grocery-web/justfile build

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

# Check the OpenAPI spec and generated client for drift.
openapi-check:
    @echo "TODO: verify the OpenAPI spec and generated client are in sync"

# Start the local Docker Compose stack.
compose-up:
    @echo "TODO: start postgres, garage, api, worker, web, and reverse proxy"

# Stop the local Docker Compose stack.
compose-down:
    @echo "TODO: stop the local Docker Compose stack"
