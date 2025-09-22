-- Create users for each service
CREATE USER rankr_admin WITH SUPERUSER CREATEDB CREATEROLE LOGIN PASSWORD 'rankr_admin_pass';
CREATE USER leaderboardstat_user WITH PASSWORD 'leaderboard_pass_123';
CREATE USER contributor_user WITH PASSWORD 'contributor_pass_123';
CREATE USER project_user WITH PASSWORD 'project_pass_123';

-- Create schemas for each service
CREATE SCHEMA IF NOT EXISTS leaderboardstat_schema;
CREATE SCHEMA IF NOT EXISTS contributor_schema;
CREATE SCHEMA IF NOT EXISTS project_schema;

-- Grant schema ownership to respective users
GRANT ALL ON SCHEMA leaderboardstat_schema TO leaderboardstat_user;
GRANT ALL ON SCHEMA contributor_schema TO contributor_user;
GRANT ALL ON SCHEMA project_schema TO project_user;

-- Grant usage on database
GRANT CONNECT ON DATABASE rankr_db TO leaderboardstat_user;
GRANT CONNECT ON DATABASE rankr_db TO contributor_user;
GRANT CONNECT ON DATABASE rankr_db TO project_user;

-- Grant usage on public schema (for migrations)
GRANT USAGE, CREATE ON SCHEMA public TO leaderboardstat_user;
GRANT USAGE, CREATE ON SCHEMA public TO contributor_user;
GRANT USAGE, CREATE ON SCHEMA public TO project_user;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA leaderboardstat_schema GRANT ALL ON TABLES TO leaderboardstat_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA contributor_schema GRANT ALL ON TABLES TO contributor_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA project_schema GRANT ALL ON TABLES TO project_user;