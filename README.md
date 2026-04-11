# PhysicsCopilot

> Real-time AI repair guidance — point your phone at the problem, get step-by-step voice instructions.

![CI](https://github.com/Franck1120/physicscopilot/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Flutter](https://img.shields.io/badge/Flutter-3.41-02569B?logo=flutter)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)

---

<!-- Screenshot placeholder — replace with actual screenshot before launch -->
<!--
![PhysicsCopilot screenshot](docs/screenshot.png)
-->

## What It Does

PhysicsCopilot uses your phone's camera and AI to guide you through physical repairs in real time. First vertical: **3D printer troubleshooting**.

Point the camera at your printer → AI detects the problem → step-by-step voice + visual instructions appear on screen.

## Features

- **Real-time vision** — streams camera frames to Gemini Vision for instant diagnosis
- **Voice guidance** — text-to-speech reads each repair step aloud
- **Session history** — every repair is logged for future reference
- **Knowledge base** — curated 3D printer repair database with RAG retrieval
- **Offline-first design** — core flows degrade gracefully without connectivity

## Tech Stack

| Layer      | Technology                            |
|------------|---------------------------------------|
| Mobile     | Flutter 3.41 (iOS + Android)          |
| Backend    | Go 1.22 + Fiber + WebSocket           |
| AI         | Google Gemini Vision API              |
| Database   | Supabase (Postgres + Auth + pgvector) |
| Hosting    | Fly.io (server) · Vercel (landing)    |
| RAG        | pgvector + custom embeddings          |

## Architecture

```
physicscopilot/
├── app/          # Flutter mobile app (camera, voice, AR overlay)
│   ├── lib/
│   └── test/
├── server/       # Go backend (WebSocket, REST, RAG)
│   ├── cmd/server/
│   └── internal/
│       ├── handlers/
│       ├── middleware/
│       ├── models/
│       └── services/
├── kb/           # Knowledge base scrapers & embeddings pipeline
├── infra/        # Infrastructure config
│   ├── Dockerfile.fly      # Fly.io optimised build
│   ├── fly.toml            # Fly.io app config
│   ├── docker-compose.yml  # Local dev stack
│   └── supabase/           # DB schema & migrations
└── web/          # Static landing page
    ├── index.html
    └── vercel.json
```

Data flow:
```
Flutter app  ──WS──►  Go server  ──►  Gemini Vision API
                          │
                          ▼
                      Supabase (session log + pgvector RAG)
```

## Getting Started

### Prerequisites

- [Flutter 3.41+](https://flutter.dev/docs/get-started/install)
- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Supabase account](https://supabase.com) or local Supabase CLI
- Google Gemini API key from [Google AI Studio](https://aistudio.google.com/apikey)

### Local Development

```bash
# Clone
git clone https://github.com/Franck1120/physicscopilot.git
cd physicscopilot

# Copy env template and fill in your secrets
cp .env.example .env
# Required: SUPABASE_URL, SUPABASE_ANON_KEY, GEMINI_API_KEY

# Option A — Docker Compose (backend + local Postgres)
docker compose -f infra/docker-compose.yml up

# Option B — bare metal
make dev-server          # starts Go server on :8080
make dev-app             # starts Flutter app (requires device/emulator)
```

### Running Tests

```bash
make test          # all tests (Go + Flutter)
make test-server   # Go unit tests only
make test-app      # Flutter tests only
```

### Deploy

#### Server → Fly.io

```bash
# First time: create the app and set secrets
flyctl auth login
flyctl apps create physicscopilot-server --config infra/fly.toml
make secrets-set GEMINI_API_KEY=... SUPABASE_URL=... SUPABASE_ANON_KEY=...

# Subsequent deploys
make deploy-server
```

Required secrets (set via `flyctl secrets set` or `make secrets-set`):

| Secret              | Description                              |
|---------------------|------------------------------------------|
| `GEMINI_API_KEY`    | Google Gemini Vision API key             |
| `SUPABASE_URL`      | Supabase project URL                     |
| `SUPABASE_ANON_KEY` | Supabase anon/public key                 |
| `GEMINI_BASE_URL`   | *(optional)* Custom Gemini proxy base URL |

#### Landing Page → Vercel

```bash
cd web
vercel deploy --prod
```

Or connect the `web/` directory to a Vercel project via the dashboard — `vercel.json` is already configured.

## Contributing

1. Fork the repository
2. Create your branch: `git checkout -b feature/your-feature`
3. Write tests first (TDD)
4. Commit with [Conventional Commits](https://www.conventionalcommits.org/): `git commit -m "feat: add X"`
5. Push and open a Pull Request against `main`

Please run `make test` and ensure all checks pass before opening a PR.

## License

MIT — see [LICENSE](LICENSE)
