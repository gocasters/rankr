-- +migrate Up
-- +migrate StatementBegin
CREATE TYPE role_enum AS ENUM ('admin', 'user');
ALTER TABLE contributors
    ADD COLUMN IF NOT EXISTS role role_enum NOT NULL DEFAULT 'user';

ALTER TABLE contributors
    ALTER COLUMN role TYPE role_enum
    USING role::role_enum;
-- +migrate StatementEnd

-- +migrate Down
ALTER TABLE contributors
    DROP COLUMN IF EXISTS role;

DROP TYPE IF EXISTS role_enum;
