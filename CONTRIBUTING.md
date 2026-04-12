# Contributing to PhysicsCopilot

Thank you for your interest in contributing. This guide covers everything you need to get the project running locally and submit a pull request.

---

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | https://go.dev/dl |
| Flutter | 3.41+ | https://flutter.dev/docs/get-started/install |
| Docker + Compose | 24+ | https://docs.docker.com/get-docker |
| Node.js | 20+ (landing page only) | https://nodejs.org |
| staticcheck | latest | `go install honnef.co/go/tools/cmd/staticcheck@latest` |

You also need a **Gemini API key** (free tier works). Get one at https://aistudio.google.com.

---

## Getting Started

```bash
# 1. Fork and clone
git clone https://github.com/<your-fork>/physicscopilot.git
cd physicscopilot

# 2. Set up environment variables
cp server/.env.example server/.env
# Edit server/.env — fill in GEMINI_API_KEY at minimum

# 3. Start the local Supabase stack (Postgres + pgvector)
make docker-up

# 4. Run the Go server
make dev-server        # listens on :8080

# 5. In another terminal, run the Flutter app
make dev-app           # requires a connected device or running emulator
```

The landing page (static HTML) lives in `web/` and needs no build step — open `web/index.html` directly or run `npx serve web/`.

---

## Project Layout

```
physicscopilot/
├── server/                # Go backend (Fiber + WebSocket)
│   ├── cmd/server/        # main package — wires everything together
│   └── internal/
│       ├── handlers/      # HTTP + WebSocket handlers
│       ├── metrics/       # Prometheus metric definitions
│       ├── middleware/     # Rate limiter
│       ├── models/        # Shared types (IncomingMessage, OutgoingMessage)
│       └── services/      # Business logic (session, conversation, Gemini)
├── app/                   # Flutter mobile app
│   └── lib/
│       ├── screens/       # UI screens
│       ├── services/      # WebSocket client, Gemini proxy
│       ├── providers/     # Riverpod providers (session, WS, settings…)
│       ├── widgets/       # Shared widgets
│       └── utils/         # Constants, helpers
├── web/                   # Static landing page (HTML/CSS/JS)
├── supabase/              # Supabase migrations and seed data
├── infra/                 # Deployment config (Fly.io)
└── kb/                    # Knowledge-base entries (optional domain context modules)
```

---

## Development Workflow

### Go server

```bash
# Run with live reload (install air first: go install github.com/air-verse/air@latest)
cd server && air

# Or just restart manually
make dev-server
```

The server exposes:
- `GET /health` — liveness check
- `GET /api/sessions` — session management (CRUD)
- `POST /api/feedback` — per-step thumbs-up/down with optional comment
- `GET /metrics` — Prometheus metrics (Basic Auth required)
- `GET /ws` — WebSocket endpoint (camera frames in, AI guidance out)

### Flutter app

Connect a physical Android/iOS device or start an emulator, then:

```bash
make dev-app
```

To point the app at a custom server address, edit `app/lib/utils/constants.dart` before running.

### Landing page

```bash
cd web && npx serve .
# or just open web/index.html in a browser
```

---

## Testing

### Go

```bash
# Run all tests (race detector enabled, no test cache)
cd server && go test -count=1 -race ./...

# Run with coverage report (outputs server/coverage.out + coverage.html)
make test-coverage

# Run a single package
cd server && go test ./internal/handlers/... -v

# Static analysis
go vet ./...
staticcheck ./...
```

All tests must pass and `go vet` / `staticcheck` must be clean before submitting a PR.
Tests live next to the code they test (`*_test.go` in the same package).

### Flutter

```bash
cd app && flutter test
cd app && flutter analyze   # must show "No issues found"
```

---

## Code Style

### Go

- `go vet ./...` and `staticcheck ./...` must pass with zero warnings
- Follow standard Go conventions (`gofmt`-formatted, idiomatic naming)
- Error handling: check every error; never use `_` on errors in production paths
- No global mutable state in new code — prefer dependency injection through struct fields
- New public types and functions need a doc comment

### Flutter / Dart

- `flutter analyze` must pass with zero issues
- Use `const` constructors wherever possible
- Prefer `final` over `var`; avoid `dynamic`
- State management via Riverpod — don't add raw `StatefulWidget` for shared state
- Widget files: one public widget per file, named to match the file

### General

- No `TODO`/`FIXME` without a linked GitHub issue
- No debug logging (`print`, `log.Println`) left in production paths
- No secrets or `.env` files committed

---

## Pull Request Process

1. **Branch** from `main` using a descriptive prefix:
   - `feature/` — new functionality
   - `bugfix/` — bug corrections
   - `refactor/` — code improvements with no behavior change
   - `test/` — test additions or fixes

2. **Commit** atomically using [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add exponential backoff to WebSocket reconnect
   fix: close idle sessions older than 30 minutes
   test: add rate limiter integration tests
   ```

3. **Before opening the PR**, verify locally:
   ```bash
   cd server && go test -count=1 -race ./...   # Go tests pass (no cache)
   cd server && go vet ./... && staticcheck ./... # Go linting clean
   cd app && flutter test                         # Flutter tests pass
   cd app && flutter analyze                      # Dart analysis clean
   ```

4. **Open the PR** against `main`. Fill in:
   - What changed and why
   - How to test it manually
   - Screenshots for UI changes

5. A maintainer will review within a few days. Address feedback with new commits (don't amend; it makes review harder). Once approved, the maintainer will squash-merge.

---

## Reporting Issues

Open a GitHub issue with:
- Steps to reproduce
- Expected vs actual behavior
- Go/Flutter/OS versions (`go version`, `flutter --version`)
- Relevant logs from the server (`make dev-server` output)

For security vulnerabilities, please email instead of opening a public issue.
