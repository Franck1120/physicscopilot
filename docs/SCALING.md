# Horizontal Scaling Guide

This document covers how to scale PhysicsCopilot beyond a single instance — from a single Render.com container to a multi-replica deployment behind a load balancer.

---

## Architecture Overview

The Go server is **stateless by design**:

- No in-process session state survives a restart (all state lives in Postgres/Supabase when `DATABASE_URL` is set).
- The knowledge base (`kb/data/problems.json`) is bundled in the Docker image — all replicas have the same copy.
- Prometheus metrics are per-process; aggregate them in your Prometheus scrape config.

The only exception to statelessness is the **WebSocket connection**: an active WS connection is tied to a specific process. This is handled by session affinity at the load balancer layer (see below).

```
Internet
    │
    ▼
┌─────────────────────────────────┐
│  Load Balancer (nginx / Caddy)  │
│  - TLS termination              │
│  - WebSocket sticky sessions    │
└──────┬────────────┬─────────────┘
       │            │
       ▼            ▼
  ┌─────────┐  ┌─────────┐
  │ Server  │  │ Server  │  ... N replicas
  │ :8080   │  │ :8080   │
  └────┬────┘  └────┬────┘
       │             │
       └──────┬──────┘
              ▼
    ┌──────────────────┐
    │  Supabase Postgres│
    │  (shared state)  │
    └──────────────────┘
```

---

## Session Affinity (WebSocket Stickiness)

WebSocket connections are long-lived. Once a client upgrades to a WebSocket connection, all subsequent messages must go to the same server process. Without sticky sessions, a load balancer round-robins messages to different servers, breaking the connection.

### nginx — ip_hash (simple stickiness)

```nginx
upstream physicscopilot {
    ip_hash;                              # sticky: same client IP → same upstream
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    server 10.0.0.3:8080;
    keepalive 32;
}

server {
    listen 443 ssl;
    server_name physicscopilot.app;

    location /ws {
        proxy_pass         http://physicscopilot;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_read_timeout 3600s;         # keep WS alive for 1 hour max
        proxy_send_timeout 3600s;
    }

    location / {
        proxy_pass http://physicscopilot;
    }
}
```

### nginx — cookie-based stickiness (more robust)

Cookie-based stickiness works even when client IPs change (mobile networks, CGNAT):

```nginx
upstream physicscopilot {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    sticky cookie srv_id expires=1h domain=physicscopilot.app path=/ws;
}
```

Requires the `nginx-sticky-module-ng` module or nginx Plus.

### Caddy

Caddy's standard `reverse_proxy` does not support sticky sessions natively. Use the `lb_policy cookie` option (Caddy v2.7+):

```caddyfile
physicscopilot.app {
    reverse_proxy /ws* 10.0.0.1:8080 10.0.0.2:8080 {
        lb_policy cookie sticky_session
        header_up Upgrade {http.upgrade}
        header_up Connection "Upgrade"
    }

    reverse_proxy * 10.0.0.1:8080 10.0.0.2:8080
}
```

---

## Redis Session Store (Planned — v0.3.0)

Currently, WebSocket session state (in-flight Gemini requests, per-connection rate limiter) is in-process. This means:

- A server crash drops all active sessions.
- True horizontal scaling (any server can handle any request) is not possible until session state is externalized.

**Planned architecture with Redis:**

```
Client ──► Load Balancer ──► Any Server replica
                                    │
                                    ▼
                              Redis (session state)
                              Postgres (message history)
```

When this ships, the `SESSION_STORE` env var will accept `memory` (default) or `redis`.

---

## Connection Limits and Tuning

Each server process enforces a **per-IP connection limit of 10** (configurable via `MAX_CONNECTIONS_PER_IP` env var). Across N replicas, the effective limit is `10 × N` per IP if sticky sessions are disabled, or still 10 if sticky sessions route all connections from one IP to the same replica.

Set the per-process limit based on your Gemini API quota:

| Replicas | Per-IP limit | Effective max WS connections |
|----------|-------------|------------------------------|
| 1 | 10 | 10 |
| 3 | 10 | 30 (with ip_hash LB) |
| 10 | 5 | 50 |

Keep the Gemini API concurrent request budget in mind: each open WS connection may generate up to 5 Gemini requests/sec (frame rate limit). Plan replicas accordingly.

**OS-level tuning for high connection counts:**

```bash
# Increase file descriptor limits (each WS connection uses 1 FD)
echo "fs.file-max = 1000000" >> /etc/sysctl.conf
sysctl -p

# Per-process limit (add to /etc/security/limits.conf)
www-data soft nofile 65536
www-data hard nofile 65536
```

---

## Database Connection Pooling (pgxpool)

The server uses `pgxpool` from `pgx` for connection pooling. The pool is configured via environment variables:

| Environment variable | Default | Description |
|---------------------|---------|-------------|
| `DB_POOL_MAX_CONNS` | `10` | Maximum open connections per process |
| `DB_POOL_MIN_CONNS` | `2` | Minimum idle connections |
| `DB_POOL_MAX_CONN_LIFETIME` | `1h` | Max age of a single connection |
| `DB_POOL_MAX_CONN_IDLE_TIME` | `30m` | Evict idle connections after this time |
| `DB_POOL_HEALTH_CHECK_PERIOD` | `1m` | Frequency of health checks |

**Scaling formula:**

