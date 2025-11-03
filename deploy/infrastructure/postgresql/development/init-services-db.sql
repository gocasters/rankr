-- ==================================================
-- Rankr Microservices PostgreSQL Initialization Script
-- ==================================================

\echo '========================================='
\echo 'Initializing Rankr microservice databases...'
\echo '========================================='

-- ==================================================
-- Create Users
-- ==================================================

-- 1. Leaderboard Stat User
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'leaderboardstat_user') THEN
        CREATE USER leaderboardstat_user WITH PASSWORD 'leaderboardstat_pass';
        RAISE NOTICE 'User leaderboardstat_user created';
ELSE
        RAISE NOTICE 'User leaderboardstat_user already exists';
END IF;
END
$$;

-- 2. Contributor User
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'contributor_user') THEN
        CREATE USER contributor_user WITH PASSWORD 'contributor_pass';
        RAISE NOTICE 'User contributor_user created';
ELSE
        RAISE NOTICE 'User contributor_user already exists';
END IF;
END
$$;

-- 3. Project User
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'project_user') THEN
        CREATE USER project_user WITH PASSWORD 'project_pass';
        RAISE NOTICE 'User project_user created';
ELSE
        RAISE NOTICE 'User project_user already exists';
END IF;
END
$$;

-- 4. Webhook User
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'webhook_user') THEN
        CREATE USER webhook_user WITH PASSWORD 'webhook_pass';
        RAISE NOTICE 'User webhook_user created';
ELSE
        RAISE NOTICE 'User webhook_user already exists';
END IF;
END
$$;

-- 5. Leaderboard Scoring User
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'leaderboardscoring_user') THEN
        CREATE USER leaderboardscoring_user WITH PASSWORD 'leaderboardscoring_pass';
        RAISE NOTICE 'User leaderboardscoring_user created';
ELSE
        RAISE NOTICE 'User leaderboardscoring_user already exists';
END IF;
END
$$;

\echo 'All users created successfully'

-- ==================================================
-- Create Databases (must be outside DO blocks)
-- ==================================================

SELECT 'CREATE DATABASE leaderboardstat_db OWNER leaderboardstat_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardstat_db')
    \gexec

SELECT 'CREATE DATABASE contributor_db OWNER contributor_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'contributor_db')
    \gexec

SELECT 'CREATE DATABASE project_db OWNER project_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'project_db')
    \gexec

SELECT 'CREATE DATABASE webhook_db OWNER webhook_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'webhook_db')
    \gexec

SELECT 'CREATE DATABASE leaderboardscoring_db OWNER leaderboardscoring_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardscoring_db')
    \gexec

    \echo 'All databases created successfully'

-- ==================================================
-- Grant Privileges
-- ==================================================

GRANT ALL PRIVILEGES ON DATABASE leaderboardstat_db TO leaderboardstat_user;
GRANT ALL PRIVILEGES ON DATABASE contributor_db TO contributor_user;
GRANT ALL PRIVILEGES ON DATABASE project_db TO project_user;
GRANT ALL PRIVILEGES ON DATABASE webhook_db TO webhook_user;
GRANT ALL PRIVILEGES ON DATABASE leaderboardscoring_db TO leaderboardscoring_user;

\echo 'All privileges granted successfully'

-- ==================================================
-- Summary
-- ==================================================
\echo '========================================='
\echo 'Database initialization completed!'
\echo '========================================='
\echo 'Available service databases and owners:'
\echo '  - leaderboardstat_db       → leaderboardstat_user'
\echo '  - contributor_db           → contributor_user'
\echo '  - project_db               → project_user'
\echo '  - webhook_db               → webhook_user'
\echo '  - leaderboardscoring_db    → leaderboardscoring_user'
\echo '========================================='
\echo 'SECURITY NOTE: Default passwords are used!'
\echo 'Change passwords in production environment.'
\echo '========================================='