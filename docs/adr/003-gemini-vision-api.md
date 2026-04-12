# ADR-003: Gemini 2.5 Flash as the Primary AI Backend

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

The core value proposition of PhysicsCopilot is accurate analysis of camera
frames showing real-world equipment. The AI backend must:

1. Accept a base64-encoded JPEG frame + a text conversation context.
2. Return structured JSON (`analysis`, `problem`, `instruction`, `overlay`)
   in under ~3 seconds on a user's mobile connection.
3. Be callable without an expensive self-hosted inference stack during the
   MVP phase.

Vision-capable models evaluated at design time:

| Model | Vision quality | Cost (input/1M tok) | Free tier |
|-------|---------------|---------------------|-----------|
| **Gemini 2.5 Flash** | Excellent | ~$0.075 | 1 000 req/day |
| GPT-4o | Excellent | ~$2.50 | None |
| Claude 3.5 Sonnet | Excellent | ~$3.00 | None |
| LLaVA (self-hosted) | Good | Hardware cost | Unlimited |

Gemini 2.5 Flash's free tier (1 000 requests/day at the time of decision) is
sufficient for development and early production with a handful of users.
Its latency profile (~1–2 s for a 720p frame) meets the interactive streaming
requirement. GPT-4o and Claude are higher quality on some tasks but offer no
free tier and cost ~30–40× more per request.

The server calls Gemini over its public REST API
(`https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent`),
with a built-in retry loop (up to 3 attempts) and per-minute rate-limit
protection via a `golang.org/x/time/rate.Limiter`.

---

## Decision

Use **Gemini 2.5 Flash** as the default AI backend, accessed via direct
HTTP calls (no official Go SDK). The `AIBackend` interface
(`server/internal/services/ai_backend.go`) abstracts the backend so any
future model can be swapped in by changing the `AI_BACKEND` environment
variable at deploy time.

---

## Consequences

### Positive

- **Free tier covers the MVP.** No billing account required to ship and
  demonstrate the product.
- **Response latency.** Flash is optimised for throughput over reasoning
  depth, giving ~1–2 s round-trips for vision tasks.
- **JSON-mode reliability.** Gemini 2.5 Flash reliably respects the
  structured JSON output instruction when temperature is kept low (0.2).
- **Pluggable via `AIBackend` interface.** Swapping to OpenAI, Claude, or a
  self-hosted model requires implementing one method (`AnalyzeFrame`) and
  updating the `AI_BACKEND` env var. No caller changes are needed.
- **No vendor lock-in in application code.** The factory function
  `NewAIBackend()` is the only place that names "gemini" as the default.

### Negative

- **Rate limit on free tier.** 1 000 requests/day means a single busy user
  can exhaust the free quota. Production deployments need a paid key.
- **Google infrastructure dependency.** Outages or API changes at Google
  affect all users simultaneously. No fallback inference path exists today.
- **No official Go SDK.** The REST API is called manually with
  `encoding/json` marshalling; changes to the API schema require manual
  updates to the request/response structs.
- **Vision quality varies by scene.** Gemini Flash trades some accuracy for
  speed; complex or dark scenes may require a heavier model (Gemini Pro,
  GPT-4o) which will be swappable via the `AIBackend` interface.
