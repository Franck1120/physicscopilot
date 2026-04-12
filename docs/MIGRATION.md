# Database Migration Guide

This document describes the Supabase/Postgres migration process for
PhysicsCopilot, including how to apply and roll back migrations, naming
conventions, and the current database schema.

---

## Overview

Migrations live in `supabase/migrations/`. Each file is a plain SQL script
applied in filename-alphabetical order by Supabase CLI (`supabase db push`) or
the Supabase dashboard "SQL editor → Run migration". The naming convention
encodes the order and intent:

```
YYYYMMDDHHMMSS_<descriptive_slug>.sql
```

Example: `20260411000000_initial_schema.sql`

---

## Prerequisites

- [Supabase CLI](https://supabase.com/docs/guides/cli) `>= 1.150`
- A Supabase project linked: `supabase link --project-ref <ref>`
- Or a local Supabase stack: `supabase start`

---

## Applying Migrations

### Against a local Supabase instance

```bash
# Start the local stack (Docker required)
supabase start

# Push all pending migrations
supabase db push
```

### Against the hosted project

```bash
# Link once (stores config in supabase/.temp/)
supabase link --project-ref <your-project-ref>

# Apply pending migrations
supabase db push
```

### Manual application (SQL editor)

Copy the content of the migration file and run it in the Supabase dashboard
**SQL editor**. Migrations are idempotent (`IF NOT EXISTS`, `IF EXISTS`) where
possible, so re-running a migration is safe.

---

## Rolling Back a Migration

Supabase does not provide automatic rollback. Each migration that makes
destructive changes must be paired with a hand-written rollback script stored
alongside it:

```
supabase/migrations/20260412000000_add_constraints.sql          # forward
supabase/migrations/20260412000000_add_constraints_rollback.sql # reverse
```

To roll back:

```bash
# Apply the rollback script manually via psql or the SQL editor
psql "$DATABASE_URL" -f supabase/migrations/20260412000000_add_constraints_rollback.sql
```

After applying the rollback, remove the original migration file from version
control so that `supabase db push` does not re-apply it.

---

## Naming Convention

| Segment          | Rule                                              |
|------------------|---------------------------------------------------|
| Timestamp prefix | `YYYYMMDDHHMMSS` UTC, sequential within a day     |
| Slug             | lowercase, underscores, describes the change      |
| Extension        | `.sql`                                            |

Good examples:
- `20260411000000_initial_schema.sql`
- `20260414000000_create_feedback_table.sql`

Bad examples:
- `migration1.sql` — not sortable
- `AddConstraints.sql` — wrong case, no timestamp

---

## Current Schema

### Migration history

| File                                          | Applied    | Description                               |
|-----------------------------------------------|------------|-------------------------------------------|
| `20260411000000_initial_schema.sql`           | 2026-04-11 | Creates `users`, `devices`, `sessions`, `session_steps`; enables RLS |
| `20260412000000_add_constraints.sql`          | 2026-04-12 | Adds unique constraint on `(session_id, step_number)`; adds `idx_sessions_status` |
| `20260413000000_add_session_device_fields.sql`| 2026-04-13 | Adds `device_brand`, `device_model`, `last_activity` to `sessions` |
| `20260414000000_create_feedback_table.sql`    | 2026-04-14 | Creates `feedback` table with RLS enabled |

### Table overview

#### `users`

| Column       | Type         | Notes                          |
|--------------|--------------|--------------------------------|
| `id`         | `uuid`       | PK, generated                  |
| `email`      | `text`       | unique, not null               |
| `plan`       | `text`       | `'free'` or `'pro'`            |
| `created_at` | `timestamptz`| default `now()`                |

RLS policy: `users_own_data` — `id = auth.uid()`

#### `devices`

| Column       | Type         | Notes                          |
|--------------|--------------|--------------------------------|
| `id`         | `uuid`       | PK                             |
| `user_id`    | `uuid`       | FK → `users.id` CASCADE DELETE |
| `brand`      | `text`       |                                |
| `model`      | `text`       |                                |
| `created_at` | `timestamptz`|                                |

RLS policy: `devices_own_data` — `user_id = auth.uid()`

#### `sessions`

| Column            | Type         | Notes                                     |
|-------------------|--------------|-------------------------------------------|
| `id`              | `uuid`       | PK                                        |
| `user_id`         | `uuid`       | FK → `users.id` CASCADE DELETE            |
| `device_id`       | `uuid`       | FK → `devices.id` SET NULL on DELETE      |
| `status`          | `text`       | `'active'` / `'completed'` / `'abandoned'`|
| `problem_detected`| `text`       | nullable                                  |
| `solution_applied`| `text`       | nullable                                  |
| `success`         | `boolean`    | nullable                                  |
| `duration_seconds`| `int`        | nullable                                  |
| `device_brand`    | `text`       | added in migration 20260413               |
| `device_model`    | `text`       | added in migration 20260413               |
| `last_activity`   | `timestamptz`| added in migration 20260413               |
| `created_at`      | `timestamptz`|                                           |

Indexes: `idx_sessions_user_id`, `idx_sessions_device_id`, `idx_sessions_status`,
`sessions_last_activity_idx`

RLS policy: `sessions_own_data` — `user_id = auth.uid()`

#### `session_steps`

| Column        | Type         | Notes                                     |
|---------------|--------------|-------------------------------------------|
| `id`          | `uuid`       | PK                                        |
| `session_id`  | `uuid`       | FK → `sessions.id` CASCADE DELETE         |
| `step_number` | `int`        | unique per session (constraint added migration 20260412) |
| `instruction` | `text`       |                                           |
| `verified`    | `boolean`    | default `false`                           |
| `created_at`  | `timestamptz`|                                           |

Index: `idx_steps_session_id`

RLS policy: `steps_own_data` — session owned by `auth.uid()`

#### `feedback`

| Column        | Type         | Notes                                     |
|---------------|--------------|-------------------------------------------|
| `id`          | `uuid`       | PK                                        |
| `session_id`  | `text`       | FK → `sessions.id` SET NULL on DELETE     |
| `step_number` | `int`        | ≥ 0                                       |
| `rating`      | `text`       | `'positive'` / `'negative'`               |
| `comment`     | `text`       | nullable                                  |
| `created_at`  | `timestamptz`|                                           |

Index: `idx_feedback_session_id`

RLS: enabled. Access controlled by the service-role key used by the Go server.

---

## Adding a New Migration

1. Create a file: `supabase/migrations/YYYYMMDDHHMMSS_<slug>.sql`
2. Write idempotent SQL (`IF NOT EXISTS`, `IF EXISTS`)
3. Add the corresponding `_rollback.sql` for destructive changes
4. Test locally: `supabase db push` against the local stack
5. Review with a peer before merging to `main`
6. Apply to production via CI or `supabase db push` after deploy

---

## Checking Migration Status

```bash
# List applied migrations (Supabase CLI)
supabase migration list

# Or query the supabase_migrations schema directly
psql "$DATABASE_URL" -c "SELECT version, name FROM supabase_migrations.schema_migrations ORDER BY version;"
```
