# Auth Service (Development)

## Prerequisites
- Run `bash deploy/script/start_infrastructure.sh up-base` or `make infra-up` once to ensure the shared `rankr-shared-development-network` and external volumes exist.
- Copy `deploy/auth/development/config.yml` to suit your environment before starting the container (the compose file binds it with `create_host_path: false`).
- Copy `.env.development.example` to `.env.development` and override the database or messaging hosts if your infrastructure differs. The compose file automatically loads this env file.

## Running
```bash
PROJECT_ROOT=$(pwd) docker compose -f deploy/auth/development/docker-compose.yml up -d
```
The `PROJECT_ROOT` variable ensures the bind mount always resolves from the repository root when running compose directly. The `Makefile` targets already export this variable automatically.
