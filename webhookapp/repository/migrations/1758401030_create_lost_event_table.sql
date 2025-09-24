-- +migrate Up
DROP TABLE IF EXISTS lost_webhook_events;

CREATE TABLE lost_webhook_events (
                                    id BIGSERIAL PRIMARY KEY,
                                    provider smallint NOT NULL,
                                    delivery_id TEXT NOT NULL,
                                    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                                    CONSTRAINT lost_webhook_events_provider_delivery_id_unique UNIQUE (provider, delivery_id)
);

-- +migrate Down
DROP TABLE IF EXISTS lost_webhook_events;