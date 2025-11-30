-- +migrate Up
-- Create scores table
CREATE TABLE IF NOT EXISTS scores (
    id SERIAL PRIMARY KEY,
    contributor_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    activity VARCHAR(50),
    score DECIMAL(10, 2) NOT NULL,
    rank INT,
    earned_at  TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_deleted BOOLEAN NULL,
    status INT DEFAULT 0
);

-- +migrate Down
DROP TABLE IF EXISTS scores;
