# Horizontal Scaling Guide

This document describes how to scale the PhysicsCopilot server horizontally
across multiple instances without requiring application changes.

## Stateless Design

The server is designed to be stateless at the HTTP layer:

- **In-memory session store** — sessions are kept in RAM by default. When
  `DATABASE_URL` is set, sessions are persisted to Postgres and hydrated on
  startup, enabling any instance to serve any client.
- **No sticky sessions required for REST** — all REST endpoints are fully
  stateless once `DATABASE_URL` is configured.
- **JWT authentication** — tokens are verified locally using `SUPABASE_JWT_SECRET`;
  no shared auth state is needed.

## WebSocket and Load Balancing

WebSocket connections maintain long-lived state on the instance that accepted
the upgrade. Horizontal scaling with WebSockets therefore requires:

- **Sticky sessions** (also called _session affinity_) at the load balancer,
  so that a client always reconnects to the same instance.
- If sticky sessions are unavailable, a shared pub/sub backend (e.g. Redis
  Pub/Sub) can be added to forward WebSocket messages across instances — this
  is not included in the default build but the `WSHandler` interface is
  designed to be replaceable.

### Example: NGINX upstream with sticky sessions

```nginx
upstream physicscopilot {
    ip_hash;  # routes each client IP to the same backend
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}
```

## Database Connection Pool

Each instance maintains its own connection pool to Postgres. Tune the pool
size to avoid exhausting the database's `max_connections`:

| Environment variable  | Default | Description                               |
|-----------------------|---------|-------------------------------------------|
| `DB_POOL_MAX_CONNS`   | `50`    | Maximum open connections per instance     |
| `DB_POOL_MIN_CONNS`   | `5`     | Minimum idle connections per instance     |

With three instances and the defaults above you would consume up to 150
connections. Adjust `DB_POOL_MAX_CONNS` accordingly, or use PgBouncer in
transaction-pooling mode in front of Postgres.

## Rate Limiting with Multiple Instances

The built-in rate limiter (`middleware.NewIPRateLimiter`) runs in-process and
does not share state across instances. In a multi-instance deployment:

- Each instance enforces its own per-IP rate limit independently.
- For stricter global rate limiting, place a shared limiter at the load balancer
  layer (e.g. NGINX `limit_req`, AWS WAF, or Cloudflare Rate Limiting).

## Prometheus Metrics Aggregation

Each instance exposes `/metrics` independently. To aggregate across instances:

1. Configure Prometheus to scrape every instance:

   ```yaml
   scrape_configs:
     - job_name: physicscopilot
       static_configs:
         - targets:
           - backend1:8080
           - backend2:8080
           - backend3:8080
   ```

2. Use `sum by (method, path, status)(rate(http_requests_total[5m]))` in
   Grafana to see aggregate request rates.

## Docker / Container Deployment

The provided `Dockerfile` builds a minimal container image. To run multiple
replicas:

```bash
# Build the image
docker build -t physicscopilot:latest .

# Run three replicas (no orchestrator)
for i in 1 2 3; do
  docker run -d \
    --name physicscopilot-$i \
    -e DATABASE_URL="$DATABASE_URL" \
    -e SUPABASE_JWT_SECRET="$SUPABASE_JWT_SECRET" \
    -e APP_ENV=production \
    -e PORT=8080 \
    -p $((8079 + i)):8080 \
    physicscopilot:latest
done
```

### Docker Compose (development / single-host)

```yaml
services:
  physicscopilot:
    build: .
    deploy:
      replicas: 3
    environment:
      DATABASE_URL: ${DATABASE_URL}
      SUPABASE_JWT_SECRET: ${SUPABASE_JWT_SECRET}
      APP_ENV: production
    ports:
      - "8080-8082:8080"
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: physicscopilot
spec:
  replicas: 3
  selector:
    matchLabels:
      app: physicscopilot
  template:
    spec:
      containers:
        - name: server
          image: physicscopilot:latest
          ports:
            - containerPort: 8080
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: physicscopilot-secrets
                  key: database-url
```

Use a `Service` with `sessionAffinity: ClientIP` to enable sticky sessions for
WebSocket clients.

## Health Checks

All orchestrators can use `GET /health` for liveness and readiness probes. The
endpoint responds in under 50 ms under normal load and returns HTTP 200 with a
JSON body including `db_status` and `active_connections`.
