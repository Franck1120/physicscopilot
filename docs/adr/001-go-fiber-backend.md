# ADR-001: Go + Fiber as the Backend Runtime

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

PhysicsCopilot streams live camera frames over WebSocket at up to ~1 fps per
connected client, forwards them to Gemini, and must sustain many concurrent
sessions without blocking. The server also needs to ship as a single static
binary inside a Docker image small enough for a free Render instance.

Candidate runtimes evaluated:

| Runtime | Concurrency model | Binary size | Cold start |
|---------|-------------------|-------------|------------|
| **Go**  | goroutines (M:N)  | ~10 MB      | < 50 ms    |
| Node.js | event loop + libuv | 50-100 MB (with Node) | ~200 ms |
| Python (FastAPI) | asyncio | 80+ MB (with runtime) | ~400 ms |

Go's goroutines make it straightforward to give each WebSocket connection its
own goroutine without the overhead of OS threads. Fiber v2 wraps fasthttp,
which avoids allocating an `http.Request` per request and achieves ~300k
req/s on a single core in benchmarks — well above what this project will ever
need, but the headroom removes performance as a bottleneck entirely.

---

## Decision

Use **Go 1.25** with **Fiber v2** (backed by fasthttp) as the only server
runtime. Deployment targets a single Docker image built with a multi-stage
`FROM scratch` final stage.

---

## Consequences

### Positive

- **Concurrency for free.** Each WebSocket session runs in its own goroutine;
  the runtime schedules thousands of them onto a handful of OS threads.
- **Minimal image size.** The final Docker layer is the compiled binary only
  (~10 MB), keeping cold starts on Render's free tier under 1 second.
- **Single binary deployment.** No interpreter, no virtual environment, no
  dependency resolver at runtime. `docker run` is the entire deployment.
- **Strong standard library.** `log/slog`, `context`, `sync`, `net/http`
  cover most infrastructure needs without external dependencies.
- **Type safety at compile time.** Eliminates a class of runtime bugs common
  in dynamic-language servers.

### Negative

- **Smaller ecosystem than Node/Python.** Fewer ready-made AI/ML SDK
  integrations. Gemini's REST API is called directly over HTTP rather than
  through an official Go SDK.
- **Developer pool.** Go engineers are less common than Node or Python
  engineers, which may slow onboarding of future contributors.
- **Dart learning curve is separate.** Hiring someone who knows both Go and
  Flutter/Dart is harder than finding a full-stack JS developer.
- **fasthttp incompatibility.** Fiber's fasthttp core is not compatible with
  the standard `net/http` interface; middleware that targets `net/http`
  requires the `gofiber/adaptor` shim (used for the Prometheus handler).
