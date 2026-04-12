# Deployment Guide

## Overview

PhysicsCopilot uses a two-tier deploy:
- **Server**: Docker container on Render.com (automatic deploy on push to `main`)
- **App**: Manual Flutter build → Play Store / App Store
- **Database**: Supabase managed Postgres (migrations via Supabase CLI)

---

## Server Deploy (Render.com)

### Automatic deploys
Every push to `main` triggers a new deploy via `render.yaml`. No manual action needed.

### Manual deploy
```bash
# Trigger a manual deploy from the Render dashboard, or:
curl -X POST "https://api.render.com/deploy/srv-<ID>?key=<API_KEY>"
```

### Environment variables (set in Render dashboard)

| Variable | Required | Description |
|----------|----------|-------------|
| `GEMINI_API_KEY` | Yes | Google AI Studio API key |
| `SUPABASE_URL` | Yes | Your Supabase project URL |
| `SUPABASE_JWT_SECRET` | Yes | JWT secret from Supabase → Settings → API |
| `SUPABASE_SERVICE_KEY` | Yes | Service role key (bypasses RLS for server ops) |
| `DATABASE_URL` | Optional | Postgres connection string (enables session persistence) |
| `PORT` | Auto | Set by Render automatically |
| `ENV` | Optional | Set to `production` to enable production logging |

### Health check
The server exposes `GET /health`. Render uses this endpoint to verify the service is running before routing traffic.

```bash
curl https://your-app.onrender.com/health
# {"status":"ok","version":"0.17.0","uptime_seconds":1234,"active_connections":3}
```

### Cold starts
Render free tier spins down after 15 minutes of inactivity. First request after inactivity takes 20–30s. Upgrade to Starter plan to avoid cold starts.

---

## Database Migrations

```bash
# Install Supabase CLI
brew install supabase/tap/supabase

# Link to your project
supabase link --project-ref <PROJECT_REF>

# Apply pending migrations
supabase db push

# Check migration status
supabase migration list
```

Migrations live in `supabase/migrations/`. Always test on a staging Supabase project before running on production.

---

## Flutter App Build

### Android

```bash
cd app

# Debug build (for testing)
flutter build apk --debug

# Release build (for Play Store)
flutter build appbundle --release \
  --dart-define=SERVER_URL=https://your-app.onrender.com

# Output: app/build/app/outputs/bundle/release/app-release.aab
```

The release build requires a signing keystore. See [Flutter docs on signing](https://docs.flutter.dev/deployment/android).

### iOS

```bash
cd app

# Release build (requires macOS + Xcode 15+)
flutter build ios --release \
  --dart-define=SERVER_URL=https://your-app.onrender.com

# Then open Xcode and archive for App Store submission
open ios/Runner.xcworkspace
```

See `docs/MOBILE_BUILD.md` for the full release checklist.

---

## Docker (self-hosted)

```bash
# Build and start
docker compose up --build -d

# View logs
docker compose logs -f server

# Stop
docker compose down
```

The production `docker-compose.yml` binds to port 8080. Place nginx in front for TLS (see `infra/nginx.conf`).

---

## Kubernetes

See `infra/k8s/` for manifests. Quick deploy:

```bash
# Create secrets
kubectl create secret generic physicscopilot-secrets \
  --from-literal=GEMINI_API_KEY=... \
  --from-literal=SUPABASE_URL=... \
  --from-literal=SUPABASE_JWT_SECRET=... \
  --from-literal=SUPABASE_SERVICE_KEY=... \
  --from-literal=DATABASE_URL=...

# Apply manifests
kubectl apply -f infra/k8s/deployment.yaml
kubectl apply -f infra/k8s/service.yaml

# Check rollout
kubectl rollout status deployment/physicscopilot-server
```

---

## Rollback

```bash
# Render: redeploy a previous deploy from the dashboard
# Docker Compose:
docker compose down && docker compose up -d --build

# Kubernetes:
kubectl rollout undo deployment/physicscopilot-server
```

---

## Monitoring

See `docs/MONITORING.md` for Prometheus + Grafana setup.
