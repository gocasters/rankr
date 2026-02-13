-- +migrate Up
-- +migrate StatementBegin
CREATE TYPE role_enum AS ENUM ('admin', 'user');
-- +migrate StatementEnd

ALTER TABLE contributors
    ADD COLUMN role role_enum NOT NULL DEFAULT 'user';

-- +migrate Down
ALTER TABLE contributors
    DROP COLUMN IF EXISTS role;

DROP TYPE IF EXISTS role_enum;
