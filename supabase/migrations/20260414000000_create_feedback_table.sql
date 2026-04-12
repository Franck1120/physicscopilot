-- Migration: create feedback table
-- Stores per-step user feedback (thumbs up/down + optional comment).
-- session_id references sessions.id; rows are kept even when the session
-- is soft-deleted so historical feedback is not lost.

CREATE TABLE IF NOT EXISTS feedback (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  TEXT        NOT NULL,
    step_number INTEGER     NOT NULL CHECK (step_number >= 0),
    rating      TEXT        NOT NULL CHECK (rating IN ('positive', 'negative')),
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_feedback_session
        FOREIGN KEY (session_id) REFERENCES sessions (id)
        ON DELETE SET NULL
);

-- Index to look up feedback by session quickly.
CREATE INDEX IF NOT EXISTS idx_feedback_session_id ON feedback (session_id);

-- Enable Row Level Security. No policies are defined here — access is
-- controlled by the service-role key used by the Go server.
ALTER TABLE feedback ENABLE ROW LEVEL SECURITY;
