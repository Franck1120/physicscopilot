# Roadmap

Status: Done · In Progress · Planned · Dropped

---

## v1.0 — Foundation (Done)

| Feature | Status |
|---------|--------|
| Live camera → WebSocket → Gemini analysis | Done |
| Streaming AI responses (chunk messages) | Done |
| Voice guidance (TTS) | Done |
| Voice input (STT) | Done |
| AR overlay (step markers) | Done |
| Multi-step procedure view | Done |
| Session history (local) | Done |
| Offline mode (cached last response) | Done |
| JWT authentication (Supabase) | Done |
| Per-user rate limiting | Done |
| Equipment catalog | Done |
| Settings screen | Done |

---

## v1.1 — Quality & Infra (In Progress)

| Feature | Status |
|---------|--------|
| CI/CD workflows (flutter.yml, security.yml) | Done |
| Kubernetes manifests | Done |
| SECURITY.md + audit log | Done |
| GoDoc + Dartdoc coverage | In Progress |
| Integration test suite | In Progress |
| Prometheus + Grafana dashboards | Done |

---

## v1.2 — UX (Planned)

| Feature | Status | Notes |
|---------|--------|-------|
| Onboarding flow redesign | Planned | First-run experience |
| Dark/light mode toggle | Planned | Currently dark-only |
| Push notifications | Planned | Session completion alerts |
| iPad/tablet layout | Planned | Responsive split view |
| Image annotation sharing | Planned | Export annotated frame as PDF |
| Chat history per session | Planned | Show Q&A thread in session screen |

---

## v1.3 — AI Features (Planned)

| Feature | Status | Notes |
|---------|--------|-------|
| Proactive alerts | Planned | AI warns about visual hazards without being asked |
| Part recognition | Planned | Tap component → AI identifies it |
| Safety checker | Planned | Pre-analysis safety scan before repair |
| Multi-language auto-detect | Planned | Detect user's language from STT |
| Knowledge base UI | Planned | In-app KB browser and search |
| Session summary PDF export | Planned | Full repair report with annotated frames |

---

## v2.0 — Platform (Planned)

| Feature | Status | Notes |
|---------|--------|-------|
| Web app (Flutter Web) | Planned | Parallel to mobile |
| Team workspaces | Planned | Shared session history, team KBs |
| Admin dashboard | Planned | Usage metrics, user management |
| API for third-party integrations | Planned | Webhook on session completion |
| On-premise deploy kit | Planned | Helm chart + docs for self-hosted |

---

## Q2 2026 — Collaboration & Distribution

| Feature | Status | Target |
|---------|--------|--------|
| Multi-user collaboration (shared sessions) | Planned | June 2026 |
| Webhooks API (session completion events) | Planned | May 2026 |
| iOS App Store submission | Planned | June 2026 |
| Token refresh on 401 in Flutter app | Planned | May 2026 |
| Trivy container scan in CI | Planned | May 2026 |
| Session summary PDF export | Planned | June 2026 |

---

## Q3 2026 — Offline & Voice

| Feature | Status | Target |
|---------|--------|--------|
| Full offline mode (local AI model fallback) | Planned | August 2026 |
| Custom KB upload via app (user-provided manuals) | Planned | July 2026 |
| Voice-only mode (no camera, STT-only guidance) | Planned | September 2026 |
| Semantic search over uploaded manuals (pgvector) | Planned | August 2026 |
| Multi-language auto-detect from STT | Planned | July 2026 |
| Push notifications for session completion | Planned | July 2026 |

---

## Q4 2026 — Enterprise & Scale

| Feature | Status | Target |
|---------|--------|--------|
| Edge deployment (on-premise inference) | Planned | November 2026 |
| Plugin system for custom AI backends | Planned | October 2026 |
| Enterprise features (SSO, audit logs, RBAC) | Planned | December 2026 |
| Admin dashboard (usage metrics, user management) | Planned | November 2026 |
| Helm chart for self-hosted Kubernetes deploy | Planned | October 2026 |
| Flutter Web client (parallel to mobile) | Planned | December 2026 |

---

## Not planned

| Feature | Reason |
|---------|--------|
| Windows/macOS desktop app | Low demand, web app covers the use case |
| Real-time collaboration | Adds complexity, low priority |
| Video recording | Privacy concerns, storage costs |

---

## Suggest a feature

Open a [GitHub Discussion](https://github.com/Franck1120/physicscopilot/discussions) or a [feature request issue](https://github.com/Franck1120/physicscopilot/issues/new?template=feature_request.md).
