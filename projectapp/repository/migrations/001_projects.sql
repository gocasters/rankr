-- +migrate Up

-- Enable UUID generation (idempotent on most setups)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Project status enum
CREATE TYPE project_status AS ENUM ('ACTIVE', 'ARCHIVED');

-- VCS provider enum
CREATE TYPE vcs_provider AS ENUM ('GITHUB', 'GITLAB', 'BITBUCKET');

-- projects table
CREATE TABLE IF NOT EXISTS projects (
                                        id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(200) NOT NULL,
    slug                 VARCHAR(120) NOT NULL UNIQUE,
    description          TEXT,
    design_reference_url TEXT,
    git_repo_id          VARCHAR(255),
    repo_provider        vcs_provider,
    status               project_status NOT NULL DEFAULT 'ACTIVE',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at          TIMESTAMPTZ
    );

-- -- Common function to maintain updated_at
-- CREATE OR REPLACE FUNCTION set_updated_at()
-- RETURNS trigger LANGUAGE plpgsql AS '
-- BEGIN
--   NEW.updated_at := now();
--   RETURN NEW;
-- END;
-- ';

-- -- Trigger for projects.updated_at
-- DROP TRIGGER IF EXISTS trg_projects_set_updated_at ON projects;
-- CREATE TRIGGER trg_projects_set_updated_at
--     BEFORE UPDATE ON projects
--     FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Helpful indexes
-- CREATE INDEX IF NOT EXISTS idx_projects_status     ON projects(status);
-- CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);

-- +migrate Down

-- Drop trigger first
-- DROP TRIGGER IF EXISTS trg_projects_set_updated_at ON projects;

-- Drop table
DROP TABLE IF EXISTS projects;
DROP TYPE IF EXISTS project_status;

-- Optionally drop the helper function (only if nothing else depends on it)
-- DROP FUNCTION IF EXISTS set_updated_at();