-- +migrate Up
DROP INDEX IF EXISTS webhook_events_global_unique_idx;

ALTER TABLE webhook_events
ALTER COLUMN resource_id TYPE TEXT USING resource_id::TEXT;

CREATE UNIQUE INDEX webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS webhook_events_global_unique_idx;

ALTER TABLE webhook_events
ALTER COLUMN resource_id TYPE BIGINT USING resource_id::BIGINT;

CREATE UNIQUE INDEX webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;