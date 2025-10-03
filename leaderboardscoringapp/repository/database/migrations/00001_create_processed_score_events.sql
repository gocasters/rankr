-- NOTE:
-- Currently using a CHECK constraint on event_type for flexibility.
-- In the future, this could be migrated to a PostgreSQL ENUM type for:
--   - Stronger type safety at the database level
--   - More efficient storage and indexing compared to VARCHAR
--   - Faster comparisons on queries involving event_type
-- Trade-off: ENUMs are harder to extend (ALTER TYPE required), so for now
-- CHECK is chosen to keep schema changes simpler.

-- +migrate Up
CREATE TABLE processed_score_events
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(100) NOT NULL,
    event_type      VARCHAR(50)  NOT NULL CHECK (
        event_type IN (
                       'pull_request_opened',
                       'pull_request_closed',
                       'pull_request_review',
                       'issue_opened',
                       'issue_closed',
                       'issue_comment',
                       'commit_push'
            )
        ),
    event_timestamp TIMESTAMP    NOT NULL,
    score_delta     BIGINT          NOT NULL,
    processed_at    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_score_events_user_id ON processed_score_events (user_id);
CREATE INDEX idx_score_events_event_type ON processed_score_events (event_type);
CREATE INDEX idx_score_events_event_timestamp ON processed_score_events (event_timestamp);

-- +migrate Down
DROP TABLE IF EXISTS processed_score_events;