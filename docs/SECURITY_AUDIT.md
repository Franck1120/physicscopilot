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

## Security Checklist

### Authentication

| # | Item | Status | Notes |
|---|------|--------|-------|
| A-1 | JWT validation with HMAC-SHA256 | ✅ | `middleware/auth.go` validates `SUPABASE_JWT_SECRET` |
| A-2 | Token expiry enforced | ✅ | Supabase default: 1 hour |
| A-3 | Unauthenticated access blocked in production | ✅ | `WSAuthMiddleware` rejects missing/invalid tokens |
| A-4 | Token refresh on 401 in Flutter app | ⚠️ | Planned — currently the user must re-login |
| A-5 | Password handling delegated to Supabase Auth | ✅ | App never sees raw passwords |

### API Security

| # | Item | Status | Notes |
|---|------|--------|-------|
| B-1 | Per-user rate limiting (30 msg/min, burst 5) | ✅ | `middleware/userlimit.go` |
| B-2 | IP-based rate limiting on HTTP routes | ✅ | `middleware/ratelimit.go` |
| B-3 | CORS restricted to allowed origins | ✅ | `ALLOWED_ORIGINS` env var; `*` only in dev |
| B-4 | HSTS header in production | ✅ | Enforced via middleware and `infra/nginx.conf` |
| B-5 | Request timeout middleware (30s) | ✅ | Prevents slow-loris attacks |
| B-6 | Panic recovery middleware | ✅ | Catches panics, logs stack trace, returns 500 |
| B-7 | Input validation on text messages | ✅ | HTML strip + 5000-char cap (XSS mitigation) |
| B-8 | JPEG magic-byte check on frames | ✅ | Rejects non-JPEG binary data |

### Data Protection

| # | Item | Status | Notes |
|---|------|--------|-------|
| C-1 | Camera frames never persisted server-side | ✅ | Processed in memory, discarded after response |
| C-2 | PII not logged | ✅ | IP addresses hashed with SHA-256 |
| C-3 | HTTPS/WSS everywhere | ✅ | TLS 1.2+ enforced in production |
| C-4 | Supabase RLS on all tables | ✅ | Users can only access their own rows |
| C-5 | Database connection encrypted | ✅ | `sslmode=require` in `DATABASE_URL` |

### Infrastructure

| # | Item | Status | Notes |
|---|------|--------|-------|
| D-1 | Docker runs as non-root user | ✅ | `USER nobody` in Dockerfile |
| D-2 | K8s: readOnlyRootFilesystem | ✅ | `infra/k8s/deployment.yaml` |
| D-3 | K8s: allowPrivilegeEscalation false | ✅ | Security context configured |
| D-4 | Security headers via nginx | ✅ | X-Frame-Options, CSP, X-Content-Type-Options |
| D-5 | Secrets not in source code | ✅ | All secrets via env vars or K8s secrets |

### Dependencies

| # | Item | Status | Notes |
|---|------|--------|-------|
| E-1 | `go.sum` locks dependency hashes | ✅ | Tamper detection on Go modules |
| E-2 | `pubspec.lock` locks Dart dependencies | ✅ | Reproducible Flutter builds |
| E-3 | Dependabot auto-updates enabled | ✅ | `.github/dependabot.yml` |
| E-4 | `govulncheck` in CI | ✅ | Weekly scan for Go CVEs |
| E-5 | `dart pub audit` in CI | ✅ | Weekly scan for Dart CVEs |
| E-6 | Trivy container scan | ⚠️ | Planned — not yet integrated into CI |

---

## Tools Used

| Tool | Purpose | Frequency |
|------|---------|-----------|
| `gosec` | Go static security analysis | Every PR via CI |
| `govulncheck` | Go dependency CVE scanner | Weekly in CI |
| `dart pub audit` | Dart dependency vulnerability check | Weekly in CI |
| `gitleaks` | Secret detection in git history | Pre-commit hook |
| `trivy` | Container image vulnerability scan | Planned for CI |

---

## Remediation Priorities

| Priority | Item | Target Date |
|----------|------|-------------|
| Medium | Add token refresh on 401 (A-4) | v0.19.0 |
| Medium | Integrate Trivy container scan (E-6) | v0.19.0 |
| Low | Add salt to IP hash (SA-005) | Backlog |
| Low | Certificate pinning in Flutter app (SA-004) | Backlog |
| Medium | Prompt injection review for Gemini input | Next audit |

---

## Next Audit

Target date: 2026-07-12 (quarterly)

Focus areas:
- Review Gemini prompt injection risks (user text sent directly to Gemini)
- Review session token storage after auth refactor
- Re-test rate limiting under load
- Evaluate Trivy scan results after CI integration
