# LeaderboardScoring gRPC Client (Example)

This example demonstrates how to connect to the **Leaderboard Scoring Service** over gRPC, fetch leaderboard data, and
display it in a CLI-friendly format.

---

## Usage

```bash
go run ./cmd/leaderboardscoring/test/rpc.go [flags]
```

### Available Flags

| Flag          | Description                                                       | Default          |
|---------------|-------------------------------------------------------------------|------------------|
| `--addr`      | gRPC server address (`host:port`)                                 | `localhost:8090` |
| `--timeframe` | Leaderboard timeframe (`all_time`, `yearly`, `monthly`, `weekly`) | `all_time`       |
| `--project`   | Project ID (empty for global leaderboard)                         | `1`              |
| `--limit`     | Number of records to fetch                                        | `10`             |
| `--offset`    | Offset for pagination                                             | `0`              |
| `--timeout`   | Request timeout duration                                          | `5s`             |

---

## Example Commands

### Fetch top 10 monthly scores for project 1

```bash
go run ./example/leaderboardscoring_getleaderboard_grpc_client/main.go \
  --addr localhost:8090 \
  --timeframe monthly \
  --project 1 \
  --limit 10
```

### Fetch the next page (offset 20)

```bash
go run ./example/leaderboardscoring_getleaderboard_grpc_client/main.go \
  --addr localhost:8090 \
  --timeframe monthly \
  --project 1 \
  --limit 10 \
  --offset 20
```

---

## Overview

This CLI uses:

* **`pkg/grpc`** for client connection setup with retry and backoff.
* **`adapter/leaderboardscoring`** as a generated gRPC client adapter.
* **`leaderboardscoringapp/service`** protobuf definitions for request/response models.

It supports configurable timeframes, pagination, and project-level leaderboards.

---

## Note

This client uses **insecure credentials** for local development.
For production, enable TLS in `pkg/grpc/client.go`.
