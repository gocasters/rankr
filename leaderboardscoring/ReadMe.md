# Leaderboard Scoring Service

## 1. Overview

The **Leaderboard Scoring Service** is a high-performance, event-driven microservice responsible for calculating and
maintaining real-time user leaderboards. It consumes contribution events (e.g., from GitHub webhooks) from a message
broker, processes them to update user scores, and exposes an API to query leaderboard data.

The architecture is designed from the ground up to be **resilient**, **fault-tolerant**, and **data-correct**, ensuring
that scores are always accurate, even in a distributed environment with potential failures.

## 2. Core Features

* **Real-Time Scoring**: Updates user scores and leaderboard rankings in near real-time as contribution events are
  received.
* **Event-Driven Architecture**: Built on a pub/sub model using [Watermill](https://watermill.io/) for broker-agnostic
  message consumption.
* **Guaranteed Data Integrity**: Implements a robust idempotency strategy to prevent duplicate event processing.
* **Dual Storage Strategy**:
    * **Redis (Hot Storage)**: Used for high-speed, in-memory leaderboards via Sorted Sets.
    * **PostgreSQL (Cold Storage)**: Used for persisting the historical log of all contribution events.
* **Disaster Recovery**: Includes a snapshot and restore mechanism to ensure fast recovery in case of a Redis failure.
* **Clean Architecture**: Follows a clean, layered architecture (`delivery` -> `service` -> `repository`) for high
  maintainability and testability.
* **HTTP API**: Exposes a RESTful API for querying leaderboard data.

## 3. Architecture Deep Dive

### 3.1. Data Flow

The service operates on a simple and effective data flow:

1. **Event Production**: An external service (e.g., a `webhook-service`) translates raw webhook payloads (from GitHub,
   etc.) into a standardized Protobuf format.
2. **Publishing**: The event is published to a message broker (e.g., NATS) on a specific topic (
   `CONTRIBUTION_REGISTERED`).
3. **Consumption**: The `leaderboardscoring` service, using its Watermill consumer, subscribes to this topic and
   receives the event.
4. **Processing**: The consumer handler validates, processes, and updates the relevant data stores.

### 3.2. Resilience and Correctness Strategy

To guarantee that every event is processed **exactly once** from a business logic perspective, we combine two patterns:

* **At-Least-Once Delivery**: The Watermill consumer uses an `ACK/NACK` protocol with the broker. A message is only
  removed from the queue after it is successfully processed (`ACK`). If any failure occurs, the message is `NACK`ed and
  will be re-delivered, ensuring no events are lost.
* **Idempotent Consumer**: To handle the potential for duplicate messages from the at-least-once delivery, the consumer
  handler implements a robust idempotency check using Redis:
    1. **Processed Check**: Before processing, it checks if the event's unique ID exists in a `processed_events` list.
       If so, the event is a duplicate and is safely ignored.
    2. **Temporary Lock**: It acquires a short-lived distributed lock on the event ID. This prevents two instances of
       the service from processing the same event concurrently.
    3. **Mark as Processed**: Only after the business logic is successfully completed is the event ID added to the
       `processed_events` list.

### 3.3. Storage Strategy

* **Redis**: All leaderboards are stored in Redis `Sorted Sets` for extremely fast reads and writes. The keys are
  structured to support various timeframes and scopes:
    * `leaderboard:global:{timeframe}:{period}`
    * `leaderboard:project:{project_id}:{timeframe}:{period}`
* **PostgreSQL**: Every processed `ContributionEvent` is saved to a PostgreSQL table. This provides a permanent,
  auditable log of all activities that contributed to user scores.

### 3.4. Disaster Recovery: Snapshot & Restore

To recover quickly from a potential Redis failure without reprocessing millions of historical events, the service
includes two key methods:

* `CreateLeaderboardSnapshot`: A function intended to be run by a scheduler (e.g., every hour). It reads the `all-time`
  leaderboards from Redis and saves their current state as a snapshot in PostgreSQL.
* `RestoreLeaderboardFromSnapshot`: A function intended to be run on service startup. If it detects that Redis is empty,
  it will repopulate the `all-time` leaderboards from the latest snapshot in PostgreSQL.

## 4. API Endpoints

The service exposes an HTTP API for querying leaderboard data.

| Method | Endpoint                 | Description                            |
|:-------|:-------------------------|:---------------------------------------|
| `GET`  | `/v1/health-check`       | Checks the health of the service.      |
| `GET`  | `/v1/leaderboard/public` | Fetches a specific leaderboard. (TODO) |

## 5. Configuration

The service is configured via a YAML file or environment variables. Key settings include:

* `http_server`: Port and timeout settings for the API.
* `postgres_db`: Connection details for the PostgreSQL database.
* `redis`: Connection details for the Redis instance.
* `consumer`: Settings for the idempotency checker (e.g., TTL for processed keys and locks).