# Backup Strategy

This document covers automated and manual backup procedures for the
PhysicsCopilot data layer, including the Postgres database and the
knowledge-base (KB) file store.

## Automated Database Backup — Supabase PITR

When using the hosted Supabase database, Point-in-Time Recovery (PITR) is
available on the Pro plan and above:

- **Frequency** — continuous WAL archiving; snapshots every 24 hours.
- **Retention** — 7 days by default (configurable to 30 days).
- **Activation** — enable in the Supabase dashboard under
  _Project Settings → Database → Point in Time Recovery_.

To restore to a specific moment:

1. Open the Supabase dashboard.
2. Navigate to _Database → Backups_.
3. Select the target timestamp and click **Restore**.
4. Monitor the restoration progress; the project will be briefly unavailable.

## Manual Backup via pg_dump

For self-hosted Postgres or when you need a portable dump:

```bash
# Full schema + data dump (plain SQL)
pg_dump \
  --no-owner \
  --no-acl \
  --format=plain \
  "$DATABASE_URL" \
  > backup_$(date +%Y%m%dT%H%M%S).sql

# Compressed custom-format dump (recommended for large databases)
pg_dump \
  --no-owner \
  --no-acl \
  --format=custom \
  "$DATABASE_URL" \
  -f backup_$(date +%Y%m%dT%H%M%S).dump
```

Store the resulting file in a secure, off-site location (e.g. S3, GCS, or an
encrypted network drive). Rotate backups and keep at least 3 copies.

## Restore Procedure

```bash
# From a plain SQL dump
psql "$TARGET_DATABASE_URL" < backup_YYYYMMDDTHHMMSS.sql

# From a custom-format dump
pg_restore \
  --no-owner \
  --no-acl \
  --dbname="$TARGET_DATABASE_URL" \
  backup_YYYYMMDDTHHMMSS.dump
```

After restoration, verify row counts in critical tables:

```sql
SELECT COUNT(*) FROM sessions;
SELECT COUNT(*) FROM feedback;
```

## Knowledge Base Backup

The knowledge base is stored as JSON files under the directory specified by
`KB_DATA_DIR` (default: `kb/`). These files are the source of truth for the
RAG service.

```bash
# Archive the entire KB directory
tar -czf kb_backup_$(date +%Y%m%dT%H%M%S).tar.gz "${KB_DATA_DIR:-kb/}"
```

Recommended practice:

- Keep the KB directory under version control (already in this repo under
  `kb/`).
- For large KB datasets, store snapshots in object storage alongside DB dumps.
- Tag KB releases to match application releases so the correct KB version can
  be restored alongside a specific server version.

## Backup Integrity Verification

Always verify a backup before relying on it for disaster recovery:

```bash
# Verify a pg_dump custom-format file (lists its table of contents)
pg_restore --list backup_YYYYMMDDTHHMMSS.dump | head -40

# Test restore into a temporary database
createdb physicscopilot_test_restore
pg_restore \
  --no-owner \
  --no-acl \
  --dbname=physicscopilot_test_restore \
  backup_YYYYMMDDTHHMMSS.dump
psql physicscopilot_test_restore -c "SELECT COUNT(*) FROM sessions;"
dropdb physicscopilot_test_restore

# Verify KB archive
tar -tzf kb_backup_YYYYMMDDTHHMMSS.tar.gz | wc -l
```

Schedule integrity checks at least monthly and after every major deployment.

## Backup Schedule Recommendation

| Data                  | Method          | Frequency   | Retention |
|-----------------------|-----------------|-------------|-----------|
| Postgres (Supabase)   | PITR            | Continuous  | 7–30 days |
| Postgres (self-hosted)| pg_dump         | Daily       | 30 days   |
| Knowledge base (JSON) | tar + git tag   | Per release | Indefinite|
