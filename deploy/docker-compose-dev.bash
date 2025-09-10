#! /bin/bash

cp ./deploy/development/traefik/overrides/dynamic.contributor-middleware.yml ./deploy/development/traefik/dynamic/dynamic.yml

docker compose \
--env-file ./deploy/.env \
--project-directory . \
-f ./deploy/development/grafana-compose.yml \
-f ./deploy/development/jaeger-compose.yml \
-f ./deploy/development/otel_collector-compose.yml \
-f ./deploy/development/prometheus-compose.yml \
-f ./deploy/development/pgadmin-compose.yml \
-f ./deploy/development/traefik-compose.yml \
-f ./deploy/development/temporal-compose.yml \
-f ./deploy/argus/development/docker-compose.yaml \
-f ./deploy/contributor/development/docker-compose.yaml \
-f ./deploy/leaderboardstat/development/docker-compose.yml \
"$@"