-- ==================================================
-- Rankr Microservices PostgreSQL Initialization Script
-- ==================================================
-- This script:
--   - Creates a dedicated user & database for each service
--   - Ensures idempotency (safe to re-run multiple times)
--   - Grants full privileges to each respective user
-- ==================================================

\echo '========================================='
\echo 'Initializing Rankr microservice databases...'
\echo '========================================='

-- ==================================================
-- 1. Leaderboard Stat Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'leaderboardstat_user') THEN
        CREATE USER leaderboardstat_user WITH PASSWORD 'leaderboard_pass_123';
        RAISE NOTICE 'User leaderboardstat_user created';
ELSE
        RAISE NOTICE 'User leaderboardstat_user already exists';
END IF;
END
$$;

SELECT 'CREATE DATABASE leaderboardstat_db OWNER leaderboardstat_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardstat_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE leaderboardstat_db TO leaderboardstat_user;

\echo '✔ leaderboardstat_db ready (owner: leaderboardstat_user)'

-- ==================================================
-- 2. Contributor Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'contributor_user') THEN
        CREATE USER contributor_user WITH PASSWORD 'contributor_pass_123';
        RAISE NOTICE 'User contributor_user created';
ELSE
        RAISE NOTICE 'User contributor_user already exists';
END IF;
END
$$;

SELECT 'CREATE DATABASE contributor_db OWNER contributor_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'contributor_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE contributor_db TO contributor_user;

\echo '✔ contributor_db ready (owner: contributor_user)'

-- ==================================================
-- 3. Project Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'project_user') THEN
        CREATE USER project_user WITH PASSWORD 'project_pass_123';
        RAISE NOTICE 'User project_user created';
ELSE
        RAISE NOTICE 'User project_user already exists';
END IF;
END
$$;

SELECT 'CREATE DATABASE project_db OWNER project_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'project_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE project_db TO project_user;

\echo '✔ project_db ready (owner: project_user)'

-- ==================================================
-- 4. Webhook Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'webhook_user') THEN
        CREATE USER webhook_user WITH PASSWORD 'webhook_pass_123';
        RAISE NOTICE 'User webhook_user created';
ELSE
        RAISE NOTICE 'User webhook_user already exists';
END IF;
END
$$;

SELECT 'CREATE DATABASE webhook_db OWNER webhook_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'webhook_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE webhook_db TO webhook_user;

\echo '✔ webhook_db ready (owner: webhook_user)'

-- ==================================================
-- 5. Leaderboard Scoring Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'leaderboardscoring_user') THEN
        CREATE USER leaderboardscoring_user WITH PASSWORD 'leaderboardscoring_pass_123';
        RAISE NOTICE 'User leaderboardscoring_user created';
ELSE
        RAISE NOTICE 'User leaderboardscoring_user already exists';
END IF;
END
$$;

SELECT 'CREATE DATABASE leaderboardscoring_db OWNER leaderboardscoring_user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardscoring_db')\gexec

GRANT ALL PRIVILEGES ON DATABASE leaderboardscoring_db TO leaderboardscoring_user;

\echo '✔ leaderboardscoring_db ready (owner: leaderboardscoring_user)'

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
