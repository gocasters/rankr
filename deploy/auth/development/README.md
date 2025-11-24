# Auth Service (Development)

## Prerequisites
- Run `bash deploy/script/start_infrastructure.sh up-base` or `make infra-up` once to ensure the shared `rankr-shared-development-network` and external volumes exist.
- Copy `deploy/auth/development/config.yml` to suit your environment before starting the container. The compose file points `CONFIG_PATH` at this file directly.
- Copy `.env.development.example` to `.env.development` and override values as needed. Env vars use the `auth_...` prefix with `__` for nesting (e.g. `auth_POSTGRES_DB__HOST`) to override the YAML without touching the file.

## Running
```bash
PROJECT_ROOT=$(pwd) docker compose -f deploy/auth/development/docker-compose.yml up -d
```
The `PROJECT_ROOT` variable ensures the bind mount always resolves from the repository root when running compose directly. The `Makefile` targets already export this variable automatically.
