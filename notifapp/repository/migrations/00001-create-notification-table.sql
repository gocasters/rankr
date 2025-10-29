-- +migrate Up
CREATE TABLE IF NOT EXISTS notifications
(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(100) NOT NULL CHECK (
        type IN ('info','warning','error','success')),
    status VARCHAR(100) NOT NULL CHECK (
        status IN ('unread','read')),
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    read_at TIMESTAMP
);


-- +migrate Down
DROP TABLE IF EXISTS notifications;

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_user_status ON notifications(user_id, status);
