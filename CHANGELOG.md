# Changelog

All notable changes to PhysicsCopilot are documented in this file.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

### Changed

### Fixed

---

## [0.18.0] - 2026-04-12

### Added
- Wave 3: CI workflows (`flutter.yml`, `security.yml`), Dependabot config, CODEOWNERS, PR template
- Wave 3: `scripts/setup.sh` and `scripts/test.sh` for dev onboarding
- Wave 3: `.editorconfig`, `.golangci.yml`, `.gitattributes` for consistent tooling
- Wave 3: `Dockerfile.dev` with Air hot-reload for development
- Wave 3: `infra/nginx.conf` with TLS, WebSocket proxy, security headers
- Wave 3: `infra/k8s/deployment.yaml` and `service.yaml` with HPA and Ingress
- Wave 3: Comprehensive documentation: DEPLOYMENT, DEVELOPMENT, TESTING, MONITORING, SECURITY_AUDIT, PERFORMANCE, MOBILE_BUILD, KB_FORMAT, FAQ, ROADMAP, GLOSSARY
- Wave 3: GoDoc comments on all Go source files
- Wave 3: Dartdoc comments on all Dart public APIs
- Wave 3: Stricter `analysis_options.yaml` with `prefer_const_constructors` and additional lint rules
- `POST /api/feedback` endpoint — per-step thumbs-up/down with optional comment; persists to Postgres when `DATABASE_URL` is set, otherwise logs the event
- `feedback` Supabase table with UUID PK, FK to sessions, CHECK constraint on rating, RLS enabled
- `SaveFeedback` on the `DBBackend` interface and `DBService` implementation

---

## [0.17.0] — 2026-04-12

### Added
- CI coverage upload: `go test -race -coverprofile=coverage.out` artefact in GitHub Actions

### Changed
- Sync from upstream main: resolved `session_screen.dart` conflict (offline cache branch ordering corrected)

---

## [0.16.0] — 2026-04-11

### Added
- Flutter Supabase auth: `AuthService`, `LoginScreen`, `GoRouter` redirect guard
- JWT bearer token passed in WebSocket query param (`?token=<jwt>`)
- `UserRateLimiter` per JWT `sub` claim (30 msg/min, burst 5)
- Server-side JWT warning when `SUPABASE_JWT_SECRET` is absent in production

### Changed
- `GeminiResponse` renamed to `AIResponse` across all layers (no breaking API change)
- `frameHashEntry` now stores `recordedAt` timestamp; TTL 30 min to prevent memory leaks
- `AnalyzeFrame` applies `apiLimiter.Wait(ctx)` before each Gemini call

### Fixed
- JPEG magic-byte check (`FF D8 FF`) before frame processing to reject non-JPEG frames
- HTML strip + 5000-char cap on incoming text messages (XSS mitigation)
- CI workflow: coverage upload artefact added

---

## [0.15.0] — 2026-04-10

### Added
- Dark/light/system theme toggle (`ThemeMode`) in settings panel

### Fixed
- `GoRouter` moved out of `build()` into `ConsumerStatefulWidget` field to avoid recreation on rebuild

---

## [0.14.0] — 2026-04-09

### Added
- Multi-step guided repair flow with `CurrentStep`/`TotalSteps` state
- `step_update` WebSocket message type for server-pushed step changes
- Overlay annotation system: bounding boxes + action arrows rendered on camera frame
- Typewriter text animation for assistant responses

---

## [0.13.0] — 2026-04-08

### Added
- BCP-47 language selection (Italian default) forwarded to Gemini prompt
- `language` field in `AppSettings` and WebSocket handshake query param
- `WsActiveSessionsByLanguage` Prometheus gauge

---

## [0.12.0] — 2026-04-07

### Added
- Prometheus metrics: `ws_active_connections`, `ws_frames_processed_total`, `gemini_errors_total`, `ai_inference_duration_seconds`
- `/metrics` endpoint protected with HTTP Basic Auth (`METRICS_USER`/`METRICS_PASSWORD`)
- Structured JSON logging via `log/slog`

---

## [0.11.0] — 2026-04-06

### Added
- TF-IDF RAG service (`MemoryVectorStore`) for knowledge-base context enrichment
- `KB_PATH` env var to point at the JSONL knowledge base file
- `RAGService.QueryKB` / `FormatForPrompt` integrated into frame and text analysis

---

## [0.10.0] — 2026-04-05

### Added
- Optional Postgres write-through via `DBBackend` interface (`DBService` + pgx/v5)
- `DATABASE_URL` env var triggers DB init; in-memory store remains authoritative
- Session hydration from DB on startup (`HydrateFromDB`)
- `session_steps` table and `SaveSessionStep`

---

## [0.9.0] — 2026-04-04

### Added
- Background session cleanup (expired after 30 min) ticker every 5 min
- `CleanupExpiredSessions` on `SessionService`

---

## [0.8.0] — 2026-04-03

### Added
- Flutter local notification service (`NotificationService`) for push alerts
- Offline cache: last assistant response shown when WebSocket is disconnected
- Reconnect back-off with jitter in `WebSocketService`

---

## [0.7.0] — 2026-04-02

### Added
- Gemini proxy fallback (`GEMINI_PROXY_URL`) when `GEMINI_API_KEY` is absent
- `AnalyzeFrame` call to proxy with same request/response shape

---

## [0.6.0] — 2026-04-01

### Added
- `GET /api/docs` — embedded OpenAPI 3.1 spec
- `WSAuthMiddleware` — validates Supabase JWT in dev (passthrough) and prod mode
- CORS middleware with `ALLOWED_ORIGINS` env var; dev defaults to `*`
- HSTS header enforced in production

---

## [0.5.0] — 2026-03-31

### Added
- IP-based rate limiter (`IPRateLimiter`) on `/health` and all `/api` routes
- Request timeout middleware (30 s per request)
- Panic recovery middleware with stack-trace logging

---

## [0.4.0] — 2026-03-30

### Added
- `GET /health` endpoint with uptime, version, active WS connections, DB ping
- Goroutine-safe `WSHandler.CloseAll()` for graceful shutdown

---

## [0.3.0] — 2026-03-29

### Added
- Session REST API: `POST /api/sessions`, `GET /api/sessions`, `GET /api/sessions/:id`, `DELETE /api/sessions/:id`
- `SessionService` in-memory store with `CreateSession`, `GetSessionSnapshot`, `DeleteSession`, `ListSessions`

---

## [0.2.0] — 2026-03-28

### Added
- WebSocket server (`/ws`) with frame and text message handling
- Conversation history per session (`BuildContextForGemini`)
- Duplicate-frame deduplication by SHA-256 hash prefix

---

## [0.1.0] — 2026-03-27

### Added
- Go Fiber server skeleton with `newFiberApp` factory
- Flutter app scaffold: `main.dart`, `WebSocketService`, `WebSocketProvider`
- GitHub Actions CI: `go build`, `go test`, `flutter analyze`, `flutter test`
