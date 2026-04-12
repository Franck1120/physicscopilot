# PhysicsCopilot — API Examples

Practical code examples for every endpoint. All examples assume the server is running at `http://localhost:8080` for local development and `https://physicscopilot-server.onrender.com` for production.

---

## Table of Contents

- [GET /health](#get-health)
- [POST /api/sessions](#post-apisessions)
- [GET /api/sessions](#get-apisessions)
- [GET /api/sessions/:id](#get-apisessionsid)
- [DELETE /api/sessions/:id](#delete-apisessionsid)
- [POST /api/feedback](#post-apifeedback)
- [GET /api/domains](#get-apidomains)
- [WS /ws](#ws-ws)

---

## GET /health

Returns server liveness, version, uptime, active connections, and memory usage. Rate-limited to 60 requests/min per IP.

### curl

```bash
# Local development
curl -s http://localhost:8080/health | jq .

# Production
curl -s https://physicscopilot-server.onrender.com/health | jq .
```

**Expected response (200 OK)**

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

### JavaScript (fetch)

```js
const res = await fetch('http://localhost:8080/health');
const health = await res.json();
console.log(`Server ${health.status}, uptime: ${health.uptime}`);
// Server ok, uptime: 3h14m22s
```

### Python (requests)

```python
import requests

resp = requests.get('http://localhost:8080/health', timeout=5)
resp.raise_for_status()
data = resp.json()
print(f"Status: {data['status']}, memory: {data['memory_mb']} MB")
# Status: ok, memory: 18 MB
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type HealthResponse struct {
    Status            string `json:"status"`
    Service           string `json:"service"`
    Version           string `json:"version"`
    Uptime            string `json:"uptime"`
    ActiveConnections int    `json:"active_connections"`
    MemoryMB          int    `json:"memory_mb"`
}

func main() {
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var h HealthResponse
    json.Unmarshal(body, &h)
    fmt.Printf("Status: %s, version: %s\n", h.Status, h.Version)
}
```

---

## POST /api/sessions

Creates a new session. Returns the session object including an `id` used for subsequent requests.

**Request headers**

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |
| `Authorization` | `Bearer <token>` (required in production once JWT is enabled) |

**Request body**

```json
{
  "domain": "printer",
  "device_id": "flutter-device-abc123"
}
```

### curl

```bash
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"domain":"printer","device_id":"flutter-device-abc123"}' \
  | jq .
```

**Expected response (201 Created)**

```json
{
  "id": "sess_01HXYZ1234ABCD",
  "domain": "printer",
  "device_id": "flutter-device-abc123",
  "created_at": "2026-04-12T14:30:00Z",
  "status": "active"
}
```

### JavaScript (fetch)

```js
const res = await fetch('http://localhost:8080/api/sessions', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    // 'Authorization': 'Bearer <token>',  // enable in production
  },
  body: JSON.stringify({
    domain: 'printer',
    device_id: 'js-client-001',
  }),
});

if (!res.ok) throw new Error(`HTTP ${res.status}`);
const session = await res.json();
console.log('Session created:', session.id);
```

### Python (requests)

```python
import requests

payload = {
    "domain": "laptop",
    "device_id": "python-client-001",
}

resp = requests.post(
    "http://localhost:8080/api/sessions",
    json=payload,
    headers={"Content-Type": "application/json"},
    timeout=10,
)
resp.raise_for_status()
session = resp.json()
print(f"Session ID: {session['id']}")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    body, _ := json.Marshal(map[string]string{
        "domain":    "automotive",
        "device_id": "go-client-001",
    })

    resp, err := http.Post(
        "http://localhost:8080/api/sessions",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    data, _ := io.ReadAll(resp.Body)
    fmt.Println(string(data))
}
```

---

## GET /api/sessions

Returns a paginated list of all sessions for the authenticated user/device.

**Query parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 20 | Maximum sessions to return (max 100) |
| `offset` | int | 0 | Pagination offset |
| `domain` | string | — | Filter by domain (e.g. `printer`) |

### curl

```bash
# List first 20 sessions
curl -s "http://localhost:8080/api/sessions" | jq .

# With pagination and domain filter
curl -s "http://localhost:8080/api/sessions?limit=10&offset=20&domain=laptop" | jq .
```

**Expected response (200 OK)**

```json
{
  "sessions": [
    {
      "id": "sess_01HXYZ1234ABCD",
      "domain": "printer",
      "device_id": "flutter-device-abc123",
      "created_at": "2026-04-12T14:30:00Z",
      "status": "active"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### JavaScript (fetch)

```js
const params = new URLSearchParams({ limit: '10', offset: '0' });
const res = await fetch(`http://localhost:8080/api/sessions?${params}`);
const { sessions, total } = await res.json();
console.log(`Fetched ${sessions.length} of ${total} sessions`);
```

### Python (requests)

```python
import requests

resp = requests.get(
    "http://localhost:8080/api/sessions",
    params={"limit": 10, "domain": "hvac"},
    timeout=10,
)
resp.raise_for_status()
data = resp.json()
for s in data["sessions"]:
    print(f"{s['id']} — {s['domain']} — {s['created_at']}")
```

---

## GET /api/sessions/:id

Retrieves a single session by its ID, including the full message history if persistence is enabled.

### curl

```bash
# Replace sess_01HXYZ1234ABCD with an actual session ID
curl -s http://localhost:8080/api/sessions/sess_01HXYZ1234ABCD | jq .
```

**Expected response (200 OK)**

```json
{
  "id": "sess_01HXYZ1234ABCD",
  "domain": "printer",
  "device_id": "flutter-device-abc123",
  "created_at": "2026-04-12T14:30:00Z",
  "status": "active",
  "messages": [
    {
      "role": "user",
      "content": "The extruder keeps clicking",
      "timestamp": "2026-04-12T14:31:00Z"
    },
    {
      "role": "assistant",
      "content": "Check the idler arm tension — it may be too tight.",
      "timestamp": "2026-04-12T14:31:02Z"
    }
  ]
}
```

**Error response (404 Not Found)**

```json
{
  "error": "session not found"
}
```

### JavaScript (fetch)

```js
const sessionId = 'sess_01HXYZ1234ABCD';
const res = await fetch(`http://localhost:8080/api/sessions/${sessionId}`);

if (res.status === 404) {
  console.error('Session not found');
} else {
  const session = await res.json();
  console.log(`Session has ${session.messages?.length ?? 0} messages`);
}
```

### Python (requests)

```python
import requests

session_id = "sess_01HXYZ1234ABCD"
resp = requests.get(
    f"http://localhost:8080/api/sessions/{session_id}",
    timeout=10,
)

if resp.status_code == 404:
    print("Session not found")
else:
    resp.raise_for_status()
    session = resp.json()
    print(f"Messages: {len(session.get('messages', []))}")
```

---

## DELETE /api/sessions/:id

Deletes a session and its associated message history.

### curl

```bash
curl -s -X DELETE http://localhost:8080/api/sessions/sess_01HXYZ1234ABCD

# With verbose output to confirm the 204 No Content response
curl -v -X DELETE http://localhost:8080/api/sessions/sess_01HXYZ1234ABCD
```

**Expected response (204 No Content)** — empty body.

**Error response (404 Not Found)**

```json
{
  "error": "session not found"
}
```

### JavaScript (fetch)

```js
const sessionId = 'sess_01HXYZ1234ABCD';
const res = await fetch(`http://localhost:8080/api/sessions/${sessionId}`, {
  method: 'DELETE',
});

if (res.status === 204) {
  console.log('Session deleted successfully');
} else if (res.status === 404) {
  console.error('Session not found');
} else {
  console.error(`Unexpected status: ${res.status}`);
}
```

### Python (requests)

```python
import requests

session_id = "sess_01HXYZ1234ABCD"
resp = requests.delete(
    f"http://localhost:8080/api/sessions/{session_id}",
    timeout=10,
)

if resp.status_code == 204:
    print("Deleted")
elif resp.status_code == 404:
    print("Not found")
else:
    resp.raise_for_status()
```

---

## POST /api/feedback

Submits user feedback for a session. Rating is 1–5 (1 = very poor, 5 = excellent). Comment is optional.

**Request headers**

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |

**Request body**

```json
{
  "session_id": "sess_01HXYZ1234ABCD",
  "rating": 5,
  "comment": "Solved the jam in under 2 minutes — incredibly accurate."
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `session_id` | string | Yes | Must be an existing session ID |
| `rating` | int | Yes | 1–5 inclusive |
| `comment` | string | No | Max 1000 characters |

### curl

```bash
curl -s -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "sess_01HXYZ1234ABCD",
    "rating": 4,
    "comment": "Very helpful, minor delay on second frame."
  }' | jq .
```

**Expected response (201 Created)**

```json
{
  "id": "fb_01HABC9876XYZ",
  "session_id": "sess_01HXYZ1234ABCD",
  "rating": 4,
  "comment": "Very helpful, minor delay on second frame.",
  "created_at": "2026-04-12T15:00:00Z"
}
```

**Validation error (400 Bad Request)**

```json
{
  "error": "rating must be between 1 and 5"
}
```

### JavaScript (fetch)

```js
const res = await fetch('http://localhost:8080/api/feedback', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    session_id: 'sess_01HXYZ1234ABCD',
    rating: 5,
    comment: 'Identified the fault immediately.',
  }),
});

const feedback = await res.json();
console.log('Feedback submitted:', feedback.id);
```

### Python (requests)

```python
import requests

resp = requests.post(
    "http://localhost:8080/api/feedback",
    json={
        "session_id": "sess_01HXYZ1234ABCD",
        "rating": 3,
        "comment": "Good but missed the cable routing step.",
    },
    timeout=10,
)
resp.raise_for_status()
print(f"Feedback ID: {resp.json()['id']}")
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type FeedbackRequest struct {
    SessionID string `json:"session_id"`
    Rating    int    `json:"rating"`
    Comment   string `json:"comment,omitempty"`
}

func main() {
    fb := FeedbackRequest{
        SessionID: "sess_01HXYZ1234ABCD",
        Rating:    5,
        Comment:   "Perfect guidance, fixed in 90 seconds.",
    }
    body, _ := json.Marshal(fb)

    resp, err := http.Post(
        "http://localhost:8080/api/feedback",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    data, _ := io.ReadAll(resp.Body)
    fmt.Println(string(data))
}
```

---

## GET /api/domains

Returns all available knowledge base domains with metadata (name, description, problem count).

### curl

```bash
curl -s http://localhost:8080/api/domains | jq .
```

**Expected response (200 OK)**

```json
{
  "domains": [
    {
      "id": "printer",
      "name": "Printer",
      "description": "Paper jams, connectivity issues, print quality problems",
      "problem_count": 24
    },
    {
      "id": "laptop",
      "name": "Laptop",
      "description": "Hardware faults, overheating, display and keyboard repairs",
      "problem_count": 31
    },
    {
      "id": "automotive",
      "name": "Automotive",
      "description": "Engine faults, electrical systems, brake and tyre guidance",
      "problem_count": 18
    }
  ],
  "total": 12
}
```

### JavaScript (fetch)

```js
const res = await fetch('http://localhost:8080/api/domains');
const { domains, total } = await res.json();
console.log(`Available domains (${total}):`);
domains.forEach(d => console.log(`  - ${d.id}: ${d.description}`));
```

### Python (requests)

```python
import requests

resp = requests.get("http://localhost:8080/api/domains", timeout=5)
resp.raise_for_status()
data = resp.json()

for domain in data["domains"]:
    print(f"{domain['id']:20s} {domain['problem_count']} problems")
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type Domain struct {
    ID           string `json:"id"`
    Name         string `json:"name"`
    Description  string `json:"description"`
    ProblemCount int    `json:"problem_count"`
}

type DomainsResponse struct {
    Domains []Domain `json:"domains"`
    Total   int      `json:"total"`
}

func main() {
    resp, err := http.Get("http://localhost:8080/api/domains")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var dr DomainsResponse
    json.Unmarshal(body, &dr)
    fmt.Printf("Total domains: %d\n", dr.Total)
    for _, d := range dr.Domains {
        fmt.Printf("  %s — %d problems\n", d.Name, d.ProblemCount)
    }
}
```

---

## WS /ws

Opens a WebSocket connection for real-time AI-guided sessions. The server accepts `frame`, `text`, and `ping` messages and responds with `response`, `pong`, and `error` messages.

**Connection URL**

| Environment | URL |
|-------------|-----|
| Local dev | `ws://localhost:8080/ws` |
| Production | `wss://physicscopilot-server.onrender.com/ws` |

**Connection limits per IP**

| Limit | Value |
|-------|-------|
| Max concurrent connections | 10 |
| Max message size | 10 MB |
| Max camera frame rate | 5 fps |
| Heartbeat ping interval | 30 s |
| Pong timeout | 40 s |

### JavaScript — Browser (WebSocket API)

```js
const ws = new WebSocket('ws://localhost:8080/ws');

ws.addEventListener('open', () => {
  console.log('Connected');

  // Send a text message
  ws.send(JSON.stringify({
    type: 'text',
    content: 'The extruder keeps clicking — what should I check?',
    timestamp: Date.now(),
  }));
});

ws.addEventListener('message', (event) => {
  const msg = JSON.parse(event.data);

  switch (msg.type) {
    case 'response':
      console.log('AI guidance:', msg.text);
      if (msg.overlay?.regions?.length) {
        console.log('Overlay regions:', msg.overlay.regions);
      }
      if (msg.step) {
        console.log(`Step ${msg.step.index + 1}/${msg.step.total}: ${msg.step.description}`);
      }
      break;

    case 'pong':
      console.log('Pong received');
      break;

    case 'error':
      console.error('Server error:', msg.error);
      break;

    default:
      console.warn('Unknown message type:', msg.type);
  }
});

ws.addEventListener('close', (event) => {
  console.log(`Disconnected — code ${event.code}: ${event.reason}`);
});

ws.addEventListener('error', (err) => {
  console.error('WebSocket error:', err);
});
```

### JavaScript — Sending a camera frame

```js
// Capture a frame from a <video> element and send it as a base64 JPEG
function sendFrame(ws, videoElement) {
  const canvas = document.createElement('canvas');
  canvas.width = videoElement.videoWidth;
  canvas.height = videoElement.videoHeight;

  const ctx = canvas.getContext('2d');
  ctx.drawImage(videoElement, 0, 0);

  // toDataURL returns "data:image/jpeg;base64,<data>" — strip the prefix
  const base64 = canvas.toDataURL('image/jpeg', 0.8).split(',')[1];

  ws.send(JSON.stringify({
    type: 'frame',
    data: base64,
    timestamp: Date.now(),
  }));
}

// Stream at 5 fps (server will drop excess frames above 5/s)
const video = document.querySelector('video');
setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    sendFrame(ws, video);
  }
}, 200);
```

### JavaScript — Application-level ping/keepalive

```js
// Send a ping every 25 s to keep the connection alive
// (Server sends WebSocket Ping at 30 s; this is an application-level ping)
setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({
      type: 'ping',
      timestamp: Date.now(),
    }));
  }
}, 25_000);
```

### JavaScript — Node.js (ws library)

```js
import WebSocket from 'ws';

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', () => {
  ws.send(JSON.stringify({
    type: 'text',
    content: 'Laptop fan is making a grinding noise',
    timestamp: Date.now(),
  }));
});

ws.on('message', (data) => {
  const msg = JSON.parse(data.toString());
  if (msg.type === 'response') {
    console.log('Response:', msg.text);
    ws.close(); // close after first response in this example
  }
});

ws.on('close', (code, reason) => {
  console.log(`Closed: ${code} ${reason}`);
});
```

### Python (websocket-client)

```python
import json
import time
import websocket  # pip install websocket-client

def on_open(ws):
    print("Connected")
    ws.send(json.dumps({
        "type": "text",
        "content": "The dishwasher is leaking from the door seal",
        "timestamp": int(time.time() * 1000),
    }))

def on_message(ws, message):
    msg = json.loads(message)
    if msg["type"] == "response":
        print("AI:", msg["text"])
        if msg.get("step"):
            step = msg["step"]
            print(f"  Step {step['index'] + 1}/{step['total']}: {step['description']}")
        ws.close()
    elif msg["type"] == "error":
        print("Error:", msg["error"])
        ws.close()

def on_error(ws, error):
    print("WS error:", error)

def on_close(ws, code, reason):
    print(f"Closed: {code} — {reason}")

ws = websocket.WebSocketApp(
    "ws://localhost:8080/ws",
    on_open=on_open,
    on_message=on_message,
    on_error=on_error,
    on_close=on_close,
)
ws.run_forever()
```

### Python — Sending a JPEG frame

```python
import base64
import json
import time

# Read a JPEG file from disk and send as a frame
with open("test_frame.jpg", "rb") as f:
    jpeg_bytes = f.read()

b64_data = base64.b64encode(jpeg_bytes).decode("utf-8")

frame_msg = json.dumps({
    "type": "frame",
    "data": b64_data,
    "timestamp": int(time.time() * 1000),
})

ws.send(frame_msg)  # ws is an open websocket.WebSocketApp instance
```

---

## Error Reference

| Scenario | HTTP / WS code | Response |
|----------|---------------|----------|
| Rate limit exceeded | HTTP 429 | `{"error":"rate limit exceeded"}` |
| Session not found | HTTP 404 | `{"error":"session not found"}` |
| Invalid request body | HTTP 400 | `{"error":"<validation message>"}` |
| Server error | HTTP 500 | `{"error":"internal server error"}` |
| Connection limit reached | WS 1008 | Close frame: `"connection limit reached"` |
| Frame too large | WS 1009 | Close frame: `"message too big"` |
| AI timeout | WS `error` msg | `{"type":"error","error":"Request timed out, please retry"}` |
| Empty text content | WS `error` msg | `{"type":"error","error":"Message content cannot be empty"}` |
