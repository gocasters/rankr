-- +migrate Up
CREATE TABLE user_total_scores
(
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL,
    total_score        INT    NOT NULL,
    snapshot_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- So that only one snapshot is recorded for each user_id at any given time.
    UNIQUE (user_id, snapshot_timestamp)
);

-- To quickly find a user's latest snapshot
CREATE INDEX idx_user_total_scores_user_ts ON user_total_scores (user_id, snapshot_timestamp DESC);

-- To quickly retrieve all snapshots within a specified time period
CREATE INDEX idx_user_total_scores_snapshot_ts ON user_total_scores (snapshot_timestamp);

-- +migrate Down
DROP TABLE IF EXISTS user_total_scores;