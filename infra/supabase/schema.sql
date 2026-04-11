-- PhysicsCopilot — Supabase Schema
-- Run via: Supabase dashboard > SQL Editor

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email      text UNIQUE NOT NULL,
    plan       text        NOT NULL DEFAULT 'free',  -- 'free' | 'pro'
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Devices (3D printers or other repairable hardware)
CREATE TABLE IF NOT EXISTS devices (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    brand      text NOT NULL,
    model      text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Repair sessions
CREATE TABLE IF NOT EXISTS sessions (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id        uuid REFERENCES devices(id) ON DELETE SET NULL,
    status           text NOT NULL DEFAULT 'active',  -- 'active' | 'completed' | 'abandoned'
    problem_detected text,
    solution_applied text,
    success          boolean,
    duration_seconds int,
    created_at       timestamptz NOT NULL DEFAULT now()
);

-- Step-by-step instructions within a session
CREATE TABLE IF NOT EXISTS session_steps (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    step_number int  NOT NULL,
    instruction text NOT NULL,
    verified   boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Row Level Security
ALTER TABLE users         ENABLE ROW LEVEL SECURITY;
ALTER TABLE devices       ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions      ENABLE ROW LEVEL SECURITY;
ALTER TABLE session_steps ENABLE ROW LEVEL SECURITY;

-- RLS Policies — users can only access their own data
CREATE POLICY "users_own_data"    ON users         FOR ALL USING (id = auth.uid());
CREATE POLICY "devices_own_data"  ON devices       FOR ALL USING (user_id = auth.uid());
CREATE POLICY "sessions_own_data" ON sessions      FOR ALL USING (user_id = auth.uid());
CREATE POLICY "steps_own_data"    ON session_steps FOR ALL
    USING (session_id IN (SELECT id FROM sessions WHERE user_id = auth.uid()));

-- Indexes
CREATE INDEX IF NOT EXISTS idx_devices_user_id    ON devices(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id   ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_device_id ON sessions(device_id);
CREATE INDEX IF NOT EXISTS idx_steps_session_id   ON session_steps(session_id);
