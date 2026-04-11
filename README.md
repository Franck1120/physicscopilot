# PhysicsCopilot

> Real-time AI repair guidance — point your phone at the problem, get step-by-step instructions.

![CI](https://github.com/Franck1120/physicscopilot/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Flutter](https://img.shields.io/badge/Flutter-3.41-02569B?logo=flutter)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)

---

## What It Does

PhysicsCopilot uses your phone's camera and AI to guide you through physical repairs in real time. First vertical: **3D printer troubleshooting**.

Point the camera at your printer → AI detects the problem → step-by-step voice + visual instructions appear on screen.

## Features

- **Real-time vision** — streams camera frames to AI for instant diagnosis
- **Voice guidance** — text-to-speech reads each repair step aloud
- **Session history** — every repair is logged for future reference
- **Knowledge base** — curated 3D printer repair database with RAG retrieval
- **Offline-first design** — core flows degrade gracefully without connectivity

## Tech Stack

| Layer      | Technology                          |
|------------|-------------------------------------|
| Mobile     | Flutter 3.41 (iOS + Android)        |
| Backend    | Go 1.22 + Fiber + WebSocket         |
| AI         | Google Gemini Vision API            |
| Database   | Supabase (Postgres + Auth + pgvector) |
| Hosting    | Fly.io (server) + App stores (app)  |
| RAG        | pgvector + custom embeddings        |

## Monorepo Structure

```
physicscopilot/
├── app/        # Flutter mobile app
├── server/     # Go backend (WebSocket + REST)
├── kb/         # Knowledge base scrapers & embeddings
├── infra/      # Fly.io, Supabase, Docker config
└── web/        # Landing page (coming soon)
```

## Getting Started

### Prerequisites

- Flutter 3.41+ ([install](https://flutter.dev/docs/get-started/install))
- Go 1.22+ ([install](https://go.dev/dl/))
- Docker ([install](https://docs.docker.com/get-docker/))
- Supabase account

### Local Development

```bash
# Clone
git clone https://github.com/Franck1120/physicscopilot.git
cd physicscopilot

# Copy env template
cp .env.example .env
# → fill in SUPABASE_URL, SUPABASE_ANON_KEY, GEMINI_API_KEY

# Start backend
make dev-server

# Start Flutter app (in a separate terminal, with device connected)
make dev-app
```

### Run Tests

```bash
make test
```

### Deploy

```bash
make deploy
```

## Contributing

1. Fork the repository
2. Create your branch: `git checkout -b feature/your-feature`
3. Commit with conventional commits: `git commit -m "feat: add X"`
4. Push and open a Pull Request

## License

MIT — see [LICENSE](LICENSE)
