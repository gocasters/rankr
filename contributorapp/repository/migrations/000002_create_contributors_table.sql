-- +migrate Up
CREATE TYPE privacy_mode_enum AS ENUM ('real', 'anonymous');
CREATE TABLE IF NOT EXISTS contributors (
    id BIGSERIAL PRIMARY KEY,
    github_id BIGINT UNIQUE,
    github_username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    password TEXT NOT NULL DEFAULT '',
    is_verified BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    privacy_mode privacy_mode_enum NOT NULL DEFAULT 'real',
    display_name VARCHAR(255),
    profile_image TEXT,
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_contributors_github_id ON contributors(github_id);
CREATE INDEX IF NOT EXISTS idx_contributors_github_username ON contributors(github_username);

-- +migrate Down
DROP INDEX IF EXISTS idx_contributors_github_username;
DROP TABLE IF EXISTS contributors;
DROP TYPE IF EXISTS privacy_mode_enum;
