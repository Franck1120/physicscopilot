# Monitoring Guide

## Overview

PhysicsCopilot exposes metrics in Prometheus format. The `infra/` directory contains ready-to-use Prometheus and Grafana configurations.

---

## Metrics Endpoint

The Go server exposes `GET /metrics` (Prometheus format).

Key metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `physicscopilot_active_ws_connections` | Gauge | Current WebSocket connections |
| `physicscopilot_frames_processed_total` | Counter | Total frames sent to Gemini |
| `physicscopilot_ai_latency_seconds` | Histogram | Time from frame receipt to response |
| `physicscopilot_errors_total` | Counter | Errors by type and handler |
| `physicscopilot_rate_limit_hits_total` | Counter | Rate limit events by user |
| `http_requests_total` | Counter | HTTP requests by method, path, status |
| `http_request_duration_seconds` | Histogram | HTTP request latency |

---

## Prometheus Setup (local)

```bash
# Start Prometheus + Grafana via Docker Compose
docker compose -f infra/prometheus/docker-compose.yml up -d

# Prometheus UI: http://localhost:9090
# Grafana:       http://localhost:3000 (admin/admin)
```

Prometheus config: `infra/prometheus/prometheus.yml`
Grafana dashboards: `infra/grafana/dashboards/`

---

## Grafana Dashboards

| Dashboard | Description |
|-----------|-------------|
| `physicscopilot-overview` | Active connections, frame throughput, AI latency |
| `physicscopilot-errors` | Error rate by type, rate limit events |
| `go-runtime` | Go memory, GC, goroutines |

Import dashboards from `infra/grafana/dashboards/*.json`.

---

## Alerts

Recommended Prometheus alert rules:

```yaml
# Add to infra/prometheus/alerts.yml
groups:
  - name: physicscopilot
    rules:
      - alert: HighAILatency
        expr: histogram_quantile(0.95, physicscopilot_ai_latency_seconds_bucket) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "AI response p95 latency > 10s"

      - alert: HighErrorRate
        expr: rate(physicscopilot_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Error rate > 0.1/s"

      - alert: ServerDown
        expr: up{job="physicscopilot"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "PhysicsCopilot server is down"
```

---

## Render.com Monitoring

Render provides built-in CPU, memory, and request metrics in the dashboard. No configuration needed.

- CPU and memory: Render dashboard → Service → Metrics
- Request logs: Render dashboard → Service → Logs
- Deploy notifications: Render dashboard → Settings → Notifications

---

## Application Logs

Structured JSON logs (production):

```json
{"time":"2026-04-12T10:00:00Z","level":"INFO","msg":"frame analyzed","user_id_hash":"a3f...","latency_ms":234,"tokens_used":1200}
{"time":"2026-04-12T10:00:05Z","level":"ERROR","msg":"gemini API error","error":"rate limit exceeded","retry_after":60}
```

Log levels:
- `DEBUG`: verbose, only in development
- `INFO`: normal operation events
- `WARN`: recoverable issues (reconnect, retry)
- `ERROR`: failures that affect users

Filter logs in production:
```bash
# On Render: use the Logs tab and filter by level
# Locally:
docker compose logs server | jq 'select(.level == "ERROR")'
```

---

## Health Check

```bash
curl https://your-server/health
# {"status":"ok","version":"0.17.0","uptime_seconds":86400,"active_connections":12}
```

Automated health monitoring: use [Uptime Robot](https://uptimerobot.com) or [Better Uptime](https://betteruptime.com) with a 1-minute check interval on `/health`.
