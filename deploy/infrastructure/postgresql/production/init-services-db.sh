#!/bin/bash
set -e

LEADERBOARDSTAT_USER=${LEADERBOARDSTAT_USER:?LEADERBOARDSTAT_USER is required}
LEADERBOARDSTAT_PASS=${LEADERBOARDSTAT_PASS:?LEADERBOARDSTAT_PASS is required}
CONTRIBUTOR_USER=${CONTRIBUTOR_USER:?CONTRIBUTOR_USER is required}
CONTRIBUTOR_PASS=${CONTRIBUTOR_PASS:?CONTRIBUTOR_PASS is required}
PROJECT_USER=${PROJECT_USER:?PROJECT_USER is required}
PROJECT_PASS=${PROJECT_PASS:?PROJECT_PASS is required}
WEBHOOK_USER=${WEBHOOK_USER:?WEBHOOK_USER is required}
WEBHOOK_PASS=${WEBHOOK_PASS:?WEBHOOK_PASS is required}
LEADERBOARDSCORING_USER=${LEADERBOARDSCORING_USER:?LEADERBOARDSCORING_USER is required}
LEADERBOARDSCORING_PASS=${LEADERBOARDSCORING_PASS:?LEADERBOARDSCORING_PASS is required}
AUTH_USER=${AUTH_USER:?AUTH_USER is required}
AUTH_PASS=${AUTH_PASS:?AUTH_PASS is required}
NOTIFAPP_USER=${NOTIFAPP_USER:?NOTIFAPP_USER is required}
NOTIFAPP_PASS=${NOTIFAPP_PASS:?NOTIFAPP_PASS is required}
REALTIME_USER=${REALTIME_USER:?REALTIME_USER is required}
REALTIME_PASS=${REALTIME_PASS:?REALTIME_PASS is required}
TASK_USER=${TASK_USER:?TASK_USER is required}
TASK_PASS=${TASK_PASS:?TASK_PASS is required}

echo "========================================="
echo "Initializing Rankr microservice databases (production)..."
echo "========================================="

create_service_db() {
    local user=$1
    local pass=$2
    local dbname=$3

    echo "Setting up $dbname..."

    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        DO \$\$
        BEGIN
            IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '$user') THEN
                EXECUTE format('CREATE USER %I WITH PASSWORD %L', '$user', '$pass');
                RAISE NOTICE 'User $user created';
            ELSE
                EXECUTE format('ALTER USER %I WITH PASSWORD %L', '$user', '$pass');
                RAISE NOTICE 'User $user already existed, password refreshed';
            END IF;
        END
        \$\$;

        SELECT 'CREATE DATABASE $dbname OWNER $user'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$dbname')
        \gexec

        GRANT ALL PRIVILEGES ON DATABASE $dbname TO $user;
EOSQL

    echo "$dbname setup completed"
}

create_service_db "$LEADERBOARDSTAT_USER" "$LEADERBOARDSTAT_PASS" "leaderboardstat_db"
create_service_db "$CONTRIBUTOR_USER" "$CONTRIBUTOR_PASS" "contributor_db"
create_service_db "$PROJECT_USER" "$PROJECT_PASS" "project_db"
create_service_db "$WEBHOOK_USER" "$WEBHOOK_PASS" "webhook_db"
create_service_db "$LEADERBOARDSCORING_USER" "$LEADERBOARDSCORING_PASS" "leaderboardscoring_db"
create_service_db "$AUTH_USER" "$AUTH_PASS" "auth_db"
create_service_db "$NOTIFAPP_USER" "$NOTIFAPP_PASS" "notification_db"
create_service_db "$REALTIME_USER" "$REALTIME_PASS" "realtime_db"
create_service_db "$TASK_USER" "$TASK_PASS" "task_db"

echo "========================================="
echo "Production database initialization completed!"
echo "========================================="