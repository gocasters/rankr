-- +migrate Up
DROP TABLE IF EXISTS webhook_events;

CREATE TABLE webhook_events (
                                id BIGSERIAL PRIMARY KEY,
                                provider smallint NOT NULL,
                                delivery_id TEXT NOT NULL,
                                event_type smallint NOT NULL,
                                payload BYTEA NOT NULL,
                                received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                                CONSTRAINT webhook_events_provider_delivery_id_unique UNIQUE (provider, delivery_id)
);

-- +migrate Down
DROP TABLE IF EXISTS webhook_events;
