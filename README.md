# PhysicsCopilot

### AI-powered real-time guidance for physical work

[![CI](https://github.com/Franck1120/physicscopilot/actions/workflows/ci.yml/badge.svg)](https://github.com/Franck1120/physicscopilot/actions)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Flutter](https://img.shields.io/badge/Flutter-3.41-02569B?logo=flutter&logoColor=white)](https://flutter.dev)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Gemini](https://img.shields.io/badge/Gemini-2.5_Flash-4285F4?logo=google&logoColor=white)](https://deepmind.google/technologies/gemini/)

**Point your phone at any machine. Get step-by-step guidance in real time.**

PhysicsCopilot turns any smartphone into an AI assistant for physical work — streaming live camera frames to Gemini Vision, analyzing what it sees, and walking you through the task with voice guidance and visual overlays. No manual needed.

---

## Use Cases

| Domain | Examples |
|--------|---------|
| 🔧 Repairs | Appliances, electronics, mechanical failures |
| 🛠 Maintenance | HVAC, industrial equipment, automotive |
| 🔩 Assembly | Furniture, kit builds, component installations |
| 🔍 Inspections | Quality checks, visual audits, safety reviews |

---

## Why It Exists

Repair manuals are static PDFs. Tutorials don't know what *your* machine looks like right now. Senior technicians spend hours guiding junior staff on problems the AI could handle in seconds.

PhysicsCopilot is the copilot for the physical world — AI that sees what you see and guides you through it.

---

## What's Implemented

- **📷 Live vision diagnosis** — streams camera frames over WebSocket to Gemini 2.5 Flash for real-time analysis
- **🔊 Voice-guided steps** — text-to-speech walks users through each action hands-free
- **🗺 Visual overlays** — bounding boxes and directional arrows drawn over the live camera feed
- **📋 RAG knowledge base** — keyword-ranked retrieval over `kb/data/problems.json` for domain-specific context
- **🔐 Secure endpoints** — JWT auth on WebSocket, HTTP Basic Auth on `/metrics`, CORS from env
- **📊 Prometheus metrics** — request count and latency exposed at `/metrics`
- **🔌 Session REST API** — `POST/GET/DELETE /api/sessions` for session lifecycle management
- **💾 Optional Postgres persistence** — session history stored in Postgres/Supabase when `DATABASE_URL` is set
- **📖 OpenAPI spec** — full API documentation served at `GET /api/docs` (OpenAPI 3.0.3, YAML)

**Planned (not yet live):**
- Semantic search over uploaded manuals (pgvector)
- Multiple domain verticals beyond the bundled equipment profiles

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER'S PHONE                            │
│                                                                 │
│   Camera frames ──►  Flutter App  ──► TTS / AR overlay         │
│                           │                                     │
└───────────────────────────┼─────────────────────────────────────┘
                            │  WebSocket (wss://)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     GO SERVER (Fiber)                           │
│                                                                 │
│   WebSocket handler ──► Frame processor ──► Gemini 2.5 Flash   │
│                                │                                │
│                                ▼                                │
│              RAG service  ──► KB context injection              │
│                                                                 │
│             Session REST API · Prometheus metrics               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Quick Start — Local

### 1. Run the server

```bash
git clone https://github.com/Franck1120/physicscopilot.git
cd physicscopilot

# Copy and fill environment variables
cp .env.example server/.env
# Required: GEMINI_API_KEY
# Optional: METRICS_PASSWORD (protects /metrics)

# Run with Docker Compose (includes Supabase)
docker compose up

# Or bare Go (no Docker)
make server-run
```

### 2. Build and run the app

```bash
cd app
flutter pub get
flutter run   # on a connected device or emulator
```

### 3. Without a Gemini API key

Set `CLIPROXY_URL` to point to a local
[CLIProxyAPI](https://github.com/simonb97/cliproxy) instance.
The server falls back to it automatically when `GEMINI_API_KEY` is unset.

---

## Deploy — Render (recommended)

Render auto-deploys from the root `Dockerfile` whenever you push to `main`.

### Step-by-step

1. **Create a Web Service** on [render.com](https://render.com) → New → Web Service.

2. **Connect your GitHub repo** (`Franck1120/physicscopilot`).

3. **Select Docker** as the environment.
   Render auto-detects the root `Dockerfile`; leave defaults:
   - Dockerfile path: `Dockerfile`
   - Docker context: `.` (repo root)

4. **Add environment variables** in the Render dashboard:

   | Variable | Value | Notes |
   |----------|-------|-------|
   | `GEMINI_API_KEY` | `your-key` | Required |
   | `APP_ENV` | `production` | Enables JSON logs and strict CORS |
   | `ALLOWED_ORIGINS` | `https://your-app.onrender.com` | Comma-separated list of allowed origins |
   | `METRICS_PASSWORD` | `choose-a-password` | Protects `/metrics` with HTTP Basic Auth |
   | `SUPABASE_URL` | `https://xxx.supabase.co` | Optional — for auth/history |
   | `SUPABASE_ANON_KEY` | `eyJ...` | Optional |
   | `SUPABASE_JWT_SECRET` | `your-secret` | Optional — enables JWT on WS |
   | `DATABASE_URL` | `postgres://...` | Optional — Postgres for persistent session history |

5. **Click Deploy**. The health check (`GET /health`) confirms when the
   service is ready.

> The knowledge base (`kb/data/problems.json`) is bundled in the Docker
> image — no extra setup needed.

---

## Build APK for Production

After deploying the server, build a release APK that points at your Render URL:

```bash
cd app

# Replace with your actual Render URL (no trailing slash, no https://)
flutter build apk --release \
  --dart-define=SERVER_URL=physicscopilot-server.onrender.com

# APK output:
# app/build/app/outputs/flutter-apk/app-release.apk
```

Install on device:
```bash
adb install build/app/outputs/flutter-apk/app-release.apk
```

---

## Makefile targets

```
make server-build     Build Go server binary → bin/server
make server-test      Run Go tests (race detector)
make server-run       Run server locally
make app-build        Build Flutter debug APK
make app-analyze      Flutter static analysis
make docker-build     Build root Docker image
make docker-up        Start local Supabase stack
make docker-down      Stop local Supabase stack
make deploy-render    Print Render deployment checklist
make lint             go vet + golangci-lint + flutter analyze
make clean            Remove all build artefacts
```

---

## Tech Stack

| Layer    | Technology                              |
|----------|-----------------------------------------|
| Mobile   | Flutter 3.41 · iOS + Android            |
| Backend  | Go 1.25 · Fiber v2 · WebSocket          |
| AI       | Gemini 2.5 Flash (vision + reasoning)   |
| KB / RAG | JSON knowledge base · keyword ranking   |
| Hosting  | Render (server) · Vercel (landing page) |

---

## Roadmap

| Vertical             | Status  |
|----------------------|---------|
| 🔧 Repairs (general) | ✅ Live  |
| 🗄 Session history   | ✅ Live  |
| 🧠 RAG semantic search | Q3 2026 |
| 🚗 Automotive        | Q4 2026 |
| ❄️ HVAC              | 2027    |

---

## License

MIT — see [LICENSE](LICENSE)
