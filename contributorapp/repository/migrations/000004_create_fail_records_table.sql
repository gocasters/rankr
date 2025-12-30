-- +migrate Up
CREATE TABLE IF NOT EXISTS fail_records(
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    record_number INT NOT NULL,
    reason TEXT NOT NULL,
    raw_data JSONB NOT NULL,
    last_error TEXT NOT NULL,
    error_type INT NOT NULL,
    retry_count INT DEFAULT 2,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
    );

CREATE INDEX IF NOT EXISTS idx_fail_records_job_id ON fail_records(job_id);

-- +migrate Down
DROP TABLE IF EXISTS fail_records;
