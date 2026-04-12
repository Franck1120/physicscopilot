-- Add device_brand / device_model columns to sessions so that app-created
-- sessions (from WebSocket connections that bypass the devices table) can
-- store the device string the client sends.
-- last_activity tracks when the session was last touched.
ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS device_brand   text          NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS device_model   text          NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_activity  timestamptz   NOT NULL DEFAULT now();

-- Efficient cleanup queries: sessions where last_activity < cutoff.
CREATE INDEX IF NOT EXISTS sessions_last_activity_idx ON sessions (last_activity);
