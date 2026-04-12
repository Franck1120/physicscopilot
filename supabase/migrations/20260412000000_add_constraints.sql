-- Migration: add UNIQUE constraint and status index
-- Applied: 2026-04-12

-- Prevent duplicate step numbers within the same session.
ALTER TABLE session_steps
  ADD CONSTRAINT uq_session_steps_step_number UNIQUE (session_id, step_number);

-- Speed up status-based queries (e.g. "list all active sessions").
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
