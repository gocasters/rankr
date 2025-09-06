# Leaderboard Scoring Service

## 1. Overview

The **Leaderboard Scoring Service** is a resilient, event-driven microservice that calculates and maintains real-time
user leaderboards. It consumes contribution events, updates user scores, and exposes an API to query the rankings. The
architecture is designed for high performance, data integrity, and fault tolerance.

## 2. Core Architecture

The service is built on a clean, layered architecture (`delivery` -> `service` -> `repository`) and relies on several
key patterns:

* **Event-Driven Consumption**: The service uses [Watermill](https://watermill.io/) to subscribe to a message broker in
  a broker-agnostic way. It processes events from a `CONTRIBUTION_REGISTERED` topic.

* **Resilience and Data Integrity**: To ensure every event is processed exactly once from a business logic perspective,
  we combine two strategies:

    1. **At-Least-Once Delivery**: The consumer uses an `ACK/NACK` protocol with the broker, guaranteeing that no events
       are lost during transient failures.

    2. **Idempotent Consumer**: A robust idempotency check using a temporary lock and a processed-event list in Redis
       prevents duplicate messages from being processed more than once.

* **Dual Storage Strategy**:

    * **Redis (Hot Storage)**: All leaderboards are stored in Redis `Sorted Sets` for extremely fast, real-time read and
      write operations.

    * **PostgreSQL (Cold Storage)**: A permanent, auditable log of every processed contribution event is persisted in a
      PostgreSQL database for analytical purposes.

* **Asynchronous Event Persistence**: To maximize performance, the persistence of historical events to PostgreSQL is
  decoupled from the critical path of real-time scoring. This is achieved using an **Asynchronous Batch Processing
  Pattern**:

    1. **Hot Path (Real-time)**: The consumer immediately updates the scores in Redis and acknowledges the message.

    2. **Cold Path (Background)**: The raw event is then pushed into a persistent, temporary queue (e.g., Kafka,
       RabbitMQ, Nats).

    3. **Batch Worker**: A separate background worker runs periodically, dequeues events from this queue in batches, and
       performs an efficient bulk insert into PostgreSQL. This ensures that any latency from the database does not
       impact the real-time scoring performance.

* **Disaster Recovery**: The service includes a snapshot mechanism to periodically save the state of the Redis
  leaderboards to PostgreSQL. A restore function can quickly rebuild the leaderboards from the latest snapshot after a
  failure, avoiding the need to reprocess the entire event history.

## 3. Usage

### Run leaderboard-scoring app

```bash
  # run compose file
 docker compose -f deploy/leaderboardscoring/development/docker-compose.no-service.yml up -d
  # run leaderboard-scoring app
 go run ./cmd/leaderboardscoring/main.go serve
 
  # for show containers log
 docker compose -f deploy/leaderboardscoring/development/docker-compose.no-service.yml logs -f
```

### Stopping service

```bash
  # stopping leaderboard-scoring service
 ctrl + c
 
  # for down containers
 docker compose -f deploy/leaderboardscoring/development/docker-compose.no-service.yml down -v
 
```

## 3. API Endpoints

The service exposes a basic HTTP API for health checks and querying leaderboard data.

| Method | Endpoint                 | Description                            |
|:-------|:-------------------------|:---------------------------------------|
| `GET`  | `/v1/health-check`       | Checks the health of the service.      |
| `GET`  | `/v1/leaderboard/public` | Fetches a specific leaderboard. (TODO) |

```