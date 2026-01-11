-- +migrate Up
CREATE TYPE status_enum AS ENUM('pending', 'pending_to_queue', 'success', 'failed', 'partial_success', 'processing');

CREATE TABLE IF NOT EXISTS jobs(
    id BIGSERIAL PRIMARY KEY,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    file_path VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_hash TEXT NOT NULL,
    status status_enum NOT NULL DEFAULT 'pending',
    total_records BIGINT DEFAULT 0,
    success_count BIGINT DEFAULT 0,
    fail_count BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_jobs_idempotency_key ON jobs(idempotency_key);

-- +migrate Down
DROP TABLE IF EXISTS jobs;
DROP TYPE IF EXISTS status_enum;
