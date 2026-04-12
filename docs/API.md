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

**curl example**

```bash
# Local development
curl -s http://localhost:8080/health | jq .

# Production (replace with your tunnel/domain)
curl -s https://physicscopilot.app/health | jq .
```

**Expected output**
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

**curl example**

```bash
# Fetch raw Prometheus metrics
curl -s http://localhost:8080/metrics | grep -E "^(ws_|ai_inference|http_requests_total)"
```

**Sample output**
```
http_requests_total{method="GET",path="/health",status="200"} 142
ws_active_connections 3
ws_messages_total{type="frame"} 87
ws_messages_total{type="text"} 12
ai_inference_duration_seconds_sum 43.2
```

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
    "regions": [
      {
        "x": 0.32,
        "y": 0.45,
        "width": 0.15,
        "height": 0.10,
        "label": "Idler arm",
        "severity": "warning"
      }
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
| `overlay` | object | Optional. AR overlay data (regions to highlight on camera feed) |
| `overlay.regions[]` | array | Normalised bounding boxes (0.0–1.0) with labels |
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

## WebSocket Payload Examples

### Complete session transcript

Below is a real JSON exchange for a typical repair session. Each entry shows direction, type, and full payload.

**1. Client sends a frame for analysis**
```json
{
  "type": "frame",
  "data": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8U...",
  "timestamp": 1712870400123
}
```

**2. Server responds with AI guidance**
```json
{
  "type": "response",
  "text": "I can see the extruder assembly. The PTFE tube appears slightly loose at the hotend coupling. Press it firmly until it clicks, then tighten the blue collet clip.",
  "overlay": {
    "boxes": [
      { "x": 0.32, "y": 0.45, "w": 0.15, "h": 0.10, "label": "PTFE coupling" }
    ],
    "arrows": [
      { "x1": 0.50, "y1": 0.30, "x2": 0.39, "y2": 0.48 }
    ]
  },
  "step": { "current": 2, "total": 5 }
}
```

**3. Client sends a follow-up text question**
```json
{
  "type": "text",
  "content": "The PTFE tube is already seated. What else could cause the clicking?",
  "timestamp": 1712870425000
}
```

**4. Server responds**
```json
{
  "type": "response",
  "text": "If the PTFE is seated correctly, clicking usually means the extruder idler arm tension is too high. Loosen the spring screw by half a turn and test.",
  "overlay": {
    "boxes": [],
    "arrows": []
  },
  "step": { "current": 2, "total": 5 }
}
```

**5. Client sends application-level ping**
```json
{ "type": "ping", "timestamp": 1712870430000 }
```

**6. Server responds with pong**
```json
{ "type": "pong" }
```

**7. Server sends an error (e.g., AI timeout)**
```json
{
  "type": "error",
  "error": "Request timed out, please retry"
}
```

> Note: Error messages are intentionally generic. Internal details (stack traces, API status codes) are logged server-side only and never sent to the client.

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

### HTTP Error Codes

| Code | Meaning | When it occurs |
|------|---------|----------------|
| `200 OK` | Success | Normal REST response |
| `426 Upgrade Required` | WebSocket upgrade missing | HTTP GET to `/ws` without `Upgrade: websocket` header |
| `429 Too Many Requests` | Rate limit exceeded | > 60 REST requests/min from same IP |
| `500 Internal Server Error` | Unexpected error | Unhandled server panic; body: `{"error":"..."}` |

### WebSocket Close Codes

| Code | Name | Meaning |
|------|------|---------|
| `1000` | Normal Closure | Client or server closed the connection cleanly |
| `1001` | Going Away | Server is shutting down (`server shutting down`) |
| `1008` | Policy Violation | Connection limit reached for your IP (`connection limit reached`) |
| `1009` | Message Too Big | Frame exceeded the 10 MB limit |
| `1011` | Internal Error | Server-side panic — reconnect and retry |

### Application-Level Error Messages

Error messages sent via `{"type":"error","error":"..."}` are generic by design:

| Error text | Root cause |
|-----------|-----------|
| `Service temporarily unavailable, please retry in a few seconds` | Gemini API rate-limited (HTTP 429) or quota exhausted |
| `Request timed out, please retry` | Gemini took > 30 s to respond |
| `Request was cancelled` | Client disconnected during AI processing |
| `Invalid frame: unsupported image format: expected JPEG or PNG` | Frame base64 decodes to non-image data |
| `Invalid frame: invalid base64 encoding` | Malformed base64 in the `data` field |
| `Message content cannot be empty` | Empty `content` field in a `text` message |
| `Message content exceeds maximum allowed length` | `content` field > 50 KB |
| `Internal server error` | All other unexpected errors |

---

## Authentication

Authentication is currently **network-level only** (Cloudflare tunnel ACL, firewall, VPN). No JWT or API key is required on the WebSocket endpoint.

### Planned: JWT Flow (v0.2.0)

The planned authentication model will work as follows:

1. **Obtain a token** — POST to `/auth/token` with your credentials:

```bash
curl -s -X POST https://physicscopilot.app/auth/token \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"your-password"}' \
  | jq .
```

Expected response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

2. **Use the token** — Pass it as a query parameter on the WebSocket upgrade:

```
wss://physicscopilot.app/ws?token=eyJhbGci...
```

Or (once supported) as a header during the HTTP upgrade handshake:

```
Authorization: Bearer eyJhbGci...
```

> **Current status:** All endpoints are open. Implement network-level ACLs in production until v0.2.0 ships.
