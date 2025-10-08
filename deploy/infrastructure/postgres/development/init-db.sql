-- Create dedicated user for leaderboardstat service (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'leaderboardstat_user') THEN
        CREATE USER leaderboardstat_user WITH PASSWORD 'leaderboard_pass_123';
END IF;
END
$$;

-- Create dedicated database for leaderboardstat (use conditional outside DO block)
SELECT 'CREATE DATABASE leaderboardstat_db OWNER leaderboardstat_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardstat_db')\gexec

-- Create dedicated user for contributor service (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'contributor_user') THEN
        CREATE USER contributor_user WITH PASSWORD 'contributor_pass_123';
END IF;
END
$$;

-- Create dedicated database for contributor
SELECT 'CREATE DATABASE contributor_db OWNER contributor_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'contributor_db')\gexec

-- Create dedicated user for project service (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'project_user') THEN
        CREATE USER project_user WITH PASSWORD 'project_pass_123';
END IF;
END
$$;

-- Create dedicated database for project
SELECT 'CREATE DATABASE project_db OWNER project_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'project_db')\gexec

-- Grant privileges (these will work even if databases already exist)
GRANT ALL PRIVILEGES ON DATABASE leaderboardstat_db TO leaderboardstat_user;
GRANT ALL PRIVILEGES ON DATABASE contributor_db TO contributor_user;
GRANT ALL PRIVILEGES ON DATABASE project_db TO project_user;