-- ==================================================
-- Rankr Microservices PostgreSQL Initialization Script
-- ==================================================

-- Load users and passwords from environment variables
\set LEADERBOARDSTAT_USER :'LEADERBOARDSTAT_USER'
\set LEADERBOARDSTAT_PASS :'LEADERBOARDSTAT_PASS'
\set CONTRIBUTOR_USER :'CONTRIBUTOR_USER'
\set CONTRIBUTOR_PASS :'CONTRIBUTOR_PASS'
\set PROJECT_USER :'PROJECT_USER'
\set PROJECT_PASS :'PROJECT_PASS'
\set WEBHOOK_USER :'WEBHOOK_USER'
\set WEBHOOK_PASS :'WEBHOOK_PASS'
\set LEADERBOARDSCORING_USER :'LEADERBOARDSCORING_USER'
\set LEADERBOARDSCORING_PASS :'LEADERBOARDSCORING_PASS'

\echo '========================================='
\echo 'Initializing Rankr microservice databases...'
\echo '========================================='

-- ==================================================
-- 1. Leaderboard Stat Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'LEADERBOARDSTAT_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'LEADERBOARDSTAT_USER',
            :'LEADERBOARDSTAT_PASS'
        );
        RAISE NOTICE 'User % created', :'LEADERBOARDSTAT_USER';
ELSE
        RAISE NOTICE 'User % already exists', :'LEADERBOARDSTAT_USER';
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE leaderboardstat_db OWNER %I',
               :'LEADERBOARDSTAT_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardstat_db') \gexec;

GRANT ALL PRIVILEGES ON DATABASE leaderboardstat_db TO :'LEADERBOARDSTAT_USER';

\echo 'leaderboardstat_db ready (owner: leaderboardstat_user)'

-- ==================================================
-- 2. Contributor Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'CONTRIBUTOR_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'CONTRIBUTOR_USER',
            :'CONTRIBUTOR_PASS'
        );
        RAISE NOTICE 'User % created', :'CONTRIBUTOR_USER';
ELSE
        RAISE NOTICE 'User % already exists', :'CONTRIBUTOR_USER';
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE contributor_db OWNER %I',
               :'CONTRIBUTOR_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'contributor_db') \gexec;

GRANT ALL PRIVILEGES ON DATABASE contributor_db TO :'CONTRIBUTOR_USER';

\echo 'contributor_db ready (owner: contributor_user)'

-- ==================================================
-- 3. Project Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'PROJECT_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'PROJECT_USER',
            :'PROJECT_PASS'
        );
        RAISE NOTICE 'User % created', :'PROJECT_USER';
ELSE
        RAISE NOTICE 'User % already exists', :'PROJECT_USER';
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE project_db OWNER %I',
               :'PROJECT_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'project_db') \gexec;

GRANT ALL PRIVILEGES ON DATABASE project_db TO :'PROJECT_USER';

\echo 'project_db ready (owner: project_user)'

-- ==================================================
-- 4. Webhook Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'WEBHOOK_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'WEBHOOK_USER',
            :'WEBHOOK_PASS'
        );
        RAISE NOTICE 'User % created', :'WEBHOOK_USER';
ELSE
        RAISE NOTICE 'User % already exists', :'WEBHOOK_USER';
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE webhook_db OWNER %I',
               :'WEBHOOK_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'webhook_db') \gexec;

GRANT ALL PRIVILEGES ON DATABASE webhook_db TO :'WEBHOOK_USER';

\echo 'webhook_db ready (owner: webhook_user)'

-- ==================================================
-- 5. Leaderboard Scoring Service
-- ==================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'LEADERBOARDSCORING_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'LEADERBOARDSCORING_USER',
            :'LEADERBOARDSCORING_PASS'
        );
        RAISE NOTICE 'User % created', :'LEADERBOARDSCORING_USER';
ELSE
        RAISE NOTICE 'User % already exists', :'LEADERBOARDSCORING_USER';
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE leaderboardscoring_db OWNER %I',
               :'LEADERBOARDSCORING_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'leaderboardscoring_db') \gexec;

GRANT ALL PRIVILEGES ON DATABASE leaderboardscoring_db TO :'LEADERBOARDSCORING_USER';

\echo 'leaderboardscoring_db ready (owner: leaderboardscoring_user)'

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
