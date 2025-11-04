-- +migrate Up
CREATE TABLE IF NOT EXISTS user_scores (
   user_id BIGINT PRIMARY KEY,
   score DOUBLE PRECISION NOT NULL DEFAULT 0
   updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_scores (
    user_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, project_id)
);

-- +migrate Down
DROP TABLE IF EXISTS user_scores;
DROP TABLE IF EXISTS project_scores;
