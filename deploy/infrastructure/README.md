## Infrastructure Overview

This directory contains all Docker Compose configurations required to run the **shared infrastructure stack** for the
project (e.g., PostgreSQL, Redis, NATS, EMQX, etc).

Each component is isolated in its own folder under `deploy/infrastructure/`, following the pattern:

---

### Base Compose File

The **base** configuration (`docker-compose.base.yml`) defines **shared networks and volumes** used by all
infrastructure components.

- **Network:**  
  `rankr-shared-development-network` — allows all infra containers to communicate.
- **Volumes:**  
  Shared persistent storage for Postgres, Redis, NATS, and EMQX.

Before running any infrastructure service, the base stack must be initialized to ensure all shared networks and volumes
exist.

Example:

```bash
docker compose -f deploy/infrastructure/base/development/docker-compose.base.yml up -d
````

---

### Individual Infrastructure Services

Each service (PostgreSQL, Redis, NATS, EMQX) has its own Docker Compose file inside its respective folder.
This modular approach allows developers to start only the required dependencies for their microservices during local
development.

---

### Adding a New Infrastructure Dependency

If you need to add a **new infrastructure component** (e.g., Kafka, MongoDB, MinIO, etc.), follow these steps:

1. **Update the Base Compose File**
   Edit the `docker-compose.base.yml` file under `deploy/infrastructure/base/development/` to include any **new volumes
   **
   or **networks** required by the new service.

2. **Create a New Folder**
   Inside `deploy/infrastructure/`, create a new directory for your component following the existing structure, for
   example:

   ```
   deploy/infrastructure/kafka/development/docker-compose.kafka.yml
   ```

3. **Add the Compose File**
   In that folder, create a new `docker-compose.<service>.yml` file containing the configuration for your component.

4. **Update the Infrastructure Script**
   Open `deploy/infrastructure/script/start_infrastructure.sh` and **add a new case block** for your service,
   following the existing pattern used for other components (e.g., PostgreSQL, Redis, NATS).
   This ensures your new infrastructure component can be started individually or as part of the full stack.

   Example snippet:

   ```bash
   "up-kafka")
       echo "Starting Kafka..."
       docker compose -f "$BASE_DIR/kafka/development/docker-compose.kafka.yml" up -d
       ;;
   ```

5. **Test It**
   Start the base infrastructure first, then bring up your new service individually to ensure it works properly:

   ```bash
   docker compose -f deploy/infrastructure/base/development/docker-compose.base.yml up -d
   docker compose -f deploy/infrastructure/kafka/development/docker-compose.kafka.yml up -d
   ```

This process ensures consistent structure, isolated service configuration, and proper integration with the shared
network and helper scripts.

---

### Usage Instructions

See [USAGE.md](USAGE.md) for detailed usage examples and a helper script to start all or individual infrastructure
components.
