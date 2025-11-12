#!/bin/bash
set -e

# Default values if environment variables are not set
LEADERBOARDSTAT_USER=${LEADERBOARDSTAT_USER:-leaderboardstat_user}
LEADERBOARDSTAT_PASS=${LEADERBOARDSTAT_PASS:-leaderboardstat_pass}
CONTRIBUTOR_USER=${CONTRIBUTOR_USER:-contributor_user}
CONTRIBUTOR_PASS=${CONTRIBUTOR_PASS:-contributor_pass}
PROJECT_USER=${PROJECT_USER:-project_user}
PROJECT_PASS=${PROJECT_PASS:-project_pass}
WEBHOOK_USER=${WEBHOOK_USER:-webhook_user}
WEBHOOK_PASS=${WEBHOOK_PASS:-webhook_pass}
LEADERBOARDSCORING_USER=${LEADERBOARDSCORING_USER:-leaderboardscoring_user}
LEADERBOARDSCORING_PASS=${LEADERBOARDSCORING_PASS:-leaderboardscoring_pass}

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
                CREATE USER $user WITH PASSWORD '$pass';
                RAISE NOTICE 'User $user created';
            ELSE
                RAISE NOTICE 'User $user already exists';
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

echo "========================================="
echo "Database initialization completed!"
echo "========================================="
echo "Available service databases:"
echo "  - leaderboardstat_db       → $LEADERBOARDSTAT_USER"
echo "  - contributor_db           → $CONTRIBUTOR_USER"
echo "  - project_db               → $PROJECT_USER"
echo "  - webhook_db               → $WEBHOOK_USER"
echo "  - leaderboardscoring_db    → $LEADERBOARDSCORING_USER"
echo "========================================="