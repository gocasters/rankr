-- +migrate Up
-- Details of the scored contribution, so a profile can list the latest
-- scored events with their repository and pull request or issue number.
ALTER TABLE processed_score_events
    ADD COLUMN IF NOT EXISTS repository_id   BIGINT,
    ADD COLUMN IF NOT EXISTS repository_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS subject_type    VARCHAR(20),
    ADD COLUMN IF NOT EXISTS subject_number  INTEGER;

CREATE INDEX IF NOT EXISTS idx_score_events_user_timestamp
    ON processed_score_events (user_id, event_timestamp DESC);

-- +migrate Down
DROP INDEX IF EXISTS idx_score_events_user_timestamp;

ALTER TABLE processed_score_events
    DROP COLUMN IF EXISTS repository_id,
    DROP COLUMN IF EXISTS repository_name,
    DROP COLUMN IF EXISTS subject_type,
    DROP COLUMN IF EXISTS subject_number;
