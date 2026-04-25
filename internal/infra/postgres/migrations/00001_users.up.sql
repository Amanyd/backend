CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    enrollment_id TEXT NOT NULL UNIQUE,
    rank TEXT NOT NULL CHECK(rank IN ('officer','agniveer')),
    role TEXT NOT NULL CHECK(role IN ('student','instructor')),
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_enrollment_id ON users(enrollment_id);
CREATE INDEX idx_users_role ON users(role);
