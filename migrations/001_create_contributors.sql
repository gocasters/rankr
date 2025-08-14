-- +migrate Up
CREATE TABLE contributors (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    github_id VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index for username
CREATE INDEX idx_contributors_username ON contributors(username);

-- Create index for email
CREATE INDEX idx_contributors_email ON contributors(email);

-- +migrate Down
DROP TABLE IF EXISTS contributors;
