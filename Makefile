.PHONY: \
  server-build server-test server-run \
  app-build app-analyze \
  docker-build docker-up docker-down \
  deploy-render \
  dev dev-server dev-app \
  test test-coverage lint \
  build-server build-apk build-apk-release \
  clean help

# ── Server ────────────────────────────────────────────────────────────────────

server-build: ## Build optimised Go server binary → bin/server
	mkdir -p bin
	cd server && CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o ../bin/server \
		./cmd/server/

server-test: ## Run Go tests with race detector
	cd server && go test -v -race -count=1 ./...

server-run: ## Run Go server locally (no Docker required)
	cd server && go run ./cmd/server/

# ── App ───────────────────────────────────────────────────────────────────────

app-build: ## Build Flutter debug APK
	cd app && flutter build apk --debug

app-analyze: ## Run Flutter static analysis (fatal on infos)
	cd app && flutter analyze --fatal-infos

# ── Docker ────────────────────────────────────────────────────────────────────

docker-build: ## Build the root Docker image (Render target)
	docker build -t physicscopilot-server:local .

docker-up: ## Start Supabase local stack (docker compose)
	docker compose up -d

docker-down: ## Stop Supabase local stack
	docker compose down

# ── Deploy ────────────────────────────────────────────────────────────────────

deploy-render: ## Show Render deploy instructions
	@echo ""
	@echo "  Render deployment checklist"
	@echo "  ─────────────────────────────────────────────────"
	@echo "  1. Push main → GitHub triggers Render auto-deploy"
	@echo "  2. Render uses: dockerfilePath=Dockerfile, dockerContext=."
	@echo "  3. Required env vars (set in Render dashboard):"
	@echo "     GEMINI_API_KEY      your Gemini API key"
	@echo "     METRICS_PASSWORD    password for /metrics endpoint"
	@echo "     ALLOWED_ORIGINS     comma-separated allowed origins"
	@echo "     SUPABASE_URL        your Supabase project URL"
	@echo "     SUPABASE_ANON_KEY   your Supabase anon key"
	@echo "     SUPABASE_JWT_SECRET your Supabase JWT secret"
	@echo "  4. Health check: GET /health"
	@echo ""

# ── Development ───────────────────────────────────────────────────────────────

dev: docker-up server-run ## Start local stack then run Go server

dev-app: ## Run Flutter app (requires connected device or emulator)
	cd app && flutter run

# ── Quality ───────────────────────────────────────────────────────────────────

test: server-test ## Alias → server-test

test-coverage: ## Run Go tests with HTML coverage report
	cd server && go test -coverprofile=coverage.out ./... \
		&& go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: server/coverage.html"

lint: ## Run Go vet + golangci-lint and Flutter analyze
	cd server && go vet ./...
	@which golangci-lint > /dev/null 2>&1 \
		&& cd server && golangci-lint run ./... \
		|| echo "golangci-lint not installed — see https://golangci-lint.run/usage/install/"
	cd app && flutter analyze

# ── Compat aliases ────────────────────────────────────────────────────────────

build-server: server-build ## Alias → server-build
build-apk: app-build ## Alias → app-build

build-apk-release: ## Build Flutter release APK
	cd app && flutter build apk --release

dev-server: server-run ## Alias → server-run

# ── Clean ─────────────────────────────────────────────────────────────────────

clean: ## Remove all build artefacts
	rm -rf bin/ server/coverage.out server/coverage.html app/build/

# ── Help ──────────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'
