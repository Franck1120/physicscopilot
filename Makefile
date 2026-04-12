.PHONY: dev dev-server dev-app test test-coverage build-server build-apk \
        lint docker-up docker-down clean help

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

# ── Clean ─────────────────────────────────────────────────────────────────────

clean: ## Remove build artefacts
	rm -rf bin/ server/coverage.out server/coverage.html app/build/

# ── Help ──────────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
