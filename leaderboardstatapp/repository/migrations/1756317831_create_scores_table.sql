-- +migrate Up
-- Create scores table
CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    contributor_id BIGINT NOT NULL,
    activity VARCHAR(50) NOT NULL,
    score DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL,
    is_deleted BOOLEAN NULL
);

-- +migrate Down
DROP TABLE IF EXISTS scores;