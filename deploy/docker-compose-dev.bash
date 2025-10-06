#!/bin/bash
set -e

# --------------------------------------------------------
# Docker Compose Dev Bootstrap Script
# --------------------------------------------------------
# Starts core infrastructure (Postgres, NATS, etc.)
# and application services (contributor, leaderboardstat, etc.)
# Usage:
#   ./deploy/docker-compose-dev.bash           → up -d (full stack)
#   ./deploy/docker-compose-dev.bash down -v   → shut down everything
#   ./deploy/docker-compose-dev.bash logs task → view logs for task service
# --------------------------------------------------------

# Step 1: Setup network and middleware override
./deploy/setup-network.bash
cp ./deploy/development/traefik/overrides/dynamic.contributor-middleware.yml \
   ./deploy/development/traefik/dynamic/dynamic.yml

# Step 2: Determine the command (default = up -d)
CMD=${1:-up}
shift || true
if [ "$CMD" = "up" ]; then
  CMD="up -d"
fi

# Step 3: Bootstrap infra only for full "up -d" with no service filter
if [ "$CMD" = "up -d" ] && [ $# -eq 0 ]; then
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
    up -d

  echo "⏳ Waiting for infrastructure to stabilize..."
  sleep 10
fi

# Step 4: Start or operate on application services
echo "🚀 Running docker compose for app stack ($CMD $*)..."
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

echo "✅ Done."
