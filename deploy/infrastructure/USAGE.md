# Using the Infrastructure Stack

This guide explains how to start and manage the shared infrastructure components for local development.

---

## Prerequisites

- Docker and Docker Compose installed
- Access to the `deploy/infrastructure/` directory
- Ensure the helper script is executable:

```bash
chmod +x deploy/script/start_infrastructure.sh
```

---

## Step 1 — Initialize Base

Before starting any infrastructure component, initialize the base network and volumes:

```bash
docker compose -f deploy/infrastructure/base/development/docker-compose.base.yml up -d
```

---

## Step 2 — Start Individual Infrastructure Components

You can bring up any infrastructure service independently:

```bash
# Start PostgreSQL
docker compose -f deploy/infrastructure/postgresql/development/docker-compose.postgresql.yml up -d

# Start Redis
docker compose -f deploy/infrastructure/redis/development/docker-compose.redis.yml up -d

# Start NATS
docker compose -f deploy/infrastructure/nats/development/docker-compose.nats.yml up -d

# Start EMQX
docker compose -f deploy/infrastructure/emqx/development/docker-compose.emqx.yml up -d
```

---

## Step 3 — Start All Infrastructure at Once

For convenience, a helper script is provided to automatically start:

* The base environment (network + volumes)
* All infrastructure services

You can find it here:
👉 [`deploy/script/start_infrastructure.sh`](../script/start_infrastructure.sh)

Run it from the project root:

```bash
bash deploy/script/start_infrastructure.sh
```

This script accepts the following commands:

```bash
# Start only base
bash deploy/script/start_infrastructure.sh up-base

# Start specific service
bash deploy/script/start_infrastructure.sh up-redis

# Start all infrastructure
bash deploy/script/start_infrastructure.sh up-all

# Stop all infrastructure
bash deploy/script/start_infrastructure.sh down-all

# Show help
bash deploy/script/start_infrastructure.sh help
```

---

## Stopping Infrastructure

To stop infrastructure services, use the helper script commands.

```bash
# Stop a specific service
bash deploy/script/start_infrastructure.sh down-redis

# Stop all infrastructure components
bash deploy/script/start_infrastructure.sh down-all
```

---

**Tip:**
If you add a new infrastructure dependency (like Kafka or MinIO), remember to:

1. Update the base compose file with required volumes/networks.
2. Add a new compose file under `deploy/infrastructure/<service>/development/`.
3. Edit `deploy/script/start_infrastructure.sh` and include `up-<service>` and `down-<service>` commands.

