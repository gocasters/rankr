#!/bin/bash
set -e

# Step 1: Setup network and middleware override
./deploy/setup-network.bash
cp ./deploy/development/traefik/overrides/dynamic.contributor-middleware.yml ./deploy/development/traefik/dynamic/dynamic.yml

# Determine command (default = up -d)
CMD=${1:-up}
shift || true
if [ "$CMD" = "up" ]; then
  CMD="up -d"
fi

# Step 2–3: Bootstrap infra only for full 'up -d'
if [ "$CMD" = "up -d" ]; then
  echo "🔧 Bringing up base infrastructure..."
  docker compose \
    --env-file ./deploy/.env \
    --project-directory . \
    -f ./deploy/infrastructure/postgres/development/docker-compose.yml \
    -f ./deploy/development/grafana-compose.yml \
    -f ./deploy/development/jaeger-compose.yml \
    -f ./deploy/development/centrifugo-compose.yml \
    -f ./deploy/development/emqx-compose.yml \
    -f ./deploy/development/otel_collector-compose.yml \
    -f ./deploy/development/prometheus-compose.yml \
    -f ./deploy/development/pgadmin-compose.yml \
    -f ./deploy/development/nats-compose.yml \
    $CMD "$@"

  echo "⏳ Waiting for infrastructure to stabilize..."
  sleep 10
fi

# Step 4: Start or operate on app services
echo "🚀 Running docker compose for app stack ($CMD)..."
docker compose \
  --env-file ./deploy/.env \
  --project-directory . \
  -f ./deploy/infrastructure/postgres/development/docker-compose.yml \
  -f ./deploy/development/nats-compose.yml \
  -f ./deploy/contributor/development/docker-compose.yaml \
  -f ./deploy/development/otel_collector-compose.yml \
  -f ./deploy/development/jaeger-compose.yml \
  -f ./deploy/task/development/docker-compose.yaml \
  -f ./deploy/leaderboardstat/development/docker-compose.yml \
  $CMD "$@"
