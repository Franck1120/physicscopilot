.PHONY: dev-server dev-app build-server build-app test deploy clean

FLUTTER := flutter
GO := go
DOCKER := docker
FLY := flyctl

# ── Development ──────────────────────────────────────────────────────────────

dev-server: ## Run Go server with hot reload
	cd server && $(GO) run ./cmd/server

dev-app: ## Run Flutter app (requires connected device or emulator)
	cd app && $(FLUTTER) run

dev: ## Start all services locally (requires Docker)
	docker-compose -f infra/docker-compose.yml up

# ── Build ─────────────────────────────────────────────────────────────────────

build-server: ## Build Docker image for Go server
	$(DOCKER) build -t physicscopilot-server ./server

build-app: ## Build Flutter APK (release)
	cd app && $(FLUTTER) build apk --release

build-server-binary: ## Build Go server binary locally
	cd server && $(GO) build -o bin/server ./cmd/server

# ── Test ──────────────────────────────────────────────────────────────────────

test: test-server test-app ## Run all tests

test-server: ## Run Go tests
	cd server && $(GO) test ./... -v -race

test-app: ## Run Flutter tests
	cd app && $(FLUTTER) test

# ── Deploy ────────────────────────────────────────────────────────────────────

deploy: ## Deploy server to Fly.io
	$(FLY) deploy --config infra/fly.toml

# ── Database ──────────────────────────────────────────────────────────────────

db-migrate: ## Apply Supabase schema
	@echo "Apply infra/supabase/schema.sql via Supabase dashboard or CLI"

# ── Utilities ─────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	cd server && rm -rf bin/
	cd app && $(FLUTTER) clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
