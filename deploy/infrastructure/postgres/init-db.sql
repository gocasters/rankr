-- Create users for each service
CREATE USER leaderboardstat_user WITH PASSWORD 'leaderboard_pass_123';
CREATE USER contributor_user WITH PASSWORD 'contributor_pass_123';
CREATE USER project_user WITH PASSWORD 'project_pass_123';

-- Create schemas for each service and set the owners
CREATE SCHEMA IF NOT EXISTS leaderboardstat_schema AUTHORIZATION leaderboardstat_user;
CREATE SCHEMA IF NOT EXISTS contributor_schema AUTHORIZATION contributor_user;
CREATE SCHEMA IF NOT EXISTS project_schema AUTHORIZATION project_user;

-- Grant additional permissions to ensure they can create tables
GRANT ALL PRIVILEGES ON SCHEMA leaderboardstat_schema TO leaderboardstat_user;
GRANT ALL PRIVILEGES ON SCHEMA contributor_schema TO contributor_user;
GRANT ALL PRIVILEGES ON SCHEMA project_schema TO project_user;

-- Set default search path for each user
ALTER USER leaderboardstat_user SET search_path TO leaderboardstat_schema;
ALTER USER contributor_user SET search_path TO contributor_schema;
ALTER USER project_user SET search_path TO project_schema;

-- Set default privileges so users can create tables in their schemas
ALTER DEFAULT PRIVILEGES FOR USER leaderboardstat_user IN SCHEMA leaderboardstat_schema GRANT ALL ON TABLES TO leaderboardstat_user;
ALTER DEFAULT PRIVILEGES FOR USER contributor_user IN SCHEMA contributor_schema GRANT ALL ON TABLES TO contributor_user;
ALTER DEFAULT PRIVILEGES FOR USER project_user IN SCHEMA project_schema GRANT ALL ON TABLES TO project_user;