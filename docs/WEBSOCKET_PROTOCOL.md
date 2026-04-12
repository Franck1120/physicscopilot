# WebSocket Protocol — PhysicsCopilot

This document describes the full WebSocket protocol used between the Flutter app (client) and the Go server.

---

## Connection

```
ws://host/ws?token=<JWT>[&lang=<BCP-47>]
wss://host/ws?token=<JWT>[&lang=<BCP-47>]
```

| Parameter | Required | Description |
|-----------|----------|-------------|
| `token`   | Yes (production) | Supabase JWT for authentication. Skipped in dev mode when `SUPABASE_JWT_SECRET` is not set. |
| `lang`    | No | BCP-47 language code for AI responses. Defaults to `it`. |

The server performs a standard HTTP → WebSocket upgrade (`101 Switching Protocols`).
If the JWT is missing or invalid in production mode, the connection is rejected with `401`.

---

## Limits

| Parameter | Value | Notes |
|-----------|-------|-------|
| Max connections per IP | 10 | Excess connections receive close code `1008` |
| Max message size | 10 MB | Server closes with `1009` on overflow |
| Frame rate cap | 5 fps (default) | Configurable via `WS_MAX_FPS` env var (range: 1–30) |
| Per-user API budget | server-side rate limit | Excess returns an `error` message (no disconnect) |

---

## Heartbeat

The server sends a WebSocket **Ping** control frame every **30 seconds**.

The client **must** respond with a **Pong** control frame within **40 seconds** (10 s grace period).
If no Pong is received within the deadline the server closes the connection.

Clients that cannot intercept raw control frames (e.g. browser `WebSocket` API) may send an
application-level `{"type":"ping"}` message instead — the server responds with `{"type":"pong"}`.

```
Server                     Client
  │── Ping (control) ────►  │
  │◄── Pong (control) ────  │   (within 40 s)
  │
  │  OR (application-level fallback):
  │◄── {"type":"ping"} ───  │
  │── {"type":"pong"} ────► │
```

---

## Frame types: Client → Server

All messages are JSON text frames unless noted.

### `frame` — Camera frame

Sends a JPEG image for real-time AI analysis.

```json
{
  "type": "frame",
  "data": "<base64-encoded JPEG>",
  "timestamp": 1744459200000
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"frame"` |
| `data` | string | Base64-encoded JPEG. Must begin with JPEG magic bytes `FF D8 FF`. |
| `timestamp` | integer | Unix timestamp in milliseconds (optional, for client-side latency tracking). |

Frame deduplication: identical consecutive frames (by SHA-256 hash of the first 1024 bytes) are
silently dropped to avoid redundant AI calls. Hashes expire after 30 minutes.

Frames exceeding the FPS cap (default 5/sec) are silently dropped — no error is sent.

---

### `text` — Text message

Sends a user text message for a text-only conversation turn (no camera frame).

```json
{
  "type": "text",
  "content": "The extruder keeps clogging after 10 minutes.",
  "timestamp": 1744459200000
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"text"` |
| `content` | string | User message. Max 5000 bytes. HTML tags are stripped before forwarding to AI. |
| `timestamp` | integer | Unix timestamp in milliseconds (optional). |

---

### `ping` — Application-level keepalive

```json
{ "type": "ping" }
```

Server responds immediately with `{"type":"pong"}`. Use when the runtime does not expose
WebSocket control frames directly.

---

## Frame types: Server → Client

All server messages are JSON text frames.

### `response` — AI analysis result

Returned after processing a `frame` or `text` message.

