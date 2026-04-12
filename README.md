# PhysicsCopilot

### Your AI-powered repair assistant

[![CI](https://github.com/Franck1120/physicscopilot/actions/workflows/ci.yml/badge.svg)](https://github.com/Franck1120/physicscopilot/actions)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Flutter](https://img.shields.io/badge/Flutter-3.41-02569B?logo=flutter&logoColor=white)](https://flutter.dev)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Gemini](https://img.shields.io/badge/Gemini-2.5_Flash-4285F4?logo=google&logoColor=white)](https://deepmind.google/technologies/gemini/)

**Point your phone at a broken machine. Get step-by-step repair guidance in real time.**

PhysicsCopilot turns any smartphone into an AI repair technician — streaming live camera frames to Gemini Vision, diagnosing problems instantly, and walking users through fixes with voice + visual overlays. No manual needed.

---

## Why It Exists

$400B is spent annually on appliance and equipment repairs. Most of it goes to technicians for jobs that owners could fix themselves — if they had guidance. Repair manuals are static PDFs. YouTube tutorials don't know what *your* machine looks like right now.

PhysicsCopilot is the copilot for the physical world.

---

## Features

- **📷 Live vision diagnosis** — streams camera frames over WebSocket to Gemini 2.5 Flash for real-time problem detection
- **🔊 Voice-guided repair steps** — text-to-speech walks users through each action hands-free
- **🧠 RAG knowledge base** — 44 curated 3D printer failure modes with pgvector semantic search
- **📋 Session history** — every repair logged to Supabase for reference and model improvement
- **🔌 Works on your LAN** — zero cloud dependency for local testing; designed for offline-first degradation

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER'S PHONE                            │
│                                                                 │
│   Camera frames ──►  Flutter App  ──► TTS / AR overlay         │
│                           │                                     │
└───────────────────────────┼─────────────────────────────────────┘
                            │  WebSocket (ws://)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                       GO SERVER (Fiber)                         │
│                                                                 │
│   WebSocket handler ──► Frame processor ──► Gemini 2.5 Flash   │
│                                │                                │
│                                ▼                                │
│                    pgvector RAG (44 KB entries)                 │
│                                │                                │
│                                ▼                                │
│                    Supabase (session log + auth)                │
└─────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Run the server

```bash
git clone https://github.com/Franck1120/physicscopilot.git
cd physicscopilot

# Set environment variables
cp .env.example .env
# Fill in: GEMINI_API_KEY, SUPABASE_URL, SUPABASE_ANON_KEY

# Start with Docker Compose (server + local Postgres)
docker compose -f infra/docker-compose.yml up

# Or bare metal
make dev-server   # Go server on :8080
```

### 2. Run the app

```bash
# Point the app at your server (edit before building)
# app/lib/utils/constants.dart → _serverHost

cd app
flutter pub get
flutter run                      # on connected device/emulator

# Build APK for Android
flutter build apk --debug
```

---

## Tech Stack

| Layer       | Technology                               |
|-------------|------------------------------------------|
| Mobile      | Flutter 3.41 · iOS + Android             |
| Backend     | Go 1.22 · Fiber · WebSocket              |
| AI          | Gemini 2.5 Flash (vision + reasoning)    |
| Database    | Supabase · Postgres · pgvector           |
| RAG         | 44-entry KB · custom embeddings pipeline |
| Hosting     | Render (server) · Vercel (landing)       |
| Codebase    | ~7,400 LOC across Flutter + Go           |

---

## Roadmap

| Vertical        | Status        |
|-----------------|---------------|
| 🖨 3D Printing  | ✅ Live        |
| 🚗 Automotive   | Q3 2026       |
| ❄️ HVAC         | Q4 2026       |
| 🔌 Electronics  | 2027          |

Each vertical adds a domain-specific knowledge base and fine-tuned diagnostic prompts. The core vision + WebSocket pipeline is shared.

---

## Deploy on Render (free tier)

The repo includes a `render.yaml` — Render picks it up automatically.

1. Go to [render.com](https://render.com) and sign in
2. Click **New → Web Service** → **Connect a repository**
3. Select `Franck1120/physicscopilot` and click **Connect**
4. Render detects `render.yaml` and pre-fills all settings — click **Deploy**
5. In the **Environment** tab, add the secret variables:
   - `GEMINI_API_KEY` — your Google AI Studio key
   - `SUPABASE_URL` and `SUPABASE_ANON_KEY` — from your Supabase project
6. The service will be live at `https://physicscopilot-server.onrender.com`

**Build the Flutter app for production:**

```bash
cd app
flutter build apk --release \
  --dart-define=BACKEND_URL=wss://physicscopilot-server.onrender.com/ws
```

> **Note:** The free Render tier spins down after 15 minutes of inactivity.
> The first request after a cold start may take ~30 seconds.

---

## License

MIT — see [LICENSE](LICENSE)
