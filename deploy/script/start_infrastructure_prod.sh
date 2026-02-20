#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PG_PATH="$PROJECT_ROOT/deploy/infrastructure/postgresql/production/docker-compose.postgresql.yml"
ENV_FILE="$PROJECT_ROOT/deploy/.env.production"
ENV_TEMPLATE="$PROJECT_ROOT/deploy/.env.production.example"

PROJECT_NAME="rankr-prod"
POSTGRES_CONTAINER="shared-postgres"
NETWORK_NAME="rankr-prod-network"
VOLUME_NAME="rankr-postgres-prod-data"

function check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        echo "ERROR: Production .env file not found at $ENV_FILE"
        if [ -f "$ENV_TEMPLATE" ]; then
            echo "Create it from template:"
            echo "  cp $ENV_TEMPLATE $ENV_FILE"
            echo "  nano $ENV_FILE  # Update passwords"
        fi
        exit 1
    fi
}

function ensure_network() {
    if ! docker network inspect "$NETWORK_NAME" &>/dev/null; then
        echo "Creating network $NETWORK_NAME..."
        docker network create "$NETWORK_NAME"
    fi
}

function ensure_volume() {
    if ! docker volume inspect "$VOLUME_NAME" &>/dev/null; then
        echo "Creating volume $VOLUME_NAME..."
        docker volume create "$VOLUME_NAME"
    fi
}

function wait_for_postgres_health() {
    local retries=30
    local delay=2
    local attempt=1

    echo "Waiting for PostgreSQL container to become healthy..."
    while [ $attempt -le $retries ]; do
        local status
        status=$(docker inspect -f '{{.State.Health.Status}}' "$POSTGRES_CONTAINER" 2>/dev/null || echo "starting")

        if [ "$status" == "healthy" ]; then
            echo "PostgreSQL is healthy."
            return 0
        fi

        if [ "$status" == "unhealthy" ]; then
            echo "PostgreSQL healthcheck failed."
            docker logs "$POSTGRES_CONTAINER" || true
            exit 1
        fi

        sleep $delay
        attempt=$((attempt + 1))
    done

    echo "PostgreSQL did not become healthy within $((retries * delay)) seconds."
    docker logs "$POSTGRES_CONTAINER" || true
    exit 1
}

function run_postgres_init_script() {
    echo "Ensuring service databases exist..."
    docker exec "$POSTGRES_CONTAINER" /docker-entrypoint-initdb.d/init-services-db.sh
}

function print_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Production Infrastructure Management"
    echo ""
    echo "Available commands:"
    echo "  up-postgres    Start PostgreSQL"
    echo "  down-postgres  Stop PostgreSQL"
    echo "  logs-postgres  Show PostgreSQL logs"
    echo "  up-all         Start all infrastructure"
    echo "  down-all       Stop all infrastructure"
    echo "  logs-all       Show all logs"
    echo "  status         Show container status"
    echo "  help           Show this help"
}

function up_postgres() {
    check_env_file
    ensure_network
    ensure_volume
    echo "Starting PostgreSQL (production)..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $PG_PATH up -d --build --force-recreate
    wait_for_postgres_health
    run_postgres_init_script
}

function down_postgres() {
    echo "Stopping PostgreSQL..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $PG_PATH down
}

function logs_postgres() {
    echo "Showing PostgreSQL logs..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $PG_PATH logs -f
}

function up_all() {
    up_postgres
    echo "Production infrastructure is up!"
}

function down_all() {
    down_postgres
    echo "Production infrastructure stopped!"
}

function logs_all() {
    logs_postgres
}

function status() {
    echo "Production containers:"
    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "shared-|rankr-prod|NAMES" || echo "No production containers found"
}

case "$1" in
    up-postgres)
        up_postgres
        ;;
    down-postgres)
        down_postgres
        ;;
    logs-postgres)
        logs_postgres
        ;;
    up-all)
        up_all
        ;;
    down-all)
        down_all
        ;;
    logs-all)
        logs_all
        ;;
    status)
        status
        ;;
    help|--help|-h|"")
        print_help
        ;;
    *)
        echo "Unknown option: $1"
        print_help
        exit 1
        ;;
esac