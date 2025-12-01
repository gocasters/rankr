# Project Service (Development)

## Prerequisites
- Run `bash deploy/script/start_infrastructure.sh up-base` or `make infra-up` so the shared services and `rankr-shared-development-network` exist.
- Copy `.env.development.example` to `.env.development` in this directory and adjust the database, Redis, or NATS endpoints if your setup differs.
- Review `config.yml` if you need to tweak ports or internal settings before starting the container.

## Running
```bash
docker compose -f deploy/project/development/docker-compose.yaml up -d
```
The Compose file mounts `config.yml` and automatically loads `.env.development`, so no additional flags are required once those files are in place.
