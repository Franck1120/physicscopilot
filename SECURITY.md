# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in PhysicsCopilot, please report it **privately** by emailing:

**security@physicscopilot.app**

Include:
- Description of the vulnerability
- Steps to reproduce
- Affected component (app / server / infra)
- Potential impact

We will acknowledge your report within 48 hours and aim to release a fix within 14 days for critical issues.

**Please do not open public GitHub issues for security vulnerabilities.**

---

## Authentication & Authorization

### JWT Flow
- The server validates JWTs issued by Supabase using `SUPABASE_JWT_SECRET`
- Token is passed as `?token=<jwt>` on the WebSocket URL and as `Authorization: Bearer <token>` on REST endpoints
- Tokens are short-lived (Supabase default: 1 hour) and automatically refreshed by the Flutter app via Supabase client
- Token revocation: sign the user out via Supabase Auth — the next request with the old token will fail validation
- In dev mode (no `SUPABASE_JWT_SECRET` configured), the server accepts unauthenticated connections — **never deploy to production without this env var**

### Session Isolation
- Sessions are scoped per user ID extracted from the JWT
- Row-Level Security (RLS) in Supabase ensures users can only read/write their own sessions

---

## Rate Limiting

The Go server applies rate limiting at the middleware layer:

| Endpoint | Limit |
|----------|-------|
| `GET /ws` (WebSocket upgrade) | 10 connections / minute per IP |
| `POST /api/sessions` | 30 requests / minute per user |
| `GET /api/sessions` | 60 requests / minute per user |

Clients that exceed limits receive `429 Too Many Requests`.

---

## Security Headers

All HTTP responses include:

| Header | Value |
|--------|-------|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Content-Security-Policy` | `default-src 'self'` |

---

## Data Privacy

- Camera frames are processed in memory and **never persisted** on the server
- AI responses are stored only if the user explicitly saves the session
- No PII is logged — user IDs are hashed in server logs
- Frames are transmitted over WSS (TLS 1.2+) — never over unencrypted WS

---

## Dependency Security

- Flutter dependencies are pinned in `pubspec.lock`
- Go dependencies are pinned in `go.sum`
- Security updates: run `flutter pub upgrade --major-versions` and `go get -u` quarterly and before each release

---

## Known Limitations

- The `?token=` parameter in WebSocket URLs may appear in server access logs — rotate tokens regularly in production
- The app does not implement certificate pinning — users on untrusted networks should use a VPN
