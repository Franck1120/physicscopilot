# Development Guide

## Prerequisites

| Tool | Min version | Install |
|------|-------------|---------|
| Go | 1.25 | https://go.dev/dl |
| Flutter | 3.41 | https://docs.flutter.dev/get-started/install |
| Docker | 24.x | https://docs.docker.com/get-docker |
| Make | any | `brew install make` / `apt install make` |

Quick install check:
```bash
bash scripts/setup.sh --check
```

---

## First-time setup

```bash
# 1. Clone
git clone https://github.com/Franck1120/physicscopilot
cd physicscopilot

# 2. Bootstrap (installs tools, deps, .env, git hooks)
bash scripts/setup.sh

# 3. Fill in secrets
cp .env.example .env   # if not done by setup.sh
# Edit .env: GEMINI_API_KEY, SUPABASE_URL, SUPABASE_JWT_SECRET
```

---

## Running locally

### Server only (no auth)
```bash
# Start Postgres via Docker (or use Supabase local)
docker compose up -d db

# Start server (hot-reload via Air)
cd server && air
# Server available at http://localhost:8080
```

### Full stack (server + app)
```bash
# Terminal 1: start server
make run

# Terminal 2: run app on connected device
cd app && flutter run --dart-define=SERVER_URL=http://192.168.x.x:8080
```

Use your LAN IP (not `localhost`) — the device can't reach the host's localhost.

### Dev container (Docker)
```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml up
```

This uses `Dockerfile.dev` which includes Air for hot-reload.

---

## Project structure

```
physicscopilot/
├── app/                    Flutter mobile app
│   ├── lib/
│   │   ├── config.dart     Runtime configuration
│   │   ├── main.dart       Entry point + theme
│   │   ├── models/         Data models
│   │   ├── providers/      Riverpod state
│   │   ├── screens/        UI screens
│   │   ├── services/       Business logic
│   │   ├── utils/          Constants, strings, extensions
│   │   └── widgets/        Reusable components
│   └── test/               Unit + widget tests
├── server/
│   ├── cmd/server/         main.go + e2e tests
│   └── internal/
│       ├── db/             Database queries
│       ├── handlers/       HTTP/WS handlers
│       ├── logger/         Structured logging
│       ├── metrics/        Error tracking
│       ├── middleware/      Auth, rate limit, logging
│       ├── models/         Data types
│       └── services/       AI, session, RAG logic
├── supabase/
│   └── migrations/         SQL migration files
├── docs/                   Documentation
├── infra/                  Nginx, K8s, monitoring
└── scripts/                Dev scripts
```

---

## Making changes

### Go server

```bash
cd server

# Build
go build ./...

# Test (with race detector)
go test -race ./...

# Lint
golangci-lint run

# Single test
go test -run TestHealthHandler ./internal/handlers/

# Coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Flutter app

```bash
cd app

# Analyze
flutter analyze --fatal-warnings

# Test
flutter test

# Format
dart format .

# Single test file
flutter test test/services/websocket_service_test.dart
```

---

## Environment variables

The server reads these env vars (from `.env` via `make run`, or from the shell directly):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `ENV` | `development` | `development` or `production` |
| `GEMINI_API_KEY` | — | Required: Google AI key |
| `SUPABASE_URL` | — | Required: Supabase project URL |
| `SUPABASE_JWT_SECRET` | — | Optional in dev, required in prod |
| `SUPABASE_SERVICE_KEY` | — | Optional: enables DB persistence |
| `DATABASE_URL` | — | Optional: Postgres DSN for sessions |
| `AI_BACKEND` | `gemini` | `gemini` or `openai` |

---

## Commit conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(ws): add streaming chunk support
fix(auth): handle expired JWT gracefully
refactor(camera): extract quality analysis to service
test(db): add pool config env-var coverage
docs: add development guide
chore(deps): bump go to 1.25.1
```

---

## Code review checklist

See `.github/pull_request_template.md` for the full checklist.
