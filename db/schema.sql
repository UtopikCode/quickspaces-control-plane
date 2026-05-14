-- QuickSpaces Control Plane PostgreSQL schema

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS workspaces (
    id TEXT PRIMARY KEY,
    repo TEXT NOT NULL,
    owner TEXT NOT NULL,
    ref TEXT NOT NULL,
    desired_state TEXT NOT NULL,
    actual_state TEXT NOT NULL,
    execution_profile JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS access_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "type" TEXT NOT NULL,
    value TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
