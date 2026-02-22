#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PG_PATH="$PROJECT_ROOT/deploy/infrastructure/postgresql/production/docker-compose.postgresql.yml"
NATS_PATH="$PROJECT_ROOT/deploy/infrastructure/nats/production/docker-compose.nats.yml"
REDIS_PATH="$PROJECT_ROOT/deploy/infrastructure/redis/production/docker-compose.redis.yml"
ENV_FILE="$PROJECT_ROOT/deploy/.env.production"
ENV_TEMPLATE="$PROJECT_ROOT/deploy/.env.production.example"

PROJECT_NAME="rankr-prod"
POSTGRES_CONTAINER="shared-postgres"
NATS_CONTAINER="shared-nats"
REDIS_CONTAINER="shared-redis"
NETWORK_NAME="rankr-prod-network"
VOLUME_NAME="rankr-postgres-prod-data"
NATS_VOLUME_NAME="rankr-nats-prod-data"
REDIS_VOLUME_NAME="rankr-redis-prod-data"

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
    if ! docker volume inspect "$NATS_VOLUME_NAME" &>/dev/null; then
        echo "Creating volume $NATS_VOLUME_NAME..."
        docker volume create "$NATS_VOLUME_NAME"
    fi
    if ! docker volume inspect "$REDIS_VOLUME_NAME" &>/dev/null; then
        echo "Creating volume $REDIS_VOLUME_NAME..."
        docker volume create "$REDIS_VOLUME_NAME"
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

function up_nats() {
    ensure_network
    ensure_volume
    echo "Starting NATS (production)..."
    docker compose -p $PROJECT_NAME -f $NATS_PATH up -d
}

function down_nats() {
    echo "Stopping NATS..."
    docker compose -p $PROJECT_NAME -f $NATS_PATH down
}

function up_redis() {
    ensure_network
    ensure_volume
    echo "Starting Redis (production)..."
    docker compose -p $PROJECT_NAME -f $REDIS_PATH up -d
}

function down_redis() {
    echo "Stopping Redis..."
    docker compose -p $PROJECT_NAME -f $REDIS_PATH down
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
    echo "  up-nats        Start NATS"
    echo "  down-nats      Stop NATS"
    echo "  up-redis       Start Redis"
    echo "  down-redis     Stop Redis"
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
    up_nats
    up_redis
    echo "Production infrastructure is up!"
}

function down_all() {
    down_redis
    down_nats
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
    up-nats)
        up_nats
        ;;
    down-nats)
        down_nats
        ;;
    up-redis)
        up_redis
        ;;
    down-redis)
        down_redis
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