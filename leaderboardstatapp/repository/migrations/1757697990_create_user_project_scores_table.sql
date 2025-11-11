-- +migrate Up
CREATE TABLE IF NOT EXISTS user_project_scores (
   id SERIAL PRIMARY KEY,
   contributor_id BIGINT NOT NULL,
   project_id BIGINT NOT NULL,
   score DOUBLE PRECISION NOT NULL DEFAULT 0,
   timeframe  CHAR(20)  NOT NULL CHECK (
           timeframe IN (
                'daily',
                'weekly',
                'monthly',
                'yearly'
            )
    ),
    time_value CHAR(20),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (contributor_id, project_id, timeframe, time_value)
);
COMMENT ON COLUMN user_project_scores.value IS 'Day ex: today date/yesterday date- Week ex: 1,2,3..- Month ex: 1..12';
COMMENT ON COLUMN user_project_scores.project_id IS '0 = global scores, >0 = specific project scores';

-- +migrate Down
DROP TABLE IF EXISTS user_project_scores;
