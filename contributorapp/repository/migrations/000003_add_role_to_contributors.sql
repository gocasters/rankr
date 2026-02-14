-- +migrate Up
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN
        CREATE TYPE role_enum AS ENUM ('admin', 'user');
    END IF;
END $$;

ALTER TABLE contributors
    ADD COLUMN IF NOT EXISTS role role_enum NOT NULL DEFAULT 'user';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'contributors'
          AND column_name = 'role'
          AND udt_name <> 'role_enum'
    ) THEN
        ALTER TABLE contributors
            ALTER COLUMN role TYPE role_enum
            USING role::role_enum;
    END IF;
END $$;

-- +migrate Down
ALTER TABLE contributors
    DROP COLUMN IF EXISTS role;

DROP TYPE IF EXISTS role_enum;
