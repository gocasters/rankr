-- +migrate Up
DROP INDEX IF EXISTS webhook_events_global_unique_idx;

ALTER TABLE webhook_events
ALTER COLUMN resource_id TYPE TEXT USING resource_id::TEXT;

CREATE UNIQUE INDEX webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;

-- +migrate Down
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM webhook_events
        WHERE resource_id IS NOT NULL
        AND resource_id !~ '^[0-9]+$'
    ) THEN
        RAISE EXCEPTION 'Cannot rollback: composite resource_id values exist (e.g., "pr_id:review_id"). Delete these rows first or this migration is one-way.';
    END IF;
END $$;

DROP INDEX IF EXISTS webhook_events_global_unique_idx;

ALTER TABLE webhook_events
ALTER COLUMN resource_id TYPE BIGINT USING resource_id::BIGINT;

CREATE UNIQUE INDEX webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;