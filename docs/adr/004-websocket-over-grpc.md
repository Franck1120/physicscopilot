# ADR-004: WebSocket Instead of gRPC or WebRTC for Frame Streaming

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

The live camera feed must be sent from the Flutter client to the Go server
continuously. The transport must:

1. Support binary payloads (base64 JPEG frames, ~20–100 KB each).
2. Carry bidirectional messages (frames up, AI responses down).
3. Work from a Flutter mobile app behind typical carrier NAT and corporate
   firewalls.
4. Be simple enough for a single-engineer implementation.

Protocols considered:

| Protocol | Bidirectional | Firewall-friendly | Flutter support | Complexity |
|----------|---------------|-------------------|-----------------|------------|
| **WebSocket** | Yes | Excellent (port 443) | `web_socket_channel` | Low |
| gRPC-Web | Unidirectional streams | Moderate (needs proxy) | `grpc` package | High |
| WebRTC | Yes | Good (STUN/TURN) | `flutter_webrtc` | Very high |
| Long-polling | No (pull) | Excellent | `dio` | Low (but inefficient) |

gRPC streaming would require a gRPC-Web proxy (e.g. Envoy) in front of the
Go server for browser compatibility, and its HTTP/2 multiplexing adds
complexity without a tangible benefit for this use case. WebRTC's peer-to-
peer capability is unnecessary here — the mobile client always talks to a
fixed server — and its signalling, ICE, and TURN infrastructure would add
significant operational overhead.

The Go server uses `github.com/gofiber/websocket/v2` (backed by
`github.com/fasthttp/websocket`) and the Flutter client uses
`web_socket_channel: ^3.0.1`.

---

## Decision

Use **WebSocket** (RFC 6455) over TLS (wss://) for all real-time
communication between the Flutter app and the Go server. Each session
maintains a single persistent connection. Message framing uses JSON text
frames for control messages and base64-encoded JPEG payloads embedded in
JSON for frame data.

---

## Consequences

### Positive

- **Universal firewall compatibility.** WebSocket upgrades from HTTP/HTTPS
  on port 443; virtually no corporate firewall or mobile carrier blocks it.
- **Simple protocol.** JSON text frames are human-readable during
  development and debuggable with browser DevTools or `wscat`.
- **Mature Flutter library.** `web_socket_channel` is the officially
  recommended package and integrates cleanly with Riverpod's
  `StreamProvider`.
- **Low server complexity.** No proxy, no signalling server, no ICE agent.
  The handler is a single goroutine per connection.
- **JWT authentication.** The server validates a Bearer token in the
  `Authorization` query parameter on upgrade — a standard WebSocket
  authentication pattern.

### Negative

- **Higher latency than WebRTC.** WebSocket relays every frame through the
  server; WebRTC's peer-to-peer data channels could theoretically achieve
  lower round-trip latency, but the bottleneck here is Gemini's inference
  time (~1–2 s), not transport.
- **No built-in multiplexing.** Each session uses one TCP connection. At
  very high concurrency, head-of-line blocking within a connection could be
  an issue; at expected scale (hundreds of sessions) it is not.
- **Base64 overhead.** Encoding frames as base64 inside JSON adds ~33%
  payload overhead compared to binary WebSocket frames. Switching to binary
  frames is a low-effort optimisation deferred until profiling shows it
  matters.
- **No built-in flow control.** The client sends frames at a fixed interval;
  if the server falls behind, frames queue in the kernel buffer. A future
  backpressure signal (e.g. a `pause`/`resume` control message) would
  address this.
