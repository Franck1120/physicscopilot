.PHONY: dev dev-server dev-app test test-coverage build-server build-apk \
        lint docker-up docker-down docker-push pre-commit clean help

# ── Development ───────────────────────────────────────────────────────────────

dev: docker-up ## Start Supabase stack, then run Go server
	cd server && go run ./cmd/server/

dev-server: ## Run Go server without Docker (no JAVA_HOME override)
	cd server && go run ./cmd/server/

dev-app: ## Run Flutter app (requires connected device or emulator)
	cd app && flutter run

# ── Testing ───────────────────────────────────────────────────────────────────

test: ## Run Go tests with race detector
	cd server && go test -v -race ./...

test-coverage: ## Run Go tests with coverage report (opens coverage.html)
	cd server && go test -coverprofile=coverage.out ./... \
		&& go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: server/coverage.html"

# ── Build ─────────────────────────────────────────────────────────────────────

build-server: ## Build optimised Go server binary → bin/server
	mkdir -p bin
	cd server && CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o ../bin/server \
		./cmd/server/

build-apk: ## Build Flutter debug APK
	cd app && flutter build apk --debug

build-apk-release: ## Build Flutter release APK
	cd app && flutter build apk --release

# ── Lint ──────────────────────────────────────────────────────────────────────

lint: ## Run Go vet + staticcheck and Flutter analyze
	cd server && go vet ./...
	@which staticcheck > /dev/null && cd server && staticcheck ./... || \
		echo "staticcheck not installed — run: go install honnef.co/go/tools/cmd/staticcheck@latest"
	cd app && flutter analyze

# ── Docker / Supabase ─────────────────────────────────────────────────────────

docker-up: ## Start Supabase local stack (docker compose)
	docker compose up -d

docker-down: ## Stop Supabase local stack
	docker compose down

docker-push: ## Build multi-platform image and push to GHCR
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t ghcr.io/franck1120/physicscopilot:latest --push .

# ── Pre-commit ───────────────────────────────────────────────────────────────

pre-commit: ## Run all pre-commit hooks on every file
	pre-commit run --all-files

# ── Coverage ─────────────────────────────────────────────────────────────────

coverage: test-coverage ## Alias: run tests with HTML coverage report

# ── Release ──────────────────────────────────────────────────────────────────

APP_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

release: ## Build release binaries for linux/amd64 with version injection
	mkdir -p bin
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
		-ldflags="-s -w -X main.version=$(APP_VERSION) -X 'main.buildDate=$(BUILD_DATE)'" \
		-o ../bin/server-linux-amd64 \
		./cmd/server/
	@echo "Release binary: bin/server-linux-amd64 (version=$(APP_VERSION) built=$(BUILD_DATE))"

# ── Clean ─────────────────────────────────────────────────────────────────────

clean: ## Remove build artefacts
	rm -rf bin/ server/coverage.out server/coverage.html app/build/

# ── Help ──────────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
