# Security Contacts and Responsible Disclosure

PhysicsCopilot takes security seriously. This document describes how to report vulnerabilities, what we consider in scope, and our response commitments.

---

## How to Report a Vulnerability

**Use GitHub's private security advisory feature.** This keeps the report confidential between you and the maintainers until a fix is released.

1. Go to: https://github.com/Franck1120/physicscopilot/security/advisories/new
2. Fill in the advisory form with:
   - A clear description of the vulnerability
   - Steps to reproduce (proof of concept if available)
   - The potential impact
   - Affected versions or components
3. Submit the advisory. You will receive an acknowledgment within **48 hours**.

**Do not open a public GitHub issue for security vulnerabilities.** Public disclosure before a fix is available puts users at risk.

---

## Response Timeline

| Stage | Target time |
|-------|------------|
| Acknowledge receipt of report | Within 48 hours |
| Confirm or reject the report | Within 7 days |
| Share a remediation plan | Within 14 days |
| Release a fix | Within 90 days |
| Public disclosure (coordinated) | After fix is released |

We will keep you informed of progress throughout the process. If you have not received acknowledgment within 48 hours, comment on the advisory to follow up.

We ask that you give us a reasonable window to fix the issue before disclosing it publicly or to third parties.

---

## Scope

### In scope

- The Go server (`server/`) — any endpoint, WebSocket handler, authentication logic, or middleware
- The Flutter mobile app (`app/`) — local data storage, network requests, certificate validation
- Supabase Row Level Security policies (`supabase/migrations/`)
- Docker images and build artifacts produced from this repository
- The knowledge base API (`/api/domains`, RAG retrieval logic)
- Dependency vulnerabilities in direct dependencies (Go modules, Flutter packages)

### Out of scope

- Vulnerabilities in third-party services we depend on (Supabase, Gemini API, Render.com) — report these to the respective vendor
- Denial of service attacks that require sending millions of requests (rate limiting is already enforced)
- Self-XSS or attacks that require physical access to the victim's device
- Issues in development tooling only (not affecting production users)
- Missing security headers on pages served by Render's default reverse proxy (report to Render)
- Theoretical vulnerabilities with no demonstrable impact

---

## Severity Levels

We assess severity using the CVSS v3.1 framework. Approximate mapping:

| Severity | CVSS Score | Examples |
|----------|------------|---------|
| Critical | 9.0–10.0 | Remote code execution, authentication bypass allowing full data access |
| High | 7.0–8.9 | Privilege escalation, mass data exfiltration, WAF bypass |
| Medium | 4.0–6.9 | Partial data exposure, CSRF on state-changing operations, reflected XSS |
| Low | 0.1–3.9 | Information disclosure (server version, error messages), minor misconfigurations |

---

## Bug Bounty Policy

PhysicsCopilot does **not currently offer a paid bug bounty program**. We are a small open-source project.

We do offer public acknowledgment in the release notes and in a `SECURITY_ACKNOWLEDGMENTS.md` file for responsibly disclosed vulnerabilities that are confirmed and fixed.

---

## Known Security Controls

The following security controls are already implemented. Please verify a bypass before reporting:

| Control | Implementation |
|---------|---------------|
| JWT authentication | Supabase JWT on the WebSocket endpoint (`SUPABASE_JWT_SECRET`) |
| Rate limiting | 60 req/min per IP on REST endpoints, 5 fps per WS connection, 10 WS connections per IP |
| Input validation | Frame size capped at 10 MB, text content capped at 50 KB, base64 validation on `frame` messages |
| Row Level Security | Enabled on all Supabase tables; service role key only used server-side |
| CORS | Restricted to `ALLOWED_ORIGINS` environment variable; wildcard only in development mode |
| Secret management | All secrets in environment variables; none hardcoded in source |
| Dependency scanning | Dependabot alerts enabled on the GitHub repository |
| Error messages | Internal details (stack traces, API keys, SQL errors) are never sent to clients |

---

## Contact

Primary contact method: **GitHub private security advisory**
https://github.com/Franck1120/physicscopilot/security/advisories/new

GitHub security policy page:
https://github.com/Franck1120/physicscopilot/security/policy
