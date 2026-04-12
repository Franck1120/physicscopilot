# ADR-006: Supabase/Postgres with In-Memory Fallback

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

PhysicsCopilot sessions (guidance steps, user feedback) benefit from
persistence across app restarts. However, for the MVP the server must also
run entirely without a database — in a local dev environment with a single
`docker run`, or on Render's free tier where a managed Postgres instance may
not be provisioned.

Requirements:

1. Sessions and steps can be stored and retrieved across reconnects.
2. The server boots and serves requests with **zero database configuration**.
3. When a database is available it is used transparently.
4. No ORM: raw SQL keeps query behaviour predictable and avoids reflection
   overhead in a performance-sensitive server.

Database options considered:

| Option | Scalability | Operational cost | Free tier | SDK quality |
|--------|-------------|------------------|-----------|-------------|
| **Supabase (Postgres + Auth)** | High | Low (managed) | 500 MB | Excellent |
| PlanetScale (MySQL) | High | Low | Limited | Good |
| SQLite (embedded) | Low-medium | Zero | Unlimited | Good |
| MongoDB Atlas | High | Low | 512 MB | Good |

Supabase provides managed Postgres with built-in Auth (JWT verification),
Storage, and Row-Level Security — all services the project will need as it
matures. The free tier (500 MB, 2 projects) is sufficient for the MVP. Raw
`pgx/v5` is used instead of an ORM or the Supabase JS client, as Go has no
official Supabase SDK.

The `DBBackend` interface (`server/internal/services/db_service.go`)
abstracts the persistence layer the same way `AIBackend` abstracts the AI
model. `SessionService` calls `DBBackend` methods and, on any error, logs a
warning and continues serving from its in-memory store.

---

## Decision

Use **Supabase (managed Postgres)** as the optional persistence backend,
connected via `pgx/v5` with a connection pool (`pgxpool`). When
`DATABASE_URL` is not set the server runs with a pure **in-memory store**
(`sync.Map` within `SessionService`). The `DBBackend` interface is the
boundary between the two modes.

---

## Consequences

### Positive

- **Zero-config local development.** Omitting `DATABASE_URL` from the
  environment gives a fully functional server with no Docker Compose
  dependency (beyond the server itself).
- **Managed infrastructure.** Supabase handles backups, failover, and
  connection pooling. No DBA work needed for the MVP.
- **Auth integration path.** Supabase Auth will issue JWTs that the
  WebSocket middleware already validates, providing a clear upgrade path for
  user accounts.
- **Non-fatal DB errors.** `SessionService` treats all `DBBackend` errors
  as warnings, so a database outage degrades gracefully to in-memory-only
  mode rather than taking the server down.
- **pgx/v5 performance.** `pgxpool` reuses connections and supports named
  prepared statements, giving near-wire-speed query performance.

### Negative

- **In-memory data is ephemeral.** Without `DATABASE_URL`, all session
  history is lost on server restart. This is acceptable for demos but not
  for production.
- **Supabase free tier limits.** 500 MB storage and 2 GB bandwidth/month.
  High-volume video session metadata could approach these limits. Upgrading
  to a paid plan or self-hosting Postgres is the mitigation.
- **No official Go SDK.** Supabase's REST and Realtime APIs are not used;
  only raw Postgres via `pgx` is supported today. Features like real-time
  subscriptions or Storage require additional integration work.
- **Single-region by default.** Supabase free projects are single-region.
  Latency for users far from the selected region affects session persistence
  writes (not reads, which come from the in-memory cache).
