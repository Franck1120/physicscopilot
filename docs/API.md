# PhysicsCopilot — API Reference

**Base URL (local dev):** `http://localhost:8080`  
**Base URL (production):** Set by `CLIPROXY_URL` / tunnel URL in the Flutter app constants.

---

## REST Endpoints

### `GET /health`

Returns server liveness and resource snapshot. Rate-limited to **60 requests/min per IP**.

**Response `200 OK`**

```json
{
  "status": "ok",
  "service": "physicscopilot",
  "version": "0.1.0",
  "uptime": "3h14m22s",
  "active_connections": 2,
  "memory_mb": 18
}
```

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always `"ok"` when the server is healthy |
| `version` | string | Server build version |
| `uptime` | string | Time since server start (`XhYmZs` format) |
| `active_connections` | int | Current open WebSocket connections |
| `memory_mb` | int | Heap memory allocated in MB (`runtime.MemStats.Alloc`) |

---

### `GET /metrics`

Exposes **Prometheus metrics** in the standard text format. Suitable for scraping by Prometheus or Grafana Agent.

**No authentication required** — restrict at the network/reverse-proxy level in production.

**Example output (excerpt)**

```
# HELP http_requests_total Total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/health",status="200"} 142

# HELP ws_active_connections Number of currently active WebSocket connections.
# TYPE ws_active_connections gauge
ws_active_connections 3

# HELP ai_inference_duration_seconds Time spent waiting for AI inference (Gemini) to respond.
# TYPE ai_inference_duration_seconds histogram
ai_inference_duration_seconds_bucket{le="0.5"} 18
ai_inference_duration_seconds_bucket{le="1"} 31
```

**Registered metrics**

| Name | Type | Labels | Description |
|------|------|--------|-------------|
| `http_requests_total` | Counter | `method`, `path`, `status` | Every HTTP request served |
| `http_request_duration_seconds` | Histogram | `method`, `path` | Request latency (DefBuckets) |
| `ws_active_connections` | Gauge | — | Open WebSocket connections |
| `ws_messages_total` | Counter | `type` | Incoming WS messages by type |
| `ai_inference_duration_seconds` | Histogram | — | Gemini round-trip latency (0.1–10 s) |

---

## WebSocket Protocol

**Endpoint:** `ws://host:8080/ws` (local) / `wss://host/ws` (production / Cloudflare tunnel)

The server enforces:
- **Max 10 concurrent connections per source IP** (returns `CloseMessage 1008 — connection limit reached`)
- **Max message size: 10 MB** per frame
- **Frame rate limit: 5 camera frames / second** per connection (excess silently dropped)
- **Server heartbeat:** Ping every 30 s; connection closed if no Pong within 40 s

---

### Client → Server messages (JSON)

All messages are UTF-8 JSON objects.

#### `frame` — Camera frame

Sends a base64-encoded JPEG camera snapshot for AI analysis.

