-- +migrate Up
CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    github_id BIGINT NOT NULL UNIQUE,
    issue_number INT NOT NULL,
    title VARCHAR(500) NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'open',
    repository_name VARCHAR(255) NOT NULL,
    labels TEXT[] DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,
    CONSTRAINT unique_issue_per_repo UNIQUE (issue_number, repository_name)
);

CREATE INDEX idx_tasks_issue_number ON tasks(issue_number);
CREATE INDEX idx_tasks_repository_name ON tasks(repository_name);
CREATE INDEX idx_tasks_state ON tasks(state);
CREATE INDEX idx_tasks_github_id ON tasks(github_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_tasks_github_id;
DROP INDEX IF EXISTS idx_tasks_state;
DROP INDEX IF EXISTS idx_tasks_repository_name;
DROP INDEX IF EXISTS idx_tasks_issue_number;
DROP TABLE IF EXISTS tasks;
