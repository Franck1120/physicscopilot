# Security Contacts

## Reporting a Vulnerability

We take security seriously. If you believe you have found a security
vulnerability in PhysicsCopilot, please **do not open a public GitHub issue**.
Instead, follow the responsible-disclosure process described below.

### How to Report

1. **Email** — Send a detailed report to the security contact listed in
   `SECURITY.md` (see also the GitHub repository's **Security** tab for
   private vulnerability reporting).

2. **GitHub private advisory** — Use
   [GitHub Security Advisories](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability)
   on this repository to report privately and collaborate on a fix before
   public disclosure.

### What to Include

A high-quality report significantly reduces response time. Please include:

- **Affected component** — endpoint, package, or service
- **Vulnerability type** — OWASP category or CWE identifier if known
- **Steps to reproduce** — minimal reproduction with exact request payloads
  or code snippets
- **Impact assessment** — what an attacker could achieve
- **Suggested fix** — optional, but appreciated
- **Your contact information** — so we can coordinate the fix and credit you

### What to Expect

| Timeline          | Action                                                       |
|-------------------|--------------------------------------------------------------|
| Within 48 hours   | Acknowledgement of receipt                                   |
| Within 7 days     | Initial triage and severity assessment                       |
| Within 30 days    | Patch or mitigation for confirmed High/Critical issues       |
| Within 90 days    | Public disclosure after patch is released                    |

These are best-effort targets. Complex vulnerabilities may take longer.
We will keep you informed of our progress.

### Severity Classification

We use the [CVSS v3.1](https://www.first.org/cvss/v3-1/) scoring system
for severity classification:

| Severity | CVSS Score |
|----------|-----------|
| Critical | 9.0–10.0  |
| High     | 7.0–8.9   |
| Medium   | 4.0–6.9   |
| Low      | 0.1–3.9   |

### Scope

In scope for responsible disclosure:

- The Go server (`server/`)
- Authentication and authorisation logic
- Database query construction
- API input validation
- WebSocket message handling
- Docker and infrastructure configuration

Out of scope:

- Vulnerabilities in third-party dependencies that have no upstream fix
- Issues that require physical access to the host
- Social engineering attacks
- Denial-of-service via resource exhaustion that requires thousands of
  authenticated requests

### Recognition

We maintain a **Hall of Fame** in the `SECURITY.md` file for researchers
who report confirmed vulnerabilities and agree to be credited.

### Contact

See `SECURITY.md` for the primary security contact address.

---

*This policy is inspired by the
[CERT Coordination Center responsible disclosure guidelines](https://www.kb.cert.org/vuls/guidance/).*
