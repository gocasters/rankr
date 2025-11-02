-- +migrate Up
CREATE TABLE IF NOT EXISTS daily_contributor_scores (
    id SERIAL PRIMARY KEY,
    contributor_id BIGINT NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    daily_score BIGINT NOT NULL,
    rank BIGINT NOT NULL,
    timeframe VARCHAR(50) NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    status INT DEFAULT 0,
    -- Ensure one record per contributor per timeframe per day
    CONSTRAINT uniq_daily_contributor_timeframe UNIQUE (contributor_id, timeframe, calculated_at::date)
    );

-- Index for efficient queries
CREATE INDEX idx_daily_scores_contributor_timeframe ON daily_contributor_scores (contributor_id, timeframe);
CREATE INDEX idx_daily_scores_calculated_at ON daily_contributor_scores (calculated_at);
CREATE INDEX idx_daily_scores_timeframe_rank ON daily_contributor_scores (timeframe, rank);

-- +migrate Down
DROP TABLE IF EXISTS daily_contributor_scores;