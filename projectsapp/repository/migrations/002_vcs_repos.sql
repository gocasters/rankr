-- +migrate Up

-- Enums for VCS meta
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vcs_provider') THEN
CREATE TYPE vcs_provider AS ENUM ('GITHUB', 'GITLAB', 'BITBUCKET');
END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vcs_visibility') THEN
CREATE TYPE vcs_visibility AS ENUM ('PUBLIC', 'PRIVATE', 'INTERNAL');
END IF;
END$$;

-- vcs_repos table
CREATE TABLE IF NOT EXISTS vcs_repos (
                                         id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider            vcs_provider NOT NULL,
    provider_repo_id    TEXT NOT NULL,
    owner               TEXT NOT NULL,
    name                TEXT NOT NULL,
    remote_url          TEXT NOT NULL,
    default_branch      TEXT,
    visibility          vcs_visibility NOT NULL,
    installation_id     TEXT,
    last_synced_at      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_vcs_repo UNIQUE (project_id, provider, provider_repo_id)
    );

-- Reuse set_updated_at() if present; safe even if already exists
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
RETURN NEW;
END$$;

-- Trigger for vcs_repos.updated_at
DROP TRIGGER IF EXISTS trg_vcs_repos_set_updated_at ON vcs_repos;
CREATE TRIGGER trg_vcs_repos_set_updated_at
    BEFORE UPDATE ON vcs_repos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_vcs_repos_project     ON vcs_repos(project_id);
CREATE INDEX IF NOT EXISTS idx_vcs_repos_owner_name  ON vcs_repos(owner, name);

-- +migrate Down

-- Drop trigger first
DROP TRIGGER IF EXISTS trg_vcs_repos_set_updated_at ON vcs_repos;

-- Drop table
DROP TABLE IF EXISTS vcs_repos;

-- Drop enums if no longer used
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_type t
    JOIN pg_depend d ON d.refobjid = t.oid
    WHERE t.typname = 'vcs_provider'
      AND d.classid = 'pg_type'::regclass
      AND d.refobjid = t.oid
      AND d.objid <> 0
  ) THEN
DROP TYPE IF EXISTS vcs_provider;
END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_type t
    JOIN pg_depend d ON d.refobjid = t.oid
    WHERE t.typname = 'vcs_visibility'
      AND d.classid = 'pg_type'::regclass
      AND d.refobjid = t.oid
      AND d.objid <> 0
  ) THEN
DROP TYPE IF EXISTS vcs_visibility;
END IF;
END$$;
