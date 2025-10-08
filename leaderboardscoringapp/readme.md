# Leaderboard Scoring Service

### Table of Contents

1. [Overview](#1-overview)
2. [Core Architecture](#2-core-architecture)
3. [Usage](#3-usage)
    - [Run leaderboard-scoring app](#run-leaderboard-scoring-app)
    - [Stopping service](#stopping-service)
4. [API Endpoints](#4-api-endpoints)
5. [gRPC API](#5-grpc-api)
    - [Service Discovery](#service-discovery)
    - [Calling the GetLeaderboard Method](#calling-the-getleaderboard-method)

---

## 1. Overview

The **Leaderboard Scoring Service** is a resilient, event-driven microservice that calculates and maintains real-time user leaderboards. It consumes contribution events, updates user scores, and exposes an API to query the rankings. The architecture is designed for high performance, data integrity, and fault tolerance.

## 2. Core Architecture

The service is built on a clean, layered architecture (`delivery` -> `service` -> `repository`) and relies on several key patterns:

* **Event-Driven Consumption**: The service uses [Watermill](https://watermill.io/) to subscribe to a message broker in a broker-agnostic way. It processes events from a `CONTRIBUTION_REGISTERED` topic.

* **Resilience and Data Integrity**: To ensure every event is processed exactly once from a business logic perspective, we combine two strategies:

    1. **At-Least-Once Delivery**: The consumer uses an `ACK/NACK` protocol with the broker, guaranteeing that no events are lost during transient failures.

    2. **Idempotent Consumer**: A robust idempotency check using a temporary lock and a processed-event list in Redis prevents duplicate messages from being processed more than once.

* **Disaster Recovery**: The service includes a snapshot mechanism to periodically save the state of the Redis leaderboards to PostgreSQL. A restore function can quickly rebuild the leaderboards from the latest snapshot after a failure, avoiding the need to reprocess the entire event history.

---

### Event-Driven Models: Push-Based vs Pull-Based (JetStream Native Batching)

Recently, while implementing **batch processing** in the `leaderboard-scoring` service, we explored two event-driven communication patterns:

#### 🔹 Push-Based Model
Used for *real-time leaderboard updates* via **Watermill**:
- JetStream automatically pushes messages to the consumer.
- Each event is processed as soon as it arrives.
- ✅ Low latency, ideal for instant score updates.
- ⚠️ Higher DB load since every event triggers a separate write.

```
Webhook → JetStream → Watermill Consumer → Process → Update Redis
```

#### 🔹 Pull-Based Model (JetStream Native)
Used for *batch persistence* of processed events:
- The consumer **actively requests** messages (e.g., 500 at a time).
- Messages are processed and stored in PostgreSQL as a single batch.
- ✅ 10x faster persistence and 99% less DB load.
- ✅ Fully fault-tolerant (JetStream handles retries).
- ✅ No in-memory buffering, reliable by design.

```
Processed Events → JetStream ← Pull Consumer (Batch 500) → PostgreSQL
```

#### Why JetStream Native Batching?
- **Performance**: Batch inserts reduce I/O and transaction overhead.
- **Efficiency**: Lower connection usage and better throughput.
- **Reliability**: Automatic redelivery and checkpointing by JetStream.
- **Control**: Consumer defines batch size, wait interval, and backpressure.
- **Cost Reduction**: Fewer DB writes → lower operational cost.

In summary,
- **Push-based** → for **real-time scoring (low latency)**
- **Pull-based (native batching)** → for **efficient event persistence (high throughput)**

Together, they achieve both **speed and scalability** in production.

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

## 4. API Endpoints

The service exposes a basic HTTP API for health checks.

| Method | Endpoint | Description |
|:---|:---|:---|
| `GET` | `/v1/health-check` | Checks the health of the service. |

## 5. gRPC API

The primary way to query leaderboard data is through the gRPC API. You can interact with this API using a tool like [`grpcurl`](https://github.com/fullstorydev/grpcurl).

### Service Discovery

First, ensure the service is running. You can then list all available services and describe the leaderboard service to see its methods. (Assuming the gRPC server is on port `8090`).

**List all services:**

```bash
grpcurl -plaintext localhost:8090 list
```

**Describe the leaderboard service:**

```bash
grpcurl -plaintext localhost:8090 describe leaderboardscoring.v1.LeaderboardScoringService
```

### Calling the GetLeaderboard Method

You can call the `GetLeaderboard` RPC to fetch data.

**Example 1: Get the global monthly leaderboard (first 10 users)**

* **For Linux/macOS:**

  ```bash
  grpcurl -plaintext -d '{ "timeframe": "TIMEFRAME_MONTHLY", "page_size": 10, "offset": 0 }' localhost:8090 leaderboardscoring.v1.LeaderboardScoringService.GetLeaderboard
  ```

* **For Windows (cmd.exe):**

  ```bash
  grpcurl -plaintext -d "{ \"timeframe\": \"TIMEFRAME_MONTHLY\", \"page_size\": 10, \"offset\": 0 }" localhost:8090 leaderboardscoring.v1.LeaderboardScoringService.GetLeaderboard
  ```

**Example 2: Get the weekly leaderboard for a specific project**

* **For Linux/macOS:**

  ```bash
  grpcurl -plaintext -d '{ "timeframe": "TIMEFRAME_WEEKLY", "project_id": "gocasters/rankr", "page_size": 10, "offset": 0 }' localhost:8090 leaderboardscoring.v1.LeaderboardScoringService.GetLeaderboard
  ```

* **For Windows (cmd.exe):**

  ```bash
  grpcurl -plaintext -d "{ \"timeframe\": \"TIMEFRAME_WEEKLY\", \"project_id\": \"gocasters/rankr\", \"page_size\": 10, \"offset\": 0 }" localhost:8090 leaderboardscoring.v1.LeaderboardScoringService.GetLeaderboard
  ```
