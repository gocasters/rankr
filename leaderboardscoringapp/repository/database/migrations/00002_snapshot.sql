-- +migrate Up
CREATE TABLE snapshot
(
    id                 BIGSERIAL PRIMARY KEY,
    rank               BIGINT       NOT NULL,
    user_id            VARCHAR(100) NOT NULL,
    total_score        BIGINT       NOT NULL CHECK (total_score >= 0),
    leaderboard_key    VARCHAR(250) NOT NULL,
    snapshot_timestamp TIMESTAMP    NOT NULL DEFAULT NOW(),

    -- ensure only one snapshot per user/leaderboard_key per timestamp
    CONSTRAINT uniq_snapshot_user_leaderboard_key_ts UNIQUE (user_id, leaderboard_key, snapshot_timestamp)
);

-- to quickly find a user's latest snapshot
CREATE INDEX idx_snapshot_user_leaderboard_key_ts_desc
    ON snapshot (user_id, leaderboard_key, snapshot_timestamp DESC);

-- to quickly retrieve all snapshots within a specified time period
CREATE INDEX idx_snapshot_leaderboard_key_ts
    ON snapshot (leaderboard_key, snapshot_timestamp);

-- +migrate Down
DROP TABLE IF EXISTS snapshot;
