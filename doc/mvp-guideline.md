## Flow

- **ProjectService**: creates new project
- **WebhookService**: fetches historical events of the project and publishes relevant events
- **LeaderboardScoringService**: gets its own events, calculates contributors scores, serves scores
- **LeaderboardStatService**: gets events from `LeaderboardScoringService` in a job and serves **PublicLeaderboard**

> **Note**: `<GITHUB_TOKEN>` is a valid GitHub access token  
> Example: `ghp_lCrJPkr2BWBLu8AmwuVwdhClPbUHGRgN2BnkHa`
> Usage: "vcsToken":"ghp_lCrJPkr2BWBLu8AmwuVwdhClPbUHGRgN2BnkHa"

---

## Steps

### 1. Start infrastructure

```bash
make infra-up
```

### 2. Project service (dev)

```bash
make start-project-app-dev
```

Create project:

```bash
curl -X POST http://localhost:8084/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "rankr",
    "slug": "rankr",
    "owner": "gocasters",
    "repo": "rankr",
    "vcsToken": <GITHUB_TOKEN>,
    "repoProvider": "GITHUB",
    "status": "active"
  }'

```
Example response:

```json
{
  "id": "df0afcaa-6a03-4e93-bc6e-b30def68d155",
  "name": "rankr",
  "slug": "rankr",
  "gitRepoId": "1028435569",
  "repoProvider": "GITHUB",
  "status": "ACTIVE",
  "createdAt": "2025-12-25T15:12:24.204899Z",
  "updatedAt": "2025-12-25T15:12:24.204899Z"
}
```

Check project added:
```bash
docker exec -it rankr-shared-postgres psql -U project_user -d project_db -c "SELECT * FROM projects"
```

### 3. Webhook service (dev)

```bash
make start-webhook-app-dev
```

Fetch historical events:
```bash
go run cmd/webhook/main.go fetch-historical \
  --owner=gocasters \
  --repo=rankr \
  --token=<GITHUB_TOKEN> \
  --event-types=pr \
  --include-reviews
```

Possible output:
```log
PostgreSQL connection established successfully (pgx v5)
{"time":"2025-12-26T04:33:26.359028-08:00","level":"INFO","msg":"NATS publisher created successfully","url":"nats://localhost:4223"}
{"time":"2025-12-26T04:33:26.359268-08:00","level":"INFO","msg":"Fetching repository info from GitHub","owner":"gocasters","repo":"rankr"}
{"time":"2025-12-26T04:33:27.160718-08:00","level":"INFO","msg":"Repository found on GitHub","repo_id":1028435569,"full_name":"gocasters/rankr"}
{"time":"2025-12-26T04:33:27.162373-08:00","level":"WARN","msg":"gRPC client is using insecure credentials. This is not suitable for production."}
{"time":"2025-12-26T04:33:27.184806-08:00","level":"INFO","msg":"Project found in database","project_id":"ac1f1bb9-f56b-4cfe-bdae-5c778d7640b3","slug":"rankr","git_repo_id":"1028435569"}
{"time":"2025-12-26T04:33:27.185538-08:00","level":"INFO","msg":"Starting historical fetch","owner":"gocasters","repo":"rankr","event_types":["pr"]}
{"time":"2025-12-26T04:33:27.185615-08:00","level":"INFO","msg":"Fetching pull requests from GitHub API"}
{"time":"2025-12-26T04:33:27.18564-08:00","level":"INFO","msg":"Fetching PR page","page":1}
{"time":"2025-12-26T04:33:28.628081-08:00","level":"INFO","msg":"Fetched PRs","count":100,"page":1}
{"time":"2025-12-26T04:33:29.182367-08:00","level":"DEBUG","msg":"Bulk saved historical events","inserted":4,"duplicates":2}
{"time":"2025-12-26T04:33:29.185694-08:00","level":"DEBUG","msg":"Published events to NATS","count":4}
...
{"time":"2025-12-26T04:34:43.662263-08:00","level":"DEBUG","msg":"Bulk saved historical events","inserted":2,"duplicates":4}
{"time":"2025-12-26T04:34:43.663486-08:00","level":"DEBUG","msg":"Published events to NATS","count":2}
{"time":"2025-12-26T04:34:43.663537-08:00","level":"INFO","msg":"Finished fetching PRs","total":119}
{"time":"2025-12-26T04:34:43.663554-08:00","level":"INFO","msg":"Historical fetch completed","success":119,"failed":0,"total":119,"duration":76000000000,"avg_rate":1.556096694120547}
{"time":"2025-12-26T04:34:43.663581-08:00","level":"INFO","msg":"Fetch historical completed successfully"}
```

### 4. LeaderboardScoring service (dev)

```bash
make start-leaderboardscoring-app-dev
```

Check added redis keys:
```bash
docker exec rankr-shared-redis redis-cli KEYS "leaderboard:*:all_time"
```
Output:
```log
leaderboard:1028435569:all_time
leaderboard:global:all_time
```

### 5. LeaderboardStat service (dev)
```bash
make start-leaderboardstat-app-dev
```

Get public leaderboard:

```bash
grpcurl -plaintext \
  -d '{"project_id": 1028435569, "page_size": 10, "offset": 0}' \
  localhost:8098 leaderboardstat.LeaderboardStatService/GetPublicLeaderboard
```

```bash
grpcurl -plaintext \
  -d '{"project_id": 1028435569, "page_size": 10, "offset": 1}' \
  localhost:8098 leaderboardstat.LeaderboardStatService/GetPublicLeaderboard
```

Example output:
```json
{
  "projectId": "1028435569",
  "rows": [
    {
      "userId": "34894549",
      "rank": "1",
      "score": 9
    },
    {
      "userId": "25051128",
      "rank": "2",
      "score": 9
    },
    {
      "userId": "93041804",
      "rank": "3",
      "score": 6
    }
  ]
}
```