```json
{
  "type": "frame",
  "data": "<base64-encoded JPEG>",
  "timestamp": 1712870400000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✅ | Must be `"frame"` |
| `data` | string | ✅ | Base64-encoded JPEG image |
| `timestamp` | int | ✅ | Unix timestamp in milliseconds |

---

#### `text` — User text message

Sends a natural-language question or command (voice-to-text or typed input).

```json
{
  "type": "text",
  "content": "The extruder keeps clicking — what should I check?",
  "timestamp": 1712870401000
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | ✅ | Must be `"text"` |
| `content` | string | ✅ | User message text |
| `timestamp` | int | ✅ | Unix timestamp in milliseconds |

---

#### `ping` — Application-level ping

Keepalive at the application layer (distinct from WebSocket protocol Ping frames).

```json
{
  "type": "ping",
  "timestamp": 1712870402000
}
```

---

### Server → Client messages (JSON)

#### `response` — AI guidance

Sent after a `frame` or `text` message has been processed by Gemini.

```json
{
  "type": "response",
  "text": "Your extruder is clicking because of filament grinding. Check that the idler arm tension is not too tight and that the filament diameter is consistent.",
  "overlay": {
    "boxes": [
      {
        "x": 0.32,
        "y": 0.45,
        "w": 0.15,
        "h": 0.10,
        "label": "Idler arm"
      }
    ],
    "arrows": [
      { "x1": 0.40, "y1": 0.50, "x2": 0.55, "y2": 0.60 }
    ]
  },
  "step": {
    "index": 2,
    "total": 5,
    "description": "Loosen the idler arm tensioner by half a turn",
    "completed": false
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"response"` |
| `text` | string | Human-readable AI guidance text (spoken aloud via TTS) |
| `overlay` | object | Optional. AR overlay data rendered on the camera feed |
| `overlay.boxes[]` | array | Bounding boxes in normalised coords (0.0–1.0): `x`, `y`, `w`, `h`, `label` |
| `overlay.arrows[]` | array | Directional arrows in normalised coords: `x1`, `y1`, `x2`, `y2` |
| `step` | object | Optional. Current repair procedure step |
| `step.index` | int | Current step (0-based) |
| `step.total` | int | Total steps in the procedure |
| `step.description` | string | Step instruction |
| `step.completed` | bool | Whether this step is marked done |

---

#### `pong` — Application-level pong

Response to a client `ping` message.

```json
{
  "type": "pong"
}
```

---

#### `error` — Error from server

Sent when frame or text processing fails. **The connection remains open.**

```json
{
  "type": "error",
  "error": "Gemini inference failed: context deadline exceeded"
}
```

---

## Sequence Diagrams

### Normal camera analysis session

```
Client                                Server
  │                                     │
  │──── WS Upgrade (GET /ws) ─────────►│
  │◄─── 101 Switching Protocols ────────│
  │                                     │
  │──── {"type":"frame","data":"..."} ──►│
  │                        ProcessFrame │ (Gemini API call ~500ms–3s)
  │◄─── {"type":"response",...} ────────│
  │                                     │
  │──── {"type":"frame","data":"..."} ──►│   (5 fps max)
  │◄─── {"type":"response",...} ────────│
  │                                     │
  │──── {"type":"text","content":"?"} ──►│
  │                     ProcessText     │ (Gemini API call)
  │◄─── {"type":"response",...} ────────│
  │                                     │
  │                     [30s elapsed]   │
  │◄─── WS Ping frame ──────────────────│
  │──── WS Pong frame ─────────────────►│   (read deadline reset to now+40s)
  │                                     │
  │──── WS Close (GoingAway) ──────────►│   (client disconnect)
  │◄─── WS Close ───────────────────────│
  │                                     │
```

### Graceful server shutdown

```
Client                                Server
  │                                     │
  │                     [SIGTERM recv]  │
  │◄─── WS Close (GoingAway, "server shutting down") ─│
  │──── WS Close ACK ──────────────────►│   (500ms grace period)
  │                                     │ app.ShutdownWithTimeout(10s)
  │                                     │ [server exits]
```

### Connection limit exceeded

```
Client                                Server
  │                                     │
  │──── WS Upgrade (11th conn from IP) ►│
  │◄─── WS Close (1008 Policy Violation, "connection limit reached") ─│
  │                                     │
```

---

## Error Codes

| HTTP / WS Code | Meaning |
|----------------|---------|
| `HTTP 429` | REST rate limit exceeded (60 req/min per IP) |
| `HTTP 426` | Upgrade Required (non-WS request to `/ws`) |
| `HTTP 500` | Internal server error (JSON body with `"error"` field) |
| `WS 1000` | Normal closure |
| `WS 1001` | Going Away (server shutdown) |
| `WS 1008` | Policy Violation (connection limit, message too large) |
| `WS 1009` | Message Too Big (frame exceeds 10 MB) |

---

## Authentication

There is currently **no authentication** on the WebSocket endpoint. All access controls are network-level (firewall, Cloudflare tunnel ACL, VPN).

A session token flow is planned for v0.2.0.
