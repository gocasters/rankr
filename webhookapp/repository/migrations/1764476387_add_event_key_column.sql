-- +migrate Up
ALTER TABLE webhook_events ADD COLUMN IF NOT EXISTS event_key TEXT;

UPDATE webhook_events
SET event_key = provider || '-' || resource_type || '-' || resource_id || '-' || event_type
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL AND event_key IS NULL;

DROP INDEX IF EXISTS webhook_events_global_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_event_key_unique_idx
ON webhook_events(event_key)
WHERE event_key IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS webhook_events_event_key_unique_idx;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS event_key;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;