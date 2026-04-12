# Database Migration Guide

PhysicsCopilot uses Supabase-managed Postgres. Migrations are SQL files tracked in `supabase/migrations/` and applied via the Supabase CLI.

---

## Migration Strategy

All migrations follow a **zero-downtime rolling approach**:

1. Schema changes are backward-compatible before code is deployed (old code reads new schema without failing).
2. New code is deployed and confirmed healthy.
3. Cleanup migrations (dropping old columns, constraints) run after the rollout is confirmed stable.

This means some migrations are intentionally split into two phases:
- **Phase 1 (additive):** Add new column/table with a default or nullable. Deploy code.
- **Phase 2 (cleanup):** Drop the old column or rename after code is stable.

Never combine a destructive schema change with a feature deploy in a single migration.

---

## Migration Tool

PhysicsCopilot uses the **Supabase CLI** to manage migrations. The CLI integrates directly with Supabase projects and can also drive `golang-migrate` against a raw Postgres database.

### Install the Supabase CLI

```bash
# macOS / Linux (Homebrew)
brew install supabase/tap/supabase

# npm (cross-platform)
npm install -g supabase

# Verify
supabase --version
```

### Link to your project

```bash
# Find your project ref in the Supabase dashboard URL:
# https://supabase.com/dashboard/project/<PROJECT_REF>

supabase link --project-ref <PROJECT_REF>
```

This writes a `.supabase/config.toml` (gitignored) with your project ref.

---

## Running Migrations

### Check current migration status

```bash
supabase migration list
```

Output shows which migrations have been applied and which are pending:

```
  LOCAL      │ REMOTE    │ TIME (UTC)
  ──────────────────────────────────────────────
  20240101000000 │ applied   │ 2024-01-01 00:00:00
  20240215123456 │ applied   │ 2024-02-15 12:34:56
  20260401000000 │ pending   │ not applied
```

### Apply pending migrations

```bash
# Dry run — shows what will be applied without making changes
supabase db push --dry-run

# Apply to the linked Supabase project
supabase db push

# Apply to a local Supabase stack (Docker)
supabase db push --local
```

### Apply migrations with golang-migrate (self-hosted Postgres)

If you are running migrations outside Supabase, set `DATABASE_URL` in your environment first (see `.env.example` for the format), then:

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply all pending migrations
migrate -path supabase/migrations -database "$DATABASE_URL" up

# Apply exactly N migrations
migrate -path supabase/migrations -database "$DATABASE_URL" up 2

# Show current version
migrate -path supabase/migrations -database "$DATABASE_URL" version
```

The `DATABASE_URL` variable follows the standard Postgres DSN format. Store it in `.env` (never commit this file) or in your secrets manager.

---

## Migration File Naming Convention

```
supabase/migrations/<timestamp>_<description>.sql
```

The timestamp is UTC in the format `YYYYMMDDHHMMSS`. The description uses underscores (no spaces, no hyphens).

**Good examples**

```
20260101000000_create_sessions_table.sql
20260115120000_add_domain_column_to_sessions.sql
20260201000000_create_feedback_table.sql
20260301090000_add_index_sessions_device_id.sql
```

**Bad examples — avoid these**

```
migration1.sql                           # no timestamp
20260101-create-sessions.sql            # hyphens not underscores
20260101_CreateSessionsTable.sql        # camelCase — use snake_case
```

### Generate a new migration file

```bash
# Supabase CLI creates an empty file with the correct timestamp
supabase migration new add_feedback_table

# Output: supabase/migrations/20260412143000_add_feedback_table.sql
```

---

## Rolling Back a Migration

### Using Supabase CLI (local dev only)

```bash
# Revert the most recent migration
supabase db reset

# WARNING: db reset drops and recreates the entire local database.
# Only use this in local development — never against production.
```

### Using golang-migrate

```bash
# Roll back 1 migration
migrate -path supabase/migrations -database "$DATABASE_URL" down 1

# Roll back to a specific version (timestamp)
migrate -path supabase/migrations -database "$DATABASE_URL" goto 20260115120000
```

### Manual rollback (production)

If a migration has already been applied to production and cannot be rolled back automatically:

1. Write a compensating migration that undoes the change:
   ```bash
   supabase migration new revert_add_domain_column_to_sessions
   ```
2. In the new file, write the inverse SQL (e.g., `ALTER TABLE sessions DROP COLUMN domain;`).
3. Apply it via the normal `supabase db push` flow.
4. Document the revert in the commit message and link to the incident.

---

## Common Migration Pitfalls

### 1. Table-level locks on large tables

`ALTER TABLE` acquires an `AccessExclusiveLock` that blocks all reads and writes. On tables with millions of rows this can take minutes and cause downtime.

**Safe approach for adding a column to a large table:**

```sql
-- Step 1: Add column as nullable with no default (instant, no lock)
ALTER TABLE sessions ADD COLUMN domain TEXT;

-- Step 2: Backfill in batches (run outside a transaction to avoid long locks)
UPDATE sessions SET domain = 'general' WHERE domain IS NULL AND id > 0 LIMIT 10000;
-- Repeat until no rows remain

-- Step 3 (Phase 2 migration, after code is deployed): add NOT NULL constraint
-- Use a CHECK CONSTRAINT approach to avoid full table scan:
ALTER TABLE sessions ADD CONSTRAINT sessions_domain_not_null CHECK (domain IS NOT NULL) NOT VALID;
ALTER TABLE sessions VALIDATE CONSTRAINT sessions_domain_not_null;
```

### 2. Index creation blocking reads

`CREATE INDEX` locks the table. Use `CREATE INDEX CONCURRENTLY` instead:

```sql
-- Bad: locks the table
CREATE INDEX idx_sessions_device_id ON sessions(device_id);

-- Good: runs in the background, no lock
CREATE INDEX CONCURRENTLY idx_sessions_device_id ON sessions(device_id);
```

Note: `CREATE INDEX CONCURRENTLY` cannot run inside a transaction block. If your migration framework wraps statements in a transaction, execute this in a separate migration file with the transaction wrapper disabled, or use Supabase's `--no-verify` flag.

### 3. Long-running transactions

A migration that touches many rows inside a single transaction holds locks for its entire duration. For bulk updates, use batched updates outside transactions.

### 4. Statement timeouts

Supabase sets a default `statement_timeout` of 60 s. Large migrations may exceed this. For migrations that need more time:

```sql
-- At the top of the migration file
SET statement_timeout = '300s';  -- 5 minutes

-- Your migration SQL here
ALTER TABLE ...
```

### 5. Forgetting to update RLS policies

Whenever a new table or column is added, review and update Row Level Security policies. Every Supabase table must have RLS enabled:

```sql
ALTER TABLE new_table ENABLE ROW LEVEL SECURITY;

CREATE POLICY "users can read own rows" ON new_table
  FOR SELECT USING (auth.uid() = user_id);
```

---

## Schema Version Tracking

Supabase tracks applied migrations in a `supabase_migrations` schema in your Postgres database. You can query the current state:

```sql
SELECT version, name, statements, inserted_at
FROM supabase_migrations.schema_migrations
ORDER BY inserted_at DESC
LIMIT 10;
```

When using `golang-migrate` directly, the version is stored in a `schema_migrations` table in the public schema:

```sql
SELECT version, dirty FROM schema_migrations;
```

A `dirty = true` flag means the last migration failed mid-run. Fix the underlying SQL, then force the version back:

```bash
migrate -path supabase/migrations -database "$DATABASE_URL" force <last_good_version>
```

---

## How to Add a New Table

1. Generate the migration file:
   ```bash
   supabase migration new create_knowledge_entries_table
   ```

2. Write the SQL in the generated file:
   ```sql
   -- supabase/migrations/20260412150000_create_knowledge_entries_table.sql

   CREATE TABLE knowledge_entries (
     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     domain      TEXT NOT NULL,
     problem     TEXT NOT NULL,
     solution    TEXT NOT NULL,
     keywords    TEXT[] NOT NULL DEFAULT '{}',
     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   -- Indexes
   CREATE INDEX idx_knowledge_entries_domain ON knowledge_entries(domain);

   -- RLS (required on every table)
   ALTER TABLE knowledge_entries ENABLE ROW LEVEL SECURITY;

   CREATE POLICY "public read" ON knowledge_entries
     FOR SELECT TO anon, authenticated USING (true);

   CREATE POLICY "service role write" ON knowledge_entries
     FOR ALL TO service_role USING (true) WITH CHECK (true);

   -- Trigger to keep updated_at current
   CREATE TRIGGER set_updated_at
     BEFORE UPDATE ON knowledge_entries
     FOR EACH ROW EXECUTE FUNCTION moddatetime(updated_at);
   ```

3. Test locally:
   ```bash
   supabase db push --local
   supabase db diff  # verify the diff looks correct
   ```

4. Apply to staging, verify, then apply to production:
   ```bash
   supabase link --project-ref <STAGING_REF>
   supabase db push

   supabase link --project-ref <PRODUCTION_REF>
   supabase db push
   ```

5. Add the new table to the Go model and DB query files in `server/internal/db/`.

6. Commit the migration file alongside the Go code changes in the same PR.

---

## How to Add a Column to an Existing Table Safely

For small tables (< 100k rows), a simple `ALTER TABLE` is fine:

```sql
ALTER TABLE sessions ADD COLUMN ended_at TIMESTAMPTZ;
```

For large tables, use the two-phase approach described in the pitfalls section above:

**Migration 1 (immediate, additive):**
```sql
-- 20260412160000_add_ended_at_to_sessions.sql
ALTER TABLE sessions ADD COLUMN ended_at TIMESTAMPTZ;
```

**Deploy code that writes `ended_at`.**

**Migration 2 (cleanup, after deploy confirmed healthy):**
```sql
-- 20260425000000_sessions_ended_at_not_null.sql
-- Only run this after all rows have been backfilled.
ALTER TABLE sessions
  ALTER COLUMN ended_at SET NOT NULL;
```

Always communicate the two-phase plan in the PR description so reviewers understand why the NOT NULL constraint is deferred.
