#! /bin/bash

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

cp ./deploy/development/traefik/overrides/dynamic.contributor-middleware.yml ./deploy/development/traefik/dynamic/dynamic.yml

docker compose \
--env-file ./deploy/.env \
--project-directory . \
-f ./deploy/infrastructure/postgresql/development/docker-compose.postgresql.yml \
-f ./deploy/development/nats-compose.yml \
-f ./deploy/contributor/development/docker-compose.yaml \
-f ./deploy/development/otel_collector-compose.yml \
-f ./deploy/development/jaeger-compose.yml \
-f ./deploy/task/development/docker-compose.yaml \
-f ./deploy/auth/development/docker-compose.yml \
"$@"