#!/usr/bin/env bash
set -e

BASE_PATH="deploy/infrastructure/base/development/docker-compose.base.yml"
PG_PATH="deploy/infrastructure/postgresql/development/docker-compose.postgresql.yml"
REDIS_PATH="deploy/infrastructure/redis/development/docker-compose.redis.yml"
NATS_PATH="deploy/infrastructure/nats/development/docker-compose.nats.yml"
EMQX_PATH="deploy/infrastructure/emqx/development/docker-compose.emqx.yml"
PROJECT_NAME="rankr-infra"

function print_help() {
  echo ""
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
  echo ""
}

function up_base() {
  echo "Starting base network and volumes..."
  docker compose -p $PROJECT_NAME -f $BASE_PATH up -d
}

function down_base() {
  echo "Down base network and volumes..."
  docker compose -p $PROJECT_NAME -f $BASE_PATH down -v
}

function logs_base() {
  echo "Showing logs for base..."
  docker compose -p $PROJECT_NAME -f $BASE_PATH logs -f
}

function up_postgres() {
  echo "Starting PostgreSQL..."
  docker compose -p $PROJECT_NAME -f $PG_PATH up -d
}

function down_postgres() {
  echo "Down PostgreSQL..."
  docker compose -p $PROJECT_NAME -f $PG_PATH down -v
}

function logs_postgres() {
  echo "Showing logs for PostgreSQL..."
  docker compose -p $PROJECT_NAME -f $PG_PATH logs -f
}

function up_redis() {
  echo "Starting Redis..."
  docker compose -p $PROJECT_NAME -f $REDIS_PATH up -d
}

function down_redis() {
  echo "Down Redis..."
  docker compose -p $PROJECT_NAME -f $REDIS_PATH down -v
}

function logs_redis() {
  echo "Showing logs for Redis..."
  docker compose -p $PROJECT_NAME -f $REDIS_PATH logs -f
}

function up_nats() {
  echo "Starting NATS..."
  docker compose -p $PROJECT_NAME -f $NATS_PATH up -d
}

function down_nats() {
  echo "Down NATS..."
  docker compose -p $PROJECT_NAME -f $NATS_PATH down -v
}

function logs_nats() {
  echo "Showing logs for NATS..."
  docker compose -p $PROJECT_NAME -f $NATS_PATH logs -f
}

function up_emqx() {
  echo "Starting EMQX..."
  docker compose -p $PROJECT_NAME -f $EMQX_PATH up -d
}

function down_emqx() {
  echo "Down EMQX..."
  docker compose -p $PROJECT_NAME -f $EMQX_PATH down -v
}

function logs_emqx() {
  echo "Showing logs for EMQX..."
  docker compose -p $PROJECT_NAME -f $EMQX_PATH logs -f
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
  down_postgres
  down_redis
  down_nats
  down_emqx
  down_base
  echo "All infrastructure services are down!"
}

function logs_all() {
  echo "Showing logs for all running containers..."
  docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  echo ""
  docker compose -p $PROJECT_NAME logs -f
}

case "$1" in
  up-base) up_base ;;
  down-base) down_base ;;
  logs-base) logs_base ;;

  up-postgres) up_base; up_postgres ;;
  down-postgres) down_postgres; down_base ;;
  logs-postgres) logs_postgres ;;

  up-redis) up_base; up_redis ;;
  down-redis) down_redis; down_base ;;
  logs-redis) logs_redis ;;

  up-nats) up_base; up_nats ;;
  down-nats) down_nats; down_base ;;
  logs-nats) logs_nats ;;

  up-emqx) up_base; up_emqx ;;
  down-emqx) down_emqx; down_base ;;
  logs-emqx) logs_emqx ;;

  up-all|"") up_all ;;
  down-all) down_all ;;
  logs-all) logs_all ;;

  help|-h|--help) print_help ;;

  *)
    echo "Unknown option: $1"
    print_help
    exit 1
    ;;
esac
