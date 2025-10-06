#!/bin/bash
set -e

# Step 1: setup network and copy middleware override
./deploy/setup-network.bash
cp ./deploy/development/traefik/overrides/dynamic.contributor-middleware.yml ./deploy/development/traefik/dynamic/dynamic.yml

# Default to 'up -d' if no command provided
CMD=${1:-up}
shift || true
if [ "$CMD" = "up" ]; then
  CMD="up -d"
fi

# Step 2: Start base infrastructure (Postgres, NATS, Jaeger, etc.)
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

# Step 3: Wait for infra to be ready
sleep 10

# Step 4: Start contributor + app services (task, leaderboardstat, etc.)
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
  -f ./deploy/argus/development/docker-compose.yaml \
  -f ./deploy/contributor/development/docker-compose.yaml \
  -f ./deploy/task/development/docker-compose.yaml \
  -f ./deploy/leaderboardstat/development/docker-compose.yml \
  $CMD "$@"
