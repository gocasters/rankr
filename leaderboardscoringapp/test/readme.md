## Leaderboard Scoring E2E Testing

This guide explains how to perform end-to-end (E2E) tests for the leaderboard scoring application using **NATS JetStream
**, **Redis**, and **PostgreSQL**.

### Overview

The E2E test simulates real GitHub-like events, publishes them to NATS JetStream, and verifies that the leaderboard
service correctly processes and stores them in Redis and PostgreSQL.

---

### Prerequisites

* Docker & Docker Compose
* Go 1.21+
* NATS, Redis, PostgreSQL services (via Docker Compose)

---

### Services Setup

Use the provided `docker-compose.yml` to start required services:

```bash
docker compose up -d
```

Services include:

* **PostgreSQL** (`leaderboardscoring-db`)
* **Redis** (`leaderboardscoring-redis`)
* **NATS JetStream** (`leaderboardscoring-nats`)
* **RedisInsight** (optional, for inspecting Redis)
* **NATS Box** (optional, for CLI testing)

---

### Start Leaderboard Service

Run the leaderboard service with database migration:

```bash
go run ./cmd/leaderboardscoring/main.go serve --migrate-up
```

---

### Generate Test Events

Use the `generateevents` script to create and publish raw events to NATS:

```bash
go run ./leaderboardscoringapp/test/cmd/generateevents/main.go -count 100 -topic raw_events
```

* Events are randomly generated GitHub-like actions (PRs, issues, pushes, reviews).
* Events are serialized using **Protobuf** and published via **Watermill + NATS adapter**.
* Each event is published with a small delay (300ms) to simulate real-world timing.

---

### Verify Results

#### 1. Redis

* Open **RedisInsight** at [http://localhost:5540/](http://localhost:5540/)
* Check that leaderboard keys such as:

  ```
  leaderboard:global:all_time
  leaderboard:global:monthly:2025-10
  leaderboard:<project_id>:weekly:2025-W42
  ```

  contain member scores.

#### 2. NATS JetStream

Access the NATS CLI via the `nats-box` container:

```bash
docker exec -it nats_box /bin/sh
nats stream ls
nats consumer info <stream> <consumer>
```

Confirm that events are received and acknowledged by the leaderboard service.

#### 3. PostgreSQL

To verify **processed events** persisted in the database:

```bash
psql -h localhost -p 5432 -U leaderboardscoring_admin -d leaderboardscoring_db
```

Enter the password when prompted:

```
password123
```

Then run the following SQL query:

```sql
SELECT *
FROM processed_score_events;
```

**Sample Output:**

```
 id  | user_id |     event_type      |      event_timestamp       | score_delta |        processed_at
-----+---------+---------------------+----------------------------+-------------+----------------------------
   1 | 100     | pull_request_opened | 2025-10-13 09:59:47.236244 |           1 | 2025-10-13 09:59:47.981153
   2 | 101     | pull_request_closed | 2025-10-13 09:59:47.534909 |           2 | 2025-10-13 09:59:47.981153
```

---

### Notes

* This test validates the **entire event flow**:

    1. Event generation
    2. Event publishing via NATS JetStream
    3. Processing by leaderboard service
    4. Storage in Redis and PostgreSQL
* Recommended for CI/CD pipelines to verify full system integrity.
* The setup mirrors a production-like environment for accurate behavior validation.
