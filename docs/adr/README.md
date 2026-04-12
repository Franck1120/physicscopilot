# Architecture Decision Records

This directory contains the Architecture Decision Records (ADRs) for
PhysicsCopilot. Each ADR documents a significant architectural choice — the
context that drove it, the decision taken, and the trade-offs accepted.

ADRs are append-only: accepted decisions are not deleted; they are superseded
by new ADRs when the architecture evolves.

---

## Index

| # | Title | Status |
|---|-------|--------|
| [001](001-go-fiber-backend.md) | Go + Fiber as the Backend Runtime | Accepted |
| [002](002-flutter-mobile.md) | Flutter for the Mobile Client | Accepted |
| [003](003-gemini-vision-api.md) | Gemini 2.5 Flash as the Primary AI Backend | Accepted |
| [004](004-websocket-over-grpc.md) | WebSocket Instead of gRPC or WebRTC for Frame Streaming | Accepted |
| [005](005-tfidf-over-embeddings.md) | TF-IDF Keyword Search Instead of Vector Embeddings for RAG | Accepted |
| [006](006-supabase-postgres.md) | Supabase/Postgres with In-Memory Fallback | Accepted |

---

## Template

New ADRs should follow this structure:

```markdown
# ADR-NNN: Title

**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-NNN
**Date:** YYYY-MM-DD

## Context
What situation or requirement prompted this decision?

## Decision
What was decided?

## Consequences
### Positive
### Negative
```
