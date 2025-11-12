-- +migrate Up
CREATE TABLE IF NOT EXISTS webhook_events (
    id BIGSERIAL PRIMARY KEY,
    provider smallint NOT NULL,
    delivery_id TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload BYTEA NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT webhook_events_provider_delivery_id_unique UNIQUE (provider, delivery_id)
);

-- +migrate Down
DROP TABLE IF EXISTS webhook_events;