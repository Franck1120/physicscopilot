# Incident Response Playbook

This playbook defines how the PhysicsCopilot team detects, responds to, and learns from production incidents.

---

## Severity Levels

| Level | Definition | Response SLA | Example |
|-------|-----------|-------------|---------|
| **P0 — Critical** | Complete service outage; all users impacted; data at risk | Respond in 15 min, fix or mitigate in 1 hour | Server unreachable, database down, security breach |
| **P1 — High** | Major feature broken; significant portion of users impacted | Respond in 30 min, fix or mitigate in 4 hours | WebSocket connections failing, Gemini inference returning errors for all requests |
| **P2 — Medium** | Degraded performance or partial feature failure | Respond in 2 hours, fix in 24 hours | Latency > 5s for AI responses, REST API slow but functional |
| **P3 — Low** | Minor issue; workaround exists | Respond in 1 business day | Single domain KB returning stale data, metrics endpoint intermittently slow |

---

## Incident Response Process

### 1. Detect

Sources of detection:
- Prometheus alertmanager alert (see `infra/monitoring/alerts.yaml`)
- Health check failure notification from Render.com
- User report via GitHub issue
- Failed CI/CD pipeline flagging a regression
- Anomalous Gemini API billing

**First responder actions:**
1. Check `GET /health` — is the server responding?
2. Check Render dashboard (or K8s pod status) — are replicas running?
3. Check Prometheus / Grafana — which metrics are anomalous?
4. Check server logs — what error messages appear?

```bash
# Render: stream logs
render logs --service physicscopilot-server --tail

# Docker Compose
docker compose logs -f server

# Kubernetes
kubectl logs -l app=physicscopilot-server --tail=100 -f
```

---

### 2. Declare

Once the issue is confirmed:

1. Open a **GitHub Issue** titled `[INCIDENT] <brief description> — <date>` with label `incident`.
2. Set the severity level (P0–P3).
3. Assign an **Incident Commander** (IC) — the person who coordinates the response.
4. Post a status update to any applicable status page or communication channel.
5. If P0/P1, page the on-call engineer (see escalation path below).

---

### 3. Respond

The IC coordinates response. Everyone else joins the response channel and focuses on their assigned task (diagnose, fix, communicate).

**Active incident checklist:**

```
[ ] Severity confirmed and declared
[ ] Incident issue opened and linked here
[ ] On-call notified (P0/P1)
[ ] Impact assessed: how many users affected?
[ ] Root cause hypothesis formed
[ ] Fix or mitigation underway
[ ] Status page updated (see communication templates)
[ ] External dependencies checked (Gemini API status, Supabase status)
```

**Useful commands during an active incident:**

```bash
# Check health and active connections
curl -s https://physicscopilot-server.onrender.com/health | jq .

# Check Prometheus metrics
curl -s https://physicscopilot-server.onrender.com/metrics | grep -E "^(ws_|ai_inference|http_requests_total|go_goroutines)"

# Kubernetes: check pod status
kubectl get pods -l app=physicscopilot-server
kubectl describe pod <pod-name>

# Force a rolling restart (clears in-memory state, reconnects DB pool)
kubectl rollout restart deployment/physicscopilot-server

# Render: manual redeploy from last successful build
# Use the Render dashboard → Manual Deploy → Select last successful build
```

---

### 4. Resolve

1. Confirm the fix is live and the service is healthy (`GET /health` returns `"status":"ok"`).
2. Verify the anomalous metrics have returned to baseline.
3. Update the incident issue to `resolved` and record the timeline.
4. Post a resolution update to the status page.
5. Schedule a postmortem within 48 hours (P0/P1) or 7 days (P2).

---

### 5. Postmortem

Use the postmortem template at the end of this document. All P0 and P1 incidents require a written postmortem. P2 postmortems are encouraged but optional.

**Key principles:**
- Blameless: focus on systems and processes, not individuals.
- Factual: use actual log timestamps and metrics, not approximations.
- Actionable: every postmortem must produce at least one concrete follow-up task.

---

## Runbooks for Common Incidents

---

### Runbook: Server down / health check failing

**Symptoms:** `GET /health` returns connection refused, 5xx, or times out. Render health check failing.

**Diagnosis:**

```bash
# 1. Is the process running?
# Render: check dashboard → Service → Events
# K8s:
kubectl get pods -l app=physicscopilot-server
kubectl describe pod <pod-name>   # look for OOMKilled, CrashLoopBackOff

# 2. Check recent logs for panic
kubectl logs <pod-name> --previous   # logs from the crashed container
docker compose logs server | tail -50

# 3. Check resource usage
kubectl top pods -l app=physicscopilot-server
```

**Remediation:**

```bash
# Rolling restart (safest — replaces pods one at a time)
kubectl rollout restart deployment/physicscopilot-server

# If OOMKilled: increase memory limit and GOMEMLIMIT
# Edit infra/k8s/deployment.yaml → resources.limits.memory
# Then: kubectl apply -f infra/k8s/deployment.yaml

# If CrashLoopBackOff due to bad env var:
kubectl set env deployment/physicscopilot-server GEMINI_API_KEY=<new-key>
```

**Escalate if:** Server restarts but keeps crashing within 5 minutes — there is a code regression. Roll back:

```bash
kubectl rollout undo deployment/physicscopilot-server
# Render: redeploy previous build from dashboard
```

---

### Runbook: Database connection pool exhausted

**Symptoms:** REST endpoints return 500. Logs show `pgxpool: all connections are busy`. `ws_active_connections` is normal but REST is failing.

**Diagnosis:**

```bash
# Check pool metrics in logs
kubectl logs -l app=physicscopilot-server | grep -i "pgxpool\|connection\|pool"

# Check active Postgres connections
# Run in Supabase SQL Editor:
SELECT count(*), state, wait_event_type
FROM pg_stat_activity
WHERE datname = current_database()
GROUP BY state, wait_event_type
ORDER BY count DESC;

# Check for long-running queries holding connections
SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state
FROM pg_stat_activity
WHERE state != 'idle' AND query_start < now() - interval '30 seconds'
ORDER BY duration DESC;
```

**Remediation:**

```bash
# 1. Kill stuck queries (replace <pid> with actual PIDs from above)
# Run in Supabase SQL Editor:
SELECT pg_terminate_backend(<pid>);

# 2. Reduce the pool size temporarily if we're hitting Supabase limits
kubectl set env deployment/physicscopilot-server DB_POOL_MAX_CONNS=5

# 3. Rolling restart to clear stale connections
kubectl rollout restart deployment/physicscopilot-server
```

**Long-term fix:** Use the Supabase connection pooler (Transaction mode) instead of the direct connection. Update `DATABASE_URL` to point to the pooler endpoint. See `docs/SCALING.md`.

---

### Runbook: High memory usage (>80% GOMEMLIMIT)

**Symptoms:** `process_resident_memory_bytes` > 80% of container limit. Pod is at risk of OOMKill.

**Diagnosis:**

```bash
# Check memory metrics
curl -s https://physicscopilot-server.onrender.com/metrics | grep -E "^(process_resident|go_mem|go_goroutines)"

# Check for goroutine leak
curl -s https://physicscopilot-server.onrender.com/metrics | grep go_goroutines
# If go_goroutines > 5000, there is likely a goroutine leak

# Get goroutine dump (requires pprof endpoint)
curl -s http://localhost:8080/debug/pprof/goroutine?debug=1 | head -100
```

**Remediation:**

```bash
# 1. Force GC (triggers full collection without restart)
# This requires the pprof endpoint or a server restart

# 2. Rolling restart (fastest fix — flushes heap, reconnects cleanly)
kubectl rollout restart deployment/physicscopilot-server

# 3. If goroutine count is growing over time: this is a leak
# Identify the source from pprof and fix in code. Do not just restart repeatedly.

# 4. Increase GOMEMLIMIT as a stopgap (then investigate root cause)
kubectl set env deployment/physicscopilot-server GOMEMLIMIT=600MiB
# Also increase the K8s memory limit to match
```

---

### Runbook: WebSocket connection flood

**Symptoms:** `ws_active_connections` is very high. Server is slow. Many 1008 close frames in logs. Possible DDoS.

**Diagnosis:**

```bash
# Check current WS connections
curl -s https://physicscopilot-server.onrender.com/health | jq .active_connections

# Check connection metrics
curl -s https://physicscopilot-server.onrender.com/metrics | grep ws_

# Check source IPs (requires access to nginx/load balancer access logs)
# nginx: check /var/log/nginx/access.log
awk '{print $1}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head -20
```

**Remediation:**

```bash
# 1. The per-IP limit (10 connections) is already enforced server-side.
# If a single IP is abusing: block at the load balancer or firewall level.

# nginx: block an IP
# Add to nginx.conf inside the server block:
deny 1.2.3.4;

# Cloudflare: add a WAF rule to block the offending IP or ASN
# Use the Cloudflare dashboard → Security → WAF → Custom Rules

# 2. Reduce the per-IP connection limit temporarily
kubectl set env deployment/physicscopilot-server MAX_CONNECTIONS_PER_IP=3

# 3. If the source is distributed (botnet), enable Cloudflare "Under Attack" mode.
```

---

### Runbook: Rate limiter triggered at scale

**Symptoms:** `http_requests_total{status="429"}` is high. Legitimate users are getting rate-limited. The rate limit is too aggressive for current traffic.

**Diagnosis:**

```bash
# Check rate limit hit rate
curl -s https://physicscopilot-server.onrender.com/metrics \
  | grep 'http_requests_total{.*429'

# Check which endpoints are being hit
curl -s https://physicscopilot-server.onrender.com/metrics \
  | grep 'http_requests_total' | grep -v '#'
```

**Remediation:**

```bash
# 1. Identify if this is legitimate traffic growth or abuse.
#    If legitimate: increase rate limits.
kubectl set env deployment/physicscopilot-server RATE_LIMIT_RPM=120

# 2. If the mobile app is polling too aggressively (e.g., health checks in a loop):
#    fix in the Flutter app and release a new build.

# 3. If abuse: apply IP blocking (see WebSocket flood runbook above).
```

---

### Runbook: KB data corruption

**Symptoms:** `GET /api/domains` returns an empty list or malformed JSON. Server logs show `failed to parse kb/data/problems.json`.

**Diagnosis:**

```bash
# Check what the server loaded
curl -s https://physicscopilot-server.onrender.com/api/domains | jq .total

# Check git history for recent changes to the KB
git log --oneline kb/data/problems.json | head -10

# Validate the JSON locally
python3 -m json.tool kb/data/problems.json > /dev/null && echo "JSON is valid"
```

**Remediation:**

```bash
# 1. Identify the last good commit for the KB file
git log --oneline kb/data/problems.json

# 2. Restore from git
git show <last-good-commit>:kb/data/problems.json > kb/data/problems.json
git add kb/data/problems.json
git commit -m "fix(kb): restore problems.json from <last-good-commit> after corruption"
git push origin main

# 3. The server auto-deploys on push to main (Render). Wait for deploy.
# 4. Verify: curl https://physicscopilot-server.onrender.com/api/domains | jq .total
```

---

## On-Call Escalation Path

| Priority | First contact | Escalate to |
|----------|--------------|------------|
| P0 | Incident Commander (IC) | Project maintainer (@Franck1120) |
| P1 | IC | Project maintainer if not resolved in 1 hour |
| P2/P3 | IC | Next business day |

Contact methods: GitHub @mention in the incident issue, or private message via GitHub.

---

## Communication Templates

### Status page update — incident started

```
[Investigating] We are aware of an issue affecting [feature/service].
Users may experience [symptom]. We are investigating and will provide
an update within [timeframe].

Started: 2026-04-12 14:30 UTC
```

### Status page update — mitigation in place

```
[Monitoring] A mitigation has been applied. [Feature/service] should
be recovering. We are monitoring the situation and will update when
fully resolved.

Mitigated: 2026-04-12 15:00 UTC
```

### Status page update — resolved

```
[Resolved] The incident has been resolved. [Feature/service] is
operating normally. A postmortem will be published within [48h/7 days].

Resolved: 2026-04-12 15:30 UTC
Total impact duration: ~60 minutes
```

### User notification (for significant outages)

```
Subject: Service disruption on 2026-04-12

We experienced a service disruption today between 14:30 and 15:30 UTC
that affected [description of impact].

What happened: [brief, non-technical explanation]
What we did: [brief fix description]
What we are doing to prevent recurrence: [follow-up action]

We apologize for the disruption. If you have questions, please open
a GitHub issue or reach out directly.

— The PhysicsCopilot team
```

---

## Postmortem Template

Copy this template into a new section of the incident GitHub issue.

```markdown
## Postmortem — [Incident Title]

**Date:** 2026-MM-DD
**Severity:** P0 / P1 / P2
**Duration:** X hours Y minutes
**Author:** @username
**Reviewed by:** @username

### Summary

One paragraph describing what happened, the impact, and how it was resolved.

### Timeline

All times in UTC.

| Time | Event |
|------|-------|
| 14:30 | Alert fired — health check failing |
| 14:35 | On-call acknowledged, began investigation |
| 14:50 | Root cause identified |
| 15:00 | Fix deployed |
| 15:30 | Service fully restored |

### Root Cause

Detailed technical explanation of what went wrong and why.

### Contributing Factors

- Factor 1 (e.g., no alerting on DB pool exhaustion)
- Factor 2 (e.g., DB pool size not documented)

### Impact

- X% of users affected
- Y sessions dropped
- Z minutes of downtime

### What Went Well

- Detection was fast (automated alert fired within 2 minutes)
- Rollback procedure was well documented

### What Went Poorly

- Root cause diagnosis took longer than expected
- Status page was not updated promptly

### Action Items

| Action | Owner | Due date |
|--------|-------|---------|
| Add Prometheus alert for DB pool exhaustion | @username | 2026-04-19 |
| Document DB pool tuning in SCALING.md | @username | 2026-04-26 |
```
