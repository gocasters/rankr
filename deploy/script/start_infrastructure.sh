#!/usr/bin/env bash
set -e

# Get the project root directory (two levels up from this script)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

BASE_PATH="$PROJECT_ROOT/deploy/infrastructure/base/development/docker-compose.base.yml"
PG_PATH="$PROJECT_ROOT/deploy/infrastructure/postgresql/development/docker-compose.postgresql.yml"
REDIS_PATH="$PROJECT_ROOT/deploy/infrastructure/redis/development/docker-compose.redis.yml"
NATS_PATH="$PROJECT_ROOT/deploy/infrastructure/nats/development/docker-compose.nats.yml"
EMQX_PATH="$PROJECT_ROOT/deploy/infrastructure/emqx/development/docker-compose.emqx.yml"
ENV_FILE="$PROJECT_ROOT/deploy/.env"
ENV_TEMPLATE="$PROJECT_ROOT/deploy/.env.example"
GITIGNORE_FILE="$PROJECT_ROOT/.gitignore"
PROJECT_NAME="rankr-infra"
POSTGRES_CONTAINER="rankr-shared-postgres"

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
    echo "Available commands:"
    echo "  up-base        Start base network and volumes"
    echo "  down-base      Stop and remove base network and volumes"
    echo "  logs-base      Show logs for base"
    echo ""
    echo "  up-postgres    Start PostgreSQL"
    echo "  down-postgres  Stop and remove PostgreSQL"
    echo "  logs-postgres  Show logs for PostgreSQL"
    echo ""
    echo "  up-redis       Start Redis"
    echo "  down-redis     Stop and remove Redis"
    echo "  logs-redis     Show logs for Redis"
    echo ""
    echo "  up-nats        Start NATS"
    echo "  down-nats      Stop and remove NATS"
    echo "  logs-nats      Show logs for NATS"
    echo ""
    echo "  up-emqx        Start EMQX"
    echo "  down-emqx      Stop and remove EMQX"
    echo "  logs-emqx      Show logs for EMQX"
    echo ""
    echo "  up-all         Start all infrastructure components"
    echo "  down-all       Stop all infrastructure components"
    echo "  logs-all       Show logs for all running containers"
    echo ""
    echo "  help           Show this help message"
}

function ensure_env_gitignore() {
    if [ -f "$GITIGNORE_FILE" ]; then
        if ! grep -qxF ".env" "$GITIGNORE_FILE"; then
            echo ".env" >> "$GITIGNORE_FILE"
            echo "Appended .env to $GITIGNORE_FILE to avoid committing secrets."
        fi
    fi
}

function check_env_file() {
    ensure_env_gitignore

    if [ -f "$ENV_FILE" ]; then
        return
    fi

    echo "Warning: .env file not found at $ENV_FILE"
    if [ -f "$ENV_TEMPLATE" ]; then
        echo "Creating $ENV_FILE from template..."
        mkdir -p "$(dirname "$ENV_FILE")"
        cp "$ENV_TEMPLATE" "$ENV_FILE"
        echo "Local .env created from $ENV_TEMPLATE. Please review and update placeholder credentials."
    else
        echo "Template file $ENV_TEMPLATE is missing."
        echo "Create $ENV_FILE manually before running this script."
        exit 1
    fi
}

function up_base() {
    echo "Starting base network and volumes..."
    docker compose -p $PROJECT_NAME -f $BASE_PATH up -d
}

function down_base() {
    echo "Stopping base network and volumes..."
    docker compose -p $PROJECT_NAME -f $BASE_PATH down
}

function logs_base() {
    echo "Showing base logs..."
    docker compose -p $PROJECT_NAME -f $BASE_PATH logs -f
}

function up_postgres() {
    check_env_file
    echo "Starting PostgreSQL..."
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

function up_redis() {
    check_env_file
    echo "Starting Redis..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $REDIS_PATH up -d
}

function down_redis() {
    echo "Stopping Redis..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $REDIS_PATH down
}

function logs_redis() {
    echo "Showing Redis logs..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $REDIS_PATH logs -f
}

function up_nats() {
    check_env_file
    echo "Starting NATS..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $NATS_PATH up -d
}

function down_nats() {
    echo "Stopping NATS..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $NATS_PATH down
}

function logs_nats() {
    echo "Showing NATS logs..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $NATS_PATH logs -f
}

function up_emqx() {
    check_env_file
    echo "Starting EMQX..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $EMQX_PATH up -d
}

function down_emqx() {
    echo "Stopping EMQX..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $EMQX_PATH down
}

function logs_emqx() {
    echo "Showing EMQX logs..."
    docker compose -p $PROJECT_NAME --env-file $ENV_FILE -f $EMQX_PATH logs -f
}

function up_all() {
    up_base
    up_postgres
    up_redis
    up_nats
    up_emqx
    echo "All infrastructure services are up and running!"
}

function down_all() {
    down_emqx
    down_nats
    down_redis
    down_postgres
    down_base
    echo "All infrastructure services are stopped!"
}

function logs_all() {
    echo "Showing logs for all infrastructure services..."
    docker compose -p $PROJECT_NAME logs -f
}

case "$1" in
    up-base)
        up_base
        ;;
    down-base)
        down_base
        ;;
    logs-base)
        logs_base
        ;;
    up-postgres)
        up_postgres
        ;;
    down-postgres)
        down_postgres
        ;;
    logs-postgres)
        logs_postgres
        ;;
    up-redis)
        up_redis
        ;;
    down-redis)
        down_redis
        ;;
    logs-redis)
        logs_redis
        ;;
    up-nats)
        up_nats
        ;;
    down-nats)
        down_nats
        ;;
    logs-nats)
        logs_nats
        ;;
    up-emqx)
        up_emqx
        ;;
    down-emqx)
        down_emqx
        ;;
    logs-emqx)
        logs_emqx
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
    help|--help|-h|"")
        print_help
        ;;
    *)
        echo "Unknown option: $1"
        echo ""
        print_help
        exit 1
        ;;
esac
