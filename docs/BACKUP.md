# Backup and Disaster Recovery

This document describes what data PhysicsCopilot stores, how it is backed up, and how to restore it after data loss or infrastructure failure.

---

## What Data to Back Up

| Data | Location | Criticality | Backup method |
|------|----------|-------------|---------------|
| Session history | Postgres (Supabase) | High | Supabase automated backups + WAL archiving |
| Feedback records | Postgres (Supabase) | High | Supabase automated backups + WAL archiving |
| Knowledge base | `kb/data/problems.json` | Medium | Git-tracked — no extra steps |
| Server config | Environment variables (Render / K8s secrets) | High | Render dashboard export / K8s etcd backup |
| App signing keys | Android keystore + iOS certificates | Critical | Encrypted off-site storage (1Password / Bitwarden) |

---

## Postgres Backup Strategies

### 1. Supabase Automated Backups (recommended for most deployments)

Supabase automatically backs up your database on every plan:

| Plan | Backup frequency | Retention | Point-in-time recovery |
|------|-----------------|-----------|------------------------|
| Free | Daily | 7 days | No |
| Pro | Daily | 30 days | Yes (last 7 days) |
| Team | Daily | 90 days | Yes (last 30 days) |

To restore from a Supabase backup:
1. Go to the Supabase dashboard → your project → **Settings → Database → Backups**.
2. Select a backup point and click **Restore**.
3. Supabase creates a new project with the restored data. Update `SUPABASE_URL` in your environment to point to the restored project.

### 2. Manual pg_dump (on-demand or scheduled)

Use `pg_dump` to create a portable SQL dump of the database. The `DATABASE_URL` environment variable must be set.

```bash
# Full dump (all tables, schema + data)
pg_dump "$DATABASE_URL" \
  --format=custom \
  --no-acl \
  --no-owner \
  --file="physicscopilot-$(date +%Y%m%d-%H%M%S).dump"

# Schema only (useful for staging setup)
pg_dump "$DATABASE_URL" \
  --schema-only \
  --file="physicscopilot-schema-$(date +%Y%m%d).sql"

# Specific tables only
pg_dump "$DATABASE_URL" \
  --format=custom \
  --table=sessions \
  --table=feedback \
  --file="physicscopilot-data-$(date +%Y%m%d).dump"
```

**Compress and store the dump:**

```bash
gzip physicscopilot-20260412-143000.dump

# Upload to S3-compatible storage
aws s3 cp physicscopilot-20260412-143000.dump.gz \
  s3://physicscopilot-backups/postgres/

# Or to Cloudflare R2
rclone copy physicscopilot-20260412-143000.dump.gz r2:physicscopilot-backups/postgres/
```

**Automate with a cron job:**

```bash
# /etc/cron.d/physicscopilot-backup
0 2 * * * root pg_dump "$DATABASE_URL" --format=custom --file=/backups/physicscopilot-$(date +\%Y\%m\%d).dump && gzip /backups/physicscopilot-$(date +\%Y\%m\%d).dump
```

### 3. WAL Archiving (continuous backups)

Write-Ahead Log archiving provides **point-in-time recovery (PITR)** — restore to any second within the archiving window.

Supabase Pro and Team plans include WAL archiving automatically. For self-hosted Postgres:

```bash
# postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'aws s3 cp %p s3://physicscopilot-wal/%f'
archive_timeout = 60  # archive at least every 60 seconds
```

With WAL archiving, `pg_basebackup` creates the base snapshot and WAL files fill in the gap:

```bash
pg_basebackup \
  --host=<host> \
  --port=5432 \
  --username=replication_user \
  --pgdata=/backups/base \
  --wal-method=stream \
  --checkpoint=fast \
  --progress
```

---

## Knowledge Base Backup

The knowledge base lives in `kb/data/problems.json` and is **fully tracked in Git**. Every commit is a backup point.

No additional backup steps are needed. To restore a previous version of the KB:

```bash
# View KB change history
git log --oneline kb/data/problems.json

# Restore to a specific commit
git show <commit-sha>:kb/data/problems.json > kb/data/problems.json
git commit -m "chore(kb): restore problems.json to <commit-sha>"
```

---

## Recovery Procedures

### Restore from pg_dump

```bash
# Create a new empty database (or use a fresh Supabase project)
# Then restore:
pg_restore \
  --dbname="$DATABASE_URL" \
  --no-acl \
  --no-owner \
  --verbose \
  physicscopilot-20260412-143000.dump

# Verify row counts after restore
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM sessions;"
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM feedback;"
```

### Restore from WAL (PITR)

1. Stop the Postgres server.
2. Replace the data directory with the base backup:
   ```bash
   rm -rf /var/lib/postgresql/data
   cp -a /backups/base /var/lib/postgresql/data
   ```
3. Create a `recovery.conf` (Postgres < 12) or `postgresql.conf` entries (Postgres 12+):
   ```ini
   restore_command = 'aws s3 cp s3://physicscopilot-wal/%f %p'
   recovery_target_time = '2026-04-12 14:30:00 UTC'
   recovery_target_action = 'promote'
   ```
4. Start Postgres — it will replay WAL until the target time and promote to primary.
5. Verify data integrity and update application connection strings.

### Restore from Supabase dashboard

1. Supabase dashboard → project → **Settings → Database → Backups**.
2. Select the restore point and click **Restore**.
3. Supabase provisions a new project (the original remains untouched until you confirm).
4. Test the restored project.
5. Update `SUPABASE_URL`, `SUPABASE_JWT_SECRET`, and `SUPABASE_SERVICE_KEY` in your environment (all change on project restore).
6. Redeploy the server with updated env vars.

---

## RTO / RPO Targets

| Scenario | RPO (data loss tolerance) | RTO (time to restore) |
|----------|--------------------------|----------------------|
| Single table corruption | 24 hours (daily backup) | 1–2 hours |
| Full database loss (Supabase Pro) | Near-zero with PITR | 30–60 minutes |
| Full database loss (Supabase Free) | Up to 24 hours | 2–4 hours |
| Knowledge base corruption | Zero (Git history) | 15 minutes |
| Server config loss | Zero (documented in `.env.example`) | 30 minutes |
| Mobile app signing key loss | — | Potentially weeks (Play Store / App Store resubmission) |

**Mitigation for signing key loss:** Store keystores encrypted in a secrets manager (1Password, Bitwarden, HashiCorp Vault). Never commit them to Git.

---

## Backup Verification (Restore Test Procedure)

A backup that has never been tested is not a backup. Run a restore drill at least **monthly** for production databases.

### Monthly restore test checklist

```
[ ] 1. Identify the latest production backup file/point.
[ ] 2. Provision a fresh Postgres instance (local Docker or a test Supabase project).
[ ] 3. Restore the backup to the test instance.
[ ] 4. Run verification queries:
         SELECT COUNT(*) FROM sessions;
         SELECT COUNT(*) FROM feedback;
         SELECT MAX(created_at) FROM sessions;
[ ] 5. Verify the row count and latest timestamp match expectations.
[ ] 6. Run the server locally pointing at the test instance:
         DATABASE_URL=<test-url> make server-run
[ ] 7. Hit GET /health and POST /api/sessions to confirm the app works.
[ ] 8. Document the result in the backup log (date, backup age, row count, pass/fail).
[ ] 9. Destroy the test instance.
```

### Automated backup integrity check

```bash
#!/usr/bin/env bash
# scripts/verify-backup.sh
# Run after each pg_dump to verify the dump file is not corrupt.

DUMP_FILE="$1"

if [ -z "$DUMP_FILE" ]; then
  echo "Usage: $0 <dump_file>"
  exit 1
fi

# pg_restore --list reads the TOC without restoring — fast integrity check
pg_restore --list "$DUMP_FILE" > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "Backup integrity OK: $DUMP_FILE"
else
  echo "ERROR: Backup is corrupt: $DUMP_FILE"
  exit 1
fi
```

---

## Monitoring Backup Health

### Prometheus alert: backup age

Add this alert to detect when backups stop running:

```yaml
# infra/monitoring/alerts.yaml
- alert: BackupTooOld
  expr: time() - backup_last_success_timestamp_seconds > 86400  # 24h
  for: 1h
  labels:
    severity: critical
  annotations:
    summary: "Database backup has not run in over 24 hours"
```

To expose this metric, a small backup script can write a timestamp to a file that the Node Exporter textfile collector scrapes:

```bash
# After a successful backup:
echo "backup_last_success_timestamp_seconds $(date +%s)" \
  > /var/lib/node_exporter/textfile/backup.prom
```

### Checklist for backup monitoring

- [ ] Daily backup job has a success/failure notification (email, Slack, PagerDuty).
- [ ] Backup file size is checked — a zero-byte dump is an error, not a success.
- [ ] Backup age alert fires if no backup in 25 hours (buffer for timing drift).
- [ ] Monthly restore drill is scheduled as a recurring calendar event.
- [ ] Signing keys are stored in a secrets manager with access logs enabled.
