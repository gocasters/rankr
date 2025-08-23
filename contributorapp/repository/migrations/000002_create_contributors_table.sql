CREATE TABLE IF NOT EXISTS contributors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    github_id BIGINT NOT NULL UNIQUE,
    avatar_url TEXT,
    score INTEGER DEFAULT 0,
    rank INTEGER DEFAULT 0,
    contributions INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contributors_score ON contributors(score DESC);
CREATE INDEX IF NOT EXISTS idx_contributors_rank ON contributors(rank);
CREATE INDEX IF NOT EXISTS idx_contributors_github_id ON contributors(github_id);


CREATE OR REPLACE FUNCTION update_last_updated_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_contributors_last_updated 
    BEFORE UPDATE ON contributors 
    FOR EACH ROW 
    EXECUTE FUNCTION update_last_updated_column();