#!/usr/bin/env bash
set -e

ENV_FILE_OPTION=""
if [ -f "./deploy/.env" ]; then
  set -a
  source ./deploy/.env
  set +a
  ENV_FILE_OPTION="--env-file ./deploy/.env"
fi

INFRASTRUCTURE_COMPOSE="./deploy/infrastructure/postgres/docker-compose.yml"
LEADERBOARD_STAT_DIR="./deploy/leaderboardstat/development"
COMPOSE_FILE="$LEADERBOARD_STAT_DIR/docker-compose.yml"

if [ ! -f "$INFRASTRUCTURE_COMPOSE" ]; then
  echo "Error: infrastructure docker-compose.yml not found"
  exit 1
fi

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "Error: docker-compose.yml not found in $LEADERBOARD_STAT_DIR"
  exit 1
fi

docker compose \
  $ENV_FILE_OPTION \
  --project-directory . \
  -f "$INFRASTRUCTURE_COMPOSE" \
  -f "$COMPOSE_FILE" \
  "$@"