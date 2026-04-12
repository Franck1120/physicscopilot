# PhysicsCopilot

### AI-powered real-time guidance for physical work

[![CI](https://github.com/Franck1120/physicscopilot/actions/workflows/ci.yml/badge.svg)](https://github.com/Franck1120/physicscopilot/actions)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Flutter](https://img.shields.io/badge/Flutter-3.41-02569B?logo=flutter&logoColor=white)](https://flutter.dev)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](https://go.dev)
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
- **📋 Context selection** — optional device profiles for domain-specific guidance (extensible KB module)
- **🔌 LAN-first** — zero cloud dependency for local testing; runs entirely on your network

**Planned (not yet live):**
- Persistent session history (Supabase)
- RAG knowledge base (pgvector semantic search over uploaded manuals)
- Multiple domain verticals beyond the bundled 3D printer profiles

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
│                     GO SERVER (Fiber)                           │
│                                                                 │
│   WebSocket handler ──► Frame processor ──► Gemini 2.5 Flash   │
│                                │                                │
│                                ▼                                │
│                    Session + conversation context               │
│                                                                 │
│                    Prometheus metrics (/metrics)                │
└─────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Run the server

```bash
git clone https://github.com/Franck1120/physicscopilot.git
cd physicscopilot

# Set environment variables
cp server/.env.example server/.env
# Fill in: GEMINI_API_KEY (required)
# Optional: METRICS_PASSWORD (default: "metrics-secret")

# Start with Docker Compose
docker compose up

# Or bare metal
make dev-server   # Go server on :8080
```

### 2. Run the app

```bash
# Point the app at your server before building:
# app/lib/utils/constants.dart → _serverHost

cd app
flutter pub get
flutter run                      # on connected device/emulator

# Build APK for Android
flutter build apk --debug
```

### 3. Without a Gemini API key

If `GEMINI_API_KEY` is not set, the server falls back to a local
[CLIProxyAPI](https://github.com/simonb97/cliproxy) Docker container at
`CLIPROXY_URL` (default: `http://localhost:8085`).

---

## Tech Stack

| Layer    | Technology                            |
|----------|---------------------------------------|
| Mobile   | Flutter 3.41 · iOS + Android          |
| Backend  | Go 1.22 · Fiber · WebSocket           |
| AI       | Gemini 2.5 Flash (vision + reasoning) |
| Hosting  | Fly.io (server) · Vercel (landing)    |

---

## Roadmap

| Vertical          | Status  |
|-------------------|---------|
| 🔧 Repairs (general) | ✅ Live  |
| 🗄 Session history  | Q3 2026 |
| 🧠 RAG knowledge base | Q3 2026 |
| 🚗 Automotive      | Q4 2026 |
| ❄️ HVAC            | 2027    |

Each domain vertical adds a context-specific knowledge base and tailored diagnostic prompts. The core vision + WebSocket pipeline is shared across all of them.

---

## License

MIT — see [LICENSE](LICENSE)
