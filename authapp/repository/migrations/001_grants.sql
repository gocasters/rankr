-- +migrate Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS grants (
    id BIGSERIAL PRIMARY KEY,
    subject TEXT NOT NULL,
    object TEXT NOT NULL,
    action TEXT NOT NULL,
    field TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT grants_subject_object_action_unique UNIQUE (subject, object, action)
);

CREATE INDEX IF NOT EXISTS idx_grants_subject_object_action ON grants(subject, object, action);
CREATE INDEX IF NOT EXISTS idx_grants_updated_at ON grants(updated_at DESC);

-- +migrate Down

DROP INDEX IF EXISTS idx_grants_updated_at;
DROP INDEX IF EXISTS idx_grants_subject_object_action;
DROP TABLE IF EXISTS grants;
