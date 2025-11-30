-- +migrate Up
DROP INDEX IF EXISTS webhook_events_webhook_unique_idx;
DROP INDEX IF EXISTS webhook_events_historical_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_global_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE resource_type IS NOT NULL AND resource_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_delivery_unique_idx
ON webhook_events(provider, delivery_id)
WHERE delivery_id IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS webhook_events_global_unique_idx;
DROP INDEX IF EXISTS webhook_events_delivery_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_webhook_unique_idx
ON webhook_events(provider, delivery_id)
WHERE source = 'webhook' AND delivery_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_historical_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE source = 'historical';