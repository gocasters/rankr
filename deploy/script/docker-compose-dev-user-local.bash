#!/bin/bash

set -a
source ./deploy/.env
set +a

cp ./deploy/development/traefik/overrides/dynamic.contributor-local.yml ./deploy/development/traefik/dynamic/dynamic.yml

docker compose --env-file ./deploy/.env \
  --project-directory . \
  --profile contributor-local \
  -f ./deploy/development/grafana-compose.yml \
  -f ./deploy/development/jaeger-compose.yml \
  -f ./deploy/development/otel_collector-compose.yml \
  -f ./deploy/development/prometheus-compose.yml \
  -f ./deploy/development/pgadmin-compose.yml \
  -f ./deploy/development/traefik-compose.yml \
  -f ./deploy/development/temporal-compose.yml \
  -f ./deploy/development/nats-compose.yml \
  -f ./deploy/contributor/development/docker-compose.no-service.yaml \
  "$@"
