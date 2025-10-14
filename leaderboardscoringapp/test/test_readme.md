# Testing Guide: Leaderboard Scoring Service

This guide covers two testing approaches: automated integration tests and manual end-to-end testing.

---

## Part 1: Automated Integration Tests

Automated tests that verify the complete event processing pipeline with real dependencies (PostgreSQL, Redis).

### Running Integration Tests

```bash
# Run all integration tests
go test -v -timeout 30m -run TestIntegrationSuite

# Run specific test
go test -v -run TestIntegrationSuite/TestProcessScoreEvent_Success

# Run with short flag (skips integration tests)
go test -short ../...
```

### What Integration Tests Verify

- Event processing with real Redis leaderboard updates
- Multiple events increasing cumulative scores
- Leaderboard ranking order and sorting
- Event persistence to PostgreSQL
- Leaderboard queries through service layer
- Concurrent event processing from multiple users
- Pagination with real data
- Error handling for invalid inputs

### Test Output

```
✅ TestProcessScoreEvent_Success - Verifies single event processing
✅ TestMultipleEvents_IncreaseScore - Validates cumulative scoring
✅ TestLeaderboardRanking_Real - Confirms ranking order
✅ TestEventPersistence_Real - Checks database storage
✅ TestGetLeaderboard_Real - Tests service queries
✅ TestConcurrentEvents - Validates thread safety
✅ TestPagination_Real - Tests pagination logic
✅ TestGetLeaderboard_InvalidOffset - Error handling

Tests complete in ~40-60 seconds with real PostgreSQL and Redis containers
```

### Key Features

- **Real Dependencies**: Uses testcontainers to start actual PostgreSQL and Redis
- **Automatic Setup/Teardown**: Containers are managed automatically
- **No External Tools Required**: Runs entirely within Go test framework
- **CI/CD Ready**: Can be integrated into pipelines

---

## Part 2: Manual End-to-End Testing

Manual testing for exploration, debugging, and performance validation using real infrastructure.

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Optional: redis-cli, psql

To run manual End-to-End testing, navigate to root directory project:

### Start Infrastructure

```bash
make lb-up-deps
```

Expected healthy services:

- `lb-db`
- `lb-redis`
- `lb-nats`

### Run Leaderboard Service

In a separate terminal:

```bash
go run ./cmd/leaderboardscoring/main.go serve --migrate-up
```

### Generate Test Events

Generate and publish events to NATS:

```bash
go run ./leaderboardscoringapp/test/generateevent/main.go -count 100
```

Options:

- `-count N`: Number of events (default: 10)
- `-nats URL`: NATS URL (default: nats://localhost:4222)
- `-topic NAME`: Topic name (default: raw_events)

Events are randomly generated GitHub-like actions (PRs, issues, commits) and published via Watermill + NATS.

### Verify Results

#### Redis Leaderboards

Open RedisInsight at http://localhost:5540

Inspect leaderboard keys:

- `leaderboard:1:all_time` - Global all-time leaderboard
- `leaderboard:1:yearly:YYYY` - Yearly leaderboard
- `leaderboard:1:monthly:YYYY-MM` - Monthly leaderboard
- `leaderboard:1:weekly:YYYY-Www` - Weekly leaderboard

Query example (in RedisInsight CLI):

```
ZREVRANGE leaderboard:1:all_time 0 -1 WITHSCORES
```

Expected: Sorted set with user IDs and aggregated scores

#### NATS JetStream

Access NATS CLI via nats-box:

```bash
docker exec -it nats_box sh
```

Inside container:

```bash
# View streams
nats stream ls

# Check processed events stream
nats stream info PROCESSED_EVENTS

# View recent messages
nats stream view PROCESSED_EVENTS --last 10

# Check batch processor consumer
nats consumer info PROCESSED_EVENTS batch-processor
```

#### PostgreSQL Database

Connect to database:

```bash
docker exec -it leaderboardscoring-db psql -U leaderboardscoring_admin -d leaderboardscoring_db
```

Verify event storage:

```sql
-- Count total events
SELECT COUNT(*)
FROM processed_score_events;

-- View recent events
SELECT user_id, event_type, score_delta, processed_at
FROM processed_score_events
ORDER BY processed_at DESC LIMIT 10;

-- Aggregate scores per user
SELECT user_id, SUM(score_delta) as total_score
FROM processed_score_events
GROUP BY user_id
ORDER BY total_score DESC LIMIT 10;
```

### Validation Checklist

After generating 100 events, verify:

- [ ] Redis leaderboard keys exist with user data
- [ ] PostgreSQL contains processed events
- [ ] NATS processed events stream updated
- [ ] Service logs show no errors
- [ ] All 100 events were processed

### Performance Metrics

With 100 events:

| Stage                | Duration       |
|----------------------|----------------|
| Event generation     | ~30 seconds    |
| Event processing     | ~5 seconds     |
| Redis update         | Immediate      |
| Database persistence | 5-10 seconds   |
| **Total**            | ~40-50 seconds |

### Troubleshooting

**Events not in Redis:**

- Check service logs for errors
- Verify NATS connection: `docker logs leaderboardscoring-nats`
- Confirm events were published: `nats stream view RAW_EVENTS --last 10`

**PostgreSQL connection errors:**

- Verify service: `docker exec leaderboardscoring-db pg_isready`
- Check logs: `docker logs leaderboardscoring-db`

**NATS issues:**

- Reset data: `docker-compose down -v && docker-compose up -d`
- Verify stream exists: `nats stream ls`

### Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes
docker-compose down -v
```

---

## When to Use Each Approach

### Use Integration Tests When:

- Running CI/CD pipelines
- Validating pull requests
- Regression testing
- Automated coverage requirements

### Use Manual E2E Testing When:

- Debugging specific issues
- Performance testing at scale (10,000+ events)
- UI inspection with RedisInsight
- Exploring new features
- Ad-hoc validation

---

## Architecture Flow

```
Integration Tests:
  Event → Service → Real Redis/PostgreSQL → Assertions ✅

Manual E2E Testing:
  Event Generator → NATS → Service → Redis → RedisInsight/Redis CLI
                                  ↓
                              PostgreSQL → SQL queries
```

---

## Notes

- Integration tests run in ~40-60 seconds with automatic container lifecycle
- Manual testing mirrors production with real services but requires manual verification
- Both approaches complement each other for comprehensive validation
- EventGenerator creates realistic GitHub-like events (PRs, issues, commits)
- All timestamps use UTC for consistency