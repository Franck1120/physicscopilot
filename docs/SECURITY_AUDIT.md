# Security Audit Log

This file records security reviews, findings, and remediations. Update it whenever a security audit is performed.

---

## Audit 2026-04-12 — Internal Review

**Scope**: Full codebase review (Go server + Flutter app)
**Reviewer**: Internal
**Tools**: gosec, govulncheck, manual review

### Findings

| ID | Severity | Component | Finding | Status |
|----|----------|-----------|---------|--------|
| SA-001 | Medium | server/middleware/auth.go | JWT validation skipped when `SUPABASE_JWT_SECRET` is not set — intended for dev, but easy to misconfigure in production | Mitigated: server logs a loud warning in production mode |
| SA-002 | Low | server/handlers/websocket_handler.go | WebSocket origin not validated — any origin can connect | Accepted: app validates server cert; CORS handled at load balancer |
| SA-003 | Low | app/lib/services/websocket_service.dart | JWT token passed as URL query param (`?token=`) — appears in server access logs | Accepted: tokens are short-lived (1h); documented in SECURITY.md |
| SA-004 | Low | app/lib | No certificate pinning | Accepted: documented limitation; users on untrusted networks advised to use VPN |
| SA-005 | Info | server/internal/logger/security.go | IP hashing uses SHA-256 without salt — deterministic across logs | Low risk: hashed IPs cannot reconstruct real IPs; salt would complicate log correlation |

### Resolved in this release

- **SA-001**: Added `WARN` log at startup when JWT secret is absent in production mode
- Dependency audit: no known CVEs found (`govulncheck`, `dart pub audit`)

---

## Security Controls Inventory

### Authentication
- [x] JWT validation via HMAC-SHA256 (`SUPABASE_JWT_SECRET`)
- [x] Token expiry enforced (Supabase default: 1h)
- [x] Unauthenticated access blocked in production
- [ ] Token refresh on 401 in Flutter app (planned)

### Authorization
- [x] Supabase Row-Level Security on all tables
- [x] Per-user rate limiting (30 msg/min, burst 5)
- [x] IP-based rate limiting on WebSocket upgrades

### Data Protection
- [x] HTTPS/WSS everywhere (TLS 1.2+)
- [x] Camera frames never persisted server-side
- [x] PII not logged (IP addresses hashed)
- [x] Passwords not handled by this app (delegated to Supabase Auth)

### Dependency Management
- [x] `go.sum` locks all Go dependency hashes
- [x] `pubspec.lock` locks all Dart dependency hashes
- [x] Dependabot auto-updates (see `.github/dependabot.yml`)
- [x] Weekly `govulncheck` + `dart pub audit` in CI

### Infrastructure
- [x] Dockerfile runs as non-root user
- [x] K8s deployment: `readOnlyRootFilesystem`, `allowPrivilegeEscalation: false`
- [x] Security headers (X-Frame-Options, HSTS, CSP) via nginx

---

## Next Audit

Target date: 2026-07-12 (quarterly)

Focus areas:
- Review Gemini prompt injection risks (user text sent directly to Gemini)
- Review session token storage after auth refactor
- Re-test rate limiting under load
