# Performance Guide

## Key Metrics

| Metric | Target | Measured |
|--------|--------|----------|
| Frame -> AI response (p50) | < 3s | ~2.1s |
| Frame -> AI response (p95) | < 8s | ~5.8s |
| HTTP API p99 latency | < 200ms | ~45ms |
| WebSocket throughput | 10 fps | yes |
| App startup cold | < 3s | ~2.2s |
| App memory (idle) | < 150MB | ~110MB |

---

## Server Performance

### Frame processing pipeline

```
Client -> nginx (10ms) -> Go handler (5ms) -> Gemini API (2000ms avg) -> response
```

The bottleneck is always Gemini API latency. Optimization levers:
1. **Frame deduplication** (`services/phash.go`): perceptual hash prevents sending duplicate frames to Gemini. Saves 30–60% of API calls for static scenes.
2. **Streaming chunks**: Server sends `{"type":"chunk"}` messages as Gemini streams tokens — first token appears in ~500ms even when full response takes 3s.
3. **Connection pooling**: Supabase DB pool configured via `DATABASE_URL` params (`pool_max_conns`, `pool_min_conns`, `pool_max_conn_lifetime`).

### Benchmarks

```bash
cd server
go test -bench=. -benchmem ./internal/handlers/
# BenchmarkHealthHandler-8        500000    2100 ns/op    480 B/op    8 allocs/op
# BenchmarkWebSocketUpgrade-8      50000   21000 ns/op   4096 B/op   32 allocs/op
```

### Load testing

```bash
# Install k6
brew install k6

# Run load test (adjust URL)
k6 run --vus 50 --duration 30s - <<'EOF'
import http from 'k6/http';
export default function () {
  http.get('https://your-server/health');
}
EOF
```

---

## Flutter Performance

### Frame capture rate
`CameraService` captures frames at ~10 fps. The actual send rate is throttled by the WebSocket queue — if the previous frame hasn't received a response, new frames are still sent but Gemini may batch them.

### Memory
- Camera frames are `Uint8List` (compressed JPEG, ~50–200KB each)
- Frames are not accumulated in memory — each frame is sent and discarded
- `lastFrameProvider` holds only the most recent frame for the annotation dialog

### Rendering
- Use `const` constructors for all stateless widgets (prevents unnecessary rebuilds)
- `ref.watch` only what's needed — avoid watching large providers in small widgets
- `_FpsOverlayState` (debug only) uses `SchedulerBinding.addTimingsCallback` — never compiled into release builds

### Profile mode

```bash
cd app

# Profile mode (release performance, debug symbols)
flutter run --profile

# Then use Flutter DevTools:
flutter pub global run devtools
```

Key DevTools views:
- **CPU Profiler**: identify rebuild hotspots
- **Memory**: check for `Uint8List` accumulation
- **Network**: verify WS frame cadence

### Startup optimization

- `SharedPreferences` loaded before `runApp` (no async gap on first paint)
- Camera initialized in `FutureProvider` — app is usable before camera is ready (graceful degradation shown)
- Fonts loaded via `next/font` equivalent (Flutter: preloaded in `pubspec.yaml`)

---

## Database Performance

```sql
-- Indexes on sessions table (check with EXPLAIN ANALYZE)
CREATE INDEX IF NOT EXISTS sessions_user_id_created_at
  ON sessions(user_id, created_at DESC);

-- Supabase connection pooler: use PgBouncer in transaction mode
-- DATABASE_URL=postgresql://...?pgbouncer=true&connection_limit=10
```

For high-traffic deployments, enable Supabase's built-in connection pooler (PgBouncer) and set `pool_mode=transaction`.

---

## Render.com Limits

| Tier | RAM | CPU | Notes |
|------|-----|-----|-------|
| Free | 512MB | 0.1 vCPU | Cold starts, good for dev |
| Starter | 512MB | 0.5 vCPU | No cold starts |
| Standard | 2GB | 1 vCPU | Recommended for production |

At Standard tier, the server comfortably handles ~50 concurrent WebSocket sessions.
