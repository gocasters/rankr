#!/usr/bin/env bash
set -e

ENV_FILE_OPTION=""
if [ -f "./deploy/.env" ]; then
  set -a
  source ./deploy/.env
  set +a
  ENV_FILE_OPTION="--env-file ./deploy/.env"
fi

WEBHOOK_DIR="./deploy/webhook/development"
COMPOSE_FILE="$WEBHOOK_DIR/docker-compose.yml"

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "Error: docker-compose.yml not found in $WEBHOOK_DIR"
  exit 1
fi

docker compose \
  $ENV_FILE_OPTION \
  --project-directory . \
  -f "$COMPOSE_FILE" \
  "$@"
