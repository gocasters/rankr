#!/bin/bash
set -e

# Default values if environment variables are not set
LEADERBOARDSTAT_USER=${LEADERBOARDSTAT_USER:-leaderboardstat_user}
LEADERBOARDSTAT_PASS=${LEADERBOARDSTAT_PASS:-change_me_stat}
CONTRIBUTOR_USER=${CONTRIBUTOR_USER:-contributor_user}
CONTRIBUTOR_PASS=${CONTRIBUTOR_PASS:-change_me_contributor}
PROJECT_USER=${PROJECT_USER:-project_user}
PROJECT_PASS=${PROJECT_PASS:-change_me_project}
WEBHOOK_USER=${WEBHOOK_USER:-webhook_user}
WEBHOOK_PASS=${WEBHOOK_PASS:-change_me_webhook}
LEADERBOARDSCORING_USER=${LEADERBOARDSCORING_USER:-leaderboardscoring_user}
LEADERBOARDSCORING_PASS=${LEADERBOARDSCORING_PASS:-change_me_leaderboardscoring}
AUTH_USER=${AUTH_USER:-auth_user}
AUTH_PASS=${AUTH_PASS:-change_me_auth}
NOTIFAPP_USER=${NOTIFAPP_USER:-notification_user}
NOTIFAPP_PASS=${NOTIFAPP_PASS:-change_me_notification}
REALTIME_USER=${REALTIME_USER:-realtime_user}
REALTIME_PASS=${REALTIME_PASS:-change_me_realtime}
TASK_USER=${TASK_USER:-task_user}
TASK_PASS=${TASK_PASS:-change_me_task}

echo "========================================="
echo "Initializing Rankr microservice databases..."
echo "========================================="

# Function to create user and database
create_service_db() {
    local user=$1
    local pass=$2
    local dbname=$3

    echo "Setting up $dbname..."

    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        -- Create user
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

        -- Create database
        SELECT 'CREATE DATABASE $dbname OWNER $user'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$dbname')
        \gexec

        -- Grant privileges
        GRANT ALL PRIVILEGES ON DATABASE $dbname TO $user;
EOSQL

    echo "$dbname setup completed"
}

# Create all service databases
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
echo "Database initialization completed!"
echo "========================================="
echo "Available service databases:"
echo "  - leaderboardstat_db       → $LEADERBOARDSTAT_USER"
echo "  - contributor_db           → $CONTRIBUTOR_USER"
echo "  - project_db               → $PROJECT_USER"
echo "  - webhook_db               → $WEBHOOK_USER"
echo "  - leaderboardscoring_db    → $LEADERBOARDSCORING_USER"
echo "  - auth_db                  → $AUTH_USER"
echo "  - notification_db          → $NOTIFAPP_USER"
echo "  - realtime_db              → $REALTIME_USER"
echo "  - task_db                  → $TASK_USER"
echo "========================================="
