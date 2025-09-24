-- +migrate Up
DROP TABLE IF EXISTS raw_webhook_events;

CREATE TABLE raw_webhook_events (
                                    id BIGSERIAL PRIMARY KEY,
                                    provider smallint NOT NULL,
                                    hook_id BIGINT not null,
                                    owner TEXT not null,
                                    repo TEXT not null,
                                    delivery_id TEXT NOT NULL,
                                    event_type smallint NOT NULL,
                                    payload_json JSONB NOT NULL,
                                    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                                    CONSTRAINT raw_webhook_events_provider_delivery_id_unique UNIQUE (provider, delivery_id)
);

-- +migrate Down
DROP TABLE IF EXISTS raw_webhook_events;