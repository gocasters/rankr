-- +migrate Up
CREATE TABLE notifications
(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(100) NOT NULL CHECK (
        type IN ('info','warning','error','success')),
    status VARCHAR(100) NOT NULL CHECK (
        status IN ('unread','read')),
    created_at TIMESTAMP DEFAULT NOW(),
    read_at TIMESTAMP
);


-- +migrate Down
DROP TABLE IF EXISTS notifications;

CREATE INDEX idx_notify_user_id ON notifications(user_id);
CREATE INDEX idx_notify_status ON notifications(status);