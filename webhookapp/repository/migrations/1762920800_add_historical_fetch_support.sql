-- +migrate Up
ALTER TABLE webhook_events
ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'webhook';

ALTER TABLE webhook_events
ADD COLUMN IF NOT EXISTS resource_type VARCHAR(20);

ALTER TABLE webhook_events
ADD COLUMN IF NOT EXISTS resource_id BIGINT;

ALTER TABLE webhook_events
ALTER COLUMN delivery_id DROP NOT NULL;

ALTER TABLE webhook_events
DROP CONSTRAINT IF EXISTS webhook_events_provider_delivery_id_unique;

DROP INDEX IF EXISTS webhook_events_provider_delivery_id_unique;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_webhook_unique_idx
ON webhook_events(provider, delivery_id)
WHERE source = 'webhook' AND delivery_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_historical_unique_idx
ON webhook_events(provider, resource_type, resource_id, event_type)
WHERE source = 'historical';

-- +migrate Down
DROP INDEX IF EXISTS webhook_events_webhook_unique_idx;
DROP INDEX IF EXISTS webhook_events_historical_unique_idx;

ALTER TABLE webhook_events DROP COLUMN IF EXISTS source;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS resource_type;
ALTER TABLE webhook_events DROP COLUMN IF EXISTS resource_id;