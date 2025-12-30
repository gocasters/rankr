-- +migrate Up
ALTER TABLE contributors
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';

-- +migrate Down
ALTER TABLE contributors
    DROP COLUMN IF EXISTS role;
