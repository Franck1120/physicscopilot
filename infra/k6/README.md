# k6 Load Tests — PhysicsCopilot

This directory contains [k6](https://k6.io) performance tests for the PhysicsCopilot API.

## Prerequisites

Install k6:
```bash
# macOS
brew install k6

# Linux
sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update && sudo apt-get install k6

# Windows (Chocolatey)
choco install k6
```

## Available Tests

| Script | Purpose | Duration |
|--------|---------|---------|
| `load_test.js` | Steady-state load (10 users, 2 min) | ~2 min |
| `spike_test.js` | Traffic spike (0→100 users) | ~70 sec |
| `soak_test.js` | Extended load (5 users, 30 min) | ~34 min |

## Running Tests

```bash
# Against local server (default)
k6 run infra/k6/load_test.js

# Against staging
k6 run --env BASE_URL=https://staging.physicscopilot.app infra/k6/spike_test.js

# Short soak test for CI (2 minutes instead of 30)
k6 run --env DURATION=2m infra/k6/soak_test.js
```

## Thresholds

Tests fail if:
- **load_test**: p95 health latency > 500ms or error rate > 1%
- **spike_test**: error rate > 5% during spike
- **soak_test**: p95 latency > 1s or error rate > 1%

## Authentication

Most `/api/*` endpoints require a JWT. Without a valid `SUPABASE_JWT_SECRET` the server runs in dev mode (no auth). For authenticated load tests:

```bash
k6 run --env JWT_TOKEN=<your-dev-token> infra/k6/load_test.js
```
