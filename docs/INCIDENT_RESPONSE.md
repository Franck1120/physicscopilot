# Incident Response

This document defines the severity levels, first-response procedures, and
post-mortem template for PhysicsCopilot production incidents.

## Severity Levels

| Level | Name     | Description                                                    | Response SLA |
|-------|----------|----------------------------------------------------------------|--------------|
| SEV-1 | Critical | Full service outage; all users impacted; data loss possible    | 15 minutes   |
| SEV-2 | High     | Significant degradation; majority of users impacted           | 30 minutes   |
| SEV-3 | Medium   | Partial outage or degraded performance; subset of users        | 2 hours      |
| SEV-4 | Low      | Minor issue; cosmetic or edge-case; workaround available       | 1 business day|

## First Response by Problem Type

### Database Unreachable

Symptoms: `db_status: "unavailable"` in `/health`, 500 errors on session
endpoints, log entries with `DB init failed`.

1. Check `GET /health` — confirm `db_status` field.
2. Verify `DATABASE_URL` is set and reachable:
   ```bash
   psql "$DATABASE_URL" -c "SELECT 1;"
   ```
3. Check Supabase dashboard for outage notices.
4. If the DB is down: the server continues operating without persistence
   (sessions are held in memory). Restart is not required.
5. Once DB is restored: the server will reconnect automatically on the next
   request attempt. If it does not, perform a rolling restart.

### Memory Exhausted / OOM Kill

Symptoms: process exits unexpectedly, `high memory usage` warnings in logs,
`memory_mb` in `/health` approaching container limits.

1. Check current memory: `GET /health` → `memory_mb`.
2. Review logs for `high memory usage` warnings (logged when heap > 80 % of
   `GOMEMLIMIT`).
3. Set or reduce `GOMEMLIMIT` to trigger Go's GC earlier:
   ```bash
   GOMEMLIMIT=512MiB  # adjust to your container limit
   ```
4. Scale horizontally (see `docs/SCALING.md`) to reduce per-instance load.
5. If OOM kill is imminent: restart the instance and investigate the root cause
   via Prometheus `mem_heap_alloc_bytes` metric.

### API Errors (5xx Spike)

Symptoms: elevated `http_requests_total{status="5xx"}` in Prometheus,
error alerts from uptime monitors.

1. Check `GET /health` for overall status.
2. Tail application logs for stack traces and error context.
3. Check the Prometheus dashboard for which `path` and `method` are failing.
4. Common causes:
   - DB timeout → see "Database Unreachable" above.
   - AI backend rate limit → check `AI_PROVIDER` API quota.
   - Panic recovered → look for `panic recovered` in logs; deploy a hotfix.
5. If the error rate is above 10 %: consider rolling back to the last known
   good deployment.

### WebSocket Connection Storm

Symptoms: `active_connections` in `/health` rising beyond normal baseline,
increased CPU, clients failing to connect.

1. Check `active_connections` via `GET /health`.
2. The server enforces per-user connection limits (`USER_CONN_LIMIT`, default 3).
3. If a single IP is causing the storm: add a temporary block at the load
   balancer.
4. Restart the instance to close all WebSocket connections as a last resort.

## Runbook References

| Topic              | Reference                        |
|--------------------|----------------------------------|
| Health check       | `GET /health` endpoint           |
| API logs           | stdout / your log aggregator     |
| Metrics dashboard  | `GET /metrics` (Prometheus)      |
| Scaling            | `docs/SCALING.md`                |
| Backup / restore   | `docs/BACKUP.md`                 |
| Security contacts  | `SECURITY_CONTACTS.md`           |

## Post-Mortem Template

Use this template within 48 hours of resolving a SEV-1 or SEV-2 incident.

```markdown
# Post-Mortem: <Short title>

**Date:** YYYY-MM-DD
**Severity:** SEV-X
**Duration:** HH:MM (from first alert to full resolution)
**Author(s):**

## Summary

One-paragraph description of what happened and the user impact.

## Timeline

| Time (UTC) | Event |
|------------|-------|
| HH:MM      | First alert received |
| HH:MM      | Root cause identified |
| HH:MM      | Mitigation applied |
| HH:MM      | Full resolution confirmed |

## Root Cause

Detailed technical explanation of the root cause.

## Contributing Factors

- Factor 1
- Factor 2

## Resolution

Steps taken to resolve the incident.

## Action Items

| Action | Owner | Due date |
|--------|-------|----------|
| ...    | ...   | ...      |

## Lessons Learned

What went well, what could be improved, what was surprising.
```