Supabase connection limits vary by plan. With N replicas each using `DB_POOL_MAX_CONNS` connections, total Postgres connections = `N × DB_POOL_MAX_CONNS`. Plan accordingly:

- Supabase Free: 60 connections max → `N × 10` replicas = max 6 replicas
- Supabase Pro: 200 connections max → max 20 replicas at 10 conns/replica
- Supabase Team: 500 connections max

For larger deployments, front Postgres with **PgBouncer** in transaction pooling mode (Supabase supports this out of the box — use the pooler connection string instead of the direct one).

---

## Prometheus Metrics for Scale Decisions

The server exposes metrics at `GET /metrics`. Key metrics for scaling decisions:

| Metric | Scale up when... |
|--------|-----------------|
| `ws_active_connections` | Approaching `MAX_CONNECTIONS_PER_IP × replicas` |
| `ai_inference_duration_seconds_p99` | > 3 s (Gemini is saturated or rate-limited) |
| `http_request_duration_seconds_p99` | > 200 ms for REST endpoints |
| `process_resident_memory_bytes` | > 80% of `GOMEMLIMIT` |
| `go_goroutines` | > 10k (goroutine leak check) |

Example Prometheus alerting rule:

```yaml
groups:
  - name: physicscopilot
    rules:
      - alert: HighWebSocketConnections
        expr: ws_active_connections > 40
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "WS connections high — consider adding replicas"

      - alert: GeminiLatencyHigh
        expr: histogram_quantile(0.99, rate(ai_inference_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Gemini p99 latency > 5s"
```

---

## Load Balancer Configuration

### nginx — full production config

```nginx
upstream physicscopilot_rest {
    least_conn;
    server 10.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 10.0.0.2:8080 max_fails=3 fail_timeout=30s;
    keepalive 64;
}

upstream physicscopilot_ws {
    ip_hash;
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
}

server {
    listen 443 ssl http2;
    server_name physicscopilot.app;

    ssl_certificate     /etc/ssl/certs/physicscopilot.crt;
    ssl_certificate_key /etc/ssl/private/physicscopilot.key;
    ssl_protocols       TLSv1.2 TLSv1.3;

    # WebSocket — sticky
    location /ws {
        proxy_pass         http://physicscopilot_ws;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # REST — round-robin
    location / {
        proxy_pass         http://physicscopilot_rest;
        proxy_http_version 1.1;
        proxy_set_header   Host $host;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 30s;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    return 301 https://$host$request_uri;
}
```

### Caddy — simpler alternative

```caddyfile
physicscopilot.app {
    # WebSocket sticky routing
    @ws path /ws
    reverse_proxy @ws 10.0.0.1:8080 10.0.0.2:8080 {
        lb_policy cookie sticky_ws
        transport http {
            dial_timeout 5s
        }
    }

    # REST round-robin
    reverse_proxy * 10.0.0.1:8080 10.0.0.2:8080 {
        lb_policy least_conn
        health_uri /health
        health_interval 10s
        health_timeout 5s
    }
}
```

---

## Kubernetes HPA Configuration

The `infra/k8s/` directory contains base manifests. Here is a HPA configuration for CPU + memory triggers:

```yaml
# infra/k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: physicscopilot-server
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: physicscopilot-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70   # scale up when CPU > 70% across pods
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 75   # scale up when memory > 75%
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60    # wait 60s before scaling up again
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60             # add max 2 pods per minute
    scaleDown:
      stabilizationWindowSeconds: 300   # wait 5 min before scaling down
      policies:
        - type: Pods
          value: 1
          periodSeconds: 120            # remove max 1 pod every 2 minutes
```

```yaml
# infra/k8s/deployment.yaml excerpt — resource requests/limits
resources:
  requests:
    cpu: "250m"
    memory: "128Mi"
  limits:
    cpu: "1000m"
    memory: "512Mi"
env:
  - name: GOMEMLIMIT
    value: "400MiB"   # must be below the memory limit
```

Apply with:

```bash
kubectl apply -f infra/k8s/hpa.yaml
kubectl get hpa physicscopilot-server --watch
```

---

## Cost Optimization Tips

1. **Right-size replicas.** The Go server is very memory-efficient (typically 18–40 MB heap). Don't over-provision — a single `t3.micro` handles hundreds of idle WS connections.

2. **Scale down during off-hours.** Use K8s CronJobs or Render's scheduled scaling to reduce replicas at night:
   ```bash
   kubectl scale deployment physicscopilot-server --replicas=1 # night
   kubectl scale deployment physicscopilot-server --replicas=4 # day
   ```

3. **Cache `/api/domains` aggressively.** The KB domain list does not change between deploys. Add a `Cache-Control: public, max-age=3600` header and let a CDN serve it.

4. **Gemini is the dominant cost.** Each frame analysis call costs money. The server already enforces a 5 fps cap per connection. Consider reducing this to 2 fps for general use (`FRAME_RATE_LIMIT=2`).

5. **Use `GOMEMLIMIT`.** Set `GOMEMLIMIT` to ~80% of your container memory limit. This prevents the Go GC from letting the heap grow beyond what the container allows, reducing OOM kills and costly restarts.

6. **Consolidate Prometheus scraping.** Use the Prometheus `remote_write` endpoint to send metrics to Grafana Cloud rather than running your own Prometheus. Eliminates the cost of a Prometheus VM.
