# API Examples

Practical `curl` examples for every PhysicsCopilot endpoint.

All examples assume the server is running at `http://localhost:8080`.  
Replace `<JWT>` with a valid Supabase JWT when authentication is required.

---

## GET /health

Returns service health, uptime, and optional database status.

**Request**

```bash
curl -s http://localhost:8080/health
```

**Response**

```json
{
  "status": "ok",
  "service": "physicscopilot",
  "version": "0.1.0",
  "uptime": "2h34m15s",
  "active_connections": 3,
  "memory_mb": 42,
  "db_status": "ok",
  "db_pool": {
    "total_conns": 4,
    "idle_conns": 3,
    "max_conns": 10
  }
}
```

---

## GET /version

Returns build metadata. Publicly cacheable for one hour.

**Request**

```bash
curl -s http://localhost:8080/version
```

**Response**

```json
{
  "version": "0.1.0",
  "build_time": "2026-04-12T00:00:00Z",
  "go_version": "go1.24.1",
  "api_version": "v1"
}
```

---

## GET /api/version

Same as `/version` but within the `/api` namespace, consistent with other API endpoints.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" http://localhost:8080/api/version
```

**Response**

```json
{
  "version": "0.1.0",
  "build_time": "2026-04-12T00:00:00Z",
  "go_version": "go1.24.1",
  "api_version": "v1"
}
```

---

## GET /api/docs

Returns the OpenAPI 3.0 specification in YAML format.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" http://localhost:8080/api/docs
```

**Response** — raw YAML (Content-Type: `application/yaml`).

---

## GET /api/swagger

Serves an HTML page embedding Swagger UI, pointed at `/api/docs`.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" http://localhost:8080/api/swagger
```

Open in a browser for the interactive experience:

```
http://localhost:8080/api/swagger
```

---

## POST /api/sessions

Creates a new repair session for a given device.

**Request**

```bash
curl -s -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{
    "device_brand": "Prusa",
    "device_model": "MK4"
  }'
```

**Response**

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "active",
  "device": {
    "brand": "Prusa",
    "model": "MK4"
  },
  "problem_detected": "",
  "current_step": 0,
  "total_steps": 0,
  "created_at": "2026-04-12T10:00:00Z",
  "last_activity": "2026-04-12T10:00:00Z"
}
```

---

## GET /api/sessions

Lists sessions with pagination, sorting, and filtering.

### Basic list

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  "http://localhost:8080/api/sessions"
```

### Pagination

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  "http://localhost:8080/api/sessions?page=2&page_size=10"
```

Query parameters:

| Parameter   | Default | Description                          |
|-------------|---------|--------------------------------------|
| `page`      | `1`     | Page number (1-indexed)              |
| `page_size` | `20`    | Items per page (max 100)             |

### Sorting

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  "http://localhost:8080/api/sessions?sort=created_at&order=desc"
```

### Filtering by status

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  "http://localhost:8080/api/sessions?status=active"
```

**Response**

```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "status": "active",
      "device": { "brand": "Prusa", "model": "MK4" },
      "current_step": 2,
      "total_steps": 5,
      "created_at": "2026-04-12T10:00:00Z",
      "last_activity": "2026-04-12T10:05:30Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

---

## GET /api/sessions/:id

Fetches a single session by ID. Returns `304 Not Modified` when the
`If-None-Match` header matches the current ETag.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/api/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Conditional request (ETag)**

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  -H 'If-None-Match: W/"aabbccdd"' \
  http://localhost:8080/api/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Response** — same shape as a single item in the list response above.

---

## GET /api/sessions/:id/steps

Returns the step-by-step instructions generated for a session.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/api/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890/steps
```

**Response**

```json
{
  "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "steps": [
    {
      "step_number": 1,
      "instruction": "Check the extruder temperature sensor wiring.",
      "verified": true
    },
    {
      "step_number": 2,
      "instruction": "Re-seat the thermistor connector on the hotend.",
      "verified": false
    }
  ]
}
```

---

## DELETE /api/sessions/:id

Soft-deletes a session (marks it as `abandoned`).

**Request**

```bash
curl -s -X DELETE -H "Authorization: Bearer <JWT>" \
  http://localhost:8080/api/sessions/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Response** — `204 No Content` on success.

---

## POST /api/feedback

Submits per-step user feedback (thumbs up/down, optional comment).

**Request**

```bash
curl -s -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{
    "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "step_number": 2,
    "rating": "positive",
    "comment": "The thermistor tip was indeed loose."
  }'
```

Fields:

| Field        | Type   | Required | Values                     |
|--------------|--------|----------|----------------------------|
| `session_id` | string | yes      | UUID of the session        |
| `step_number`| int    | yes      | ≥ 0                        |
| `rating`     | string | yes      | `"positive"` / `"negative"`|
| `comment`    | string | no       | Free text                  |

**Response** — `201 Created` with the stored feedback record.

---

## GET /api/stats

Returns aggregate runtime statistics.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" http://localhost:8080/api/stats
```

**Response**

```json
{
  "uptime": "3h10m00s",
  "version": "0.1.0",
  "total_sessions": 48,
  "active_sessions": 3,
  "active_ws_connections": 2,
  "kb_documents": 120,
  "kb_loaded": true
}
```

---

## GET /api/domains

Returns the list of physics/repair domains supported by the knowledge base.

**Request**

```bash
curl -s -H "Authorization: Bearer <JWT>" http://localhost:8080/api/domains
```

**Response**

```json
{
  "domains": ["FDM printers", "electronics", "optics", "thermodynamics"]
}
```

---

## GET /metrics

Prometheus metrics endpoint — protected by HTTP Basic Auth.

Set `METRICS_USER` and `METRICS_PASSWORD` environment variables on the server.

**Request**

```bash
curl -s -u admin:secret http://localhost:8080/metrics
```

**Response** — Prometheus text exposition format.

---

## WebSocket — GET /ws

Real-time conversation interface. Pass the JWT as a query parameter.

**Connect with `websocat`**

```bash
websocat "ws://localhost:8080/ws?token=<JWT>"
```

**Send a camera frame**

```json
{
  "type": "frame",
  "device_brand": "Prusa",
  "device_model": "MK4",
  "image_b64": "<base64-encoded JPEG>"
}
```

**Receive a step response**

```json
{
  "type": "step",
  "step_number": 1,
  "instruction": "Check the PTFE tube at the hotend entry.",
  "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Ping/Pong** — the server sends a `{"type":"ping"}` heartbeat every 30 seconds.
Reply with `{"type":"pong"}` to keep the connection alive.