```json
{
  "type": "response",
  "text": "**Analysis:** Stringing detected near the extruder nozzle.\n\n**Instruction:** Increase retraction distance to 6 mm and reduce travel speed.",
  "voice_text": "Stringing detected near the extruder nozzle. Increase retraction distance to 6 mm and reduce travel speed.",
  "overlay": {
    "boxes": [
      { "x": 0.32, "y": 0.45, "w": 0.18, "h": 0.12, "label": "Stringing" }
    ],
    "arrows": [
      { "x1": 0.40, "y1": 0.57, "x2": 0.40, "y2": 0.75 }
    ]
  },
  "step": {
    "current": 2,
    "total": 5
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"response"` |
| `text` | string | Full AI response with Markdown formatting (analysis + instruction). |
| `voice_text` | string | TTS-optimised instruction — Markdown stripped, whitespace collapsed. Omitted when empty. |
| `overlay` | object | Visual annotations to render over the camera feed. See [Overlay schema](#overlay-schema). |
| `step` | object | User's position in the guided repair flow. |
| `step.current` | integer | Current step index (0-based). |
| `step.total` | integer | Total steps in the current flow (0 when unknown). |

#### Overlay schema

```json
{
  "boxes": [
    { "x": 0.32, "y": 0.45, "w": 0.18, "h": 0.12, "label": "Stringing" }
  ],
  "arrows": [
    { "x1": 0.40, "y1": 0.57, "x2": 0.40, "y2": 0.75 }
  ]
}
```

All coordinates are **normalized (0.0–1.0)** relative to the camera frame dimensions.

| Field | Description |
|-------|-------------|
| `boxes[].x` | Left edge of bounding box |
| `boxes[].y` | Top edge of bounding box |
| `boxes[].w` | Width of bounding box |
| `boxes[].h` | Height of bounding box |
| `boxes[].label` | Human-readable annotation label |
| `arrows[].x1, y1` | Arrow start point |
| `arrows[].x2, y2` | Arrow end point (direction indicator) |

---

### `error` — Non-fatal error

Sent when a message cannot be processed. The connection remains open.

```json
{
  "type": "error",
  "error": "rate limit exceeded — slow down"
}
```

| Field | Description |
|-------|-------------|
| `type` | `"error"` |
| `error` | Human-readable error message. |

Common error messages:

| Message | Cause |
|---------|-------|
| `"rate limit exceeded — slow down"` | Per-user AI API budget exceeded |
| `"frame must be a base64-encoded JPEG image"` | Frame data is not a valid JPEG |
| `"message content is empty after sanitization"` | Text message was empty or HTML-only |

---

### `pong` — Keepalive response

Sent in response to an application-level `{"type":"ping"}` message.

```json
{ "type": "pong" }
```

---

## Close codes

| Code | Description |
|------|-------------|
| `1000` | Normal closure |
| `1001` | Server is shutting down (`"server shutting down"`) |
| `1008` | Connection limit reached for this IP (`"connection limit reached"`) |
| `1009` | Message too large (> 10 MB) |

---

## Authentication

JWT is passed as a query parameter on the upgrade request:

```
GET /ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... HTTP/1.1
Upgrade: websocket
```

Alternatively, the standard `Authorization: Bearer <token>` header is accepted.

When `SUPABASE_JWT_SECRET` is not set the server runs in **dev mode** — all connections are
accepted without a token and the authenticated user ID is empty.

---

## Reconnection strategy

There is no built-in server-side reconnection. Clients should implement exponential back-off:

| Attempt | Delay |
|---------|-------|
| 1 | 1 s |
| 2 | 2 s |
| 3 | 4 s |
| 4+ | cap at 30 s |

Re-authenticate with a fresh JWT on each reconnection attempt (tokens may expire during
disconnection). A new session is created automatically by the server on every new connection.

---

## Typical sequence

```
Client                             Server
  │── GET /ws?token=JWT ─────────► │
  │◄── 101 Switching Protocols ─── │  (session created)
  │                                │
  │── {"type":"frame","data":"..."} │
  │◄── {"type":"response", ...} ── │
  │                                │
  │── {"type":"text","content":"..."} │
  │◄── {"type":"response", ...} ── │
  │                                │
  │       (30 s later)             │
  │◄── Ping (control frame) ─────  │
  │── Pong (control frame) ──────► │
  │                                │
  │── {"type":"ping"} ──────────── │  (optional app-level)
  │◄── {"type":"pong"} ─────────── │
  │                                │
  │── Close (1000) ──────────────► │
  │◄── Close (1000) ─────────────  │
```

---

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SUPABASE_JWT_SECRET` | — | JWT signing secret. When absent, auth is skipped (dev mode). |
| `WS_MAX_FPS` | `5` | Per-connection frame rate cap. Clamped to `[1, 30]`. |
