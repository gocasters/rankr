## Flow

- **ProjectService**: creates new project
- **WebhookService**: fetches historical events of the project and publishes relevant events
- **LeaderboardScoringService**: gets its own events, calculates contributors scores, serves scores
- **LeaderboardStatService**: gets events from `LeaderboardScoringService` in a job and serves **PublicLeaderboard**

> **Note**: `<GITHUB_TOKEN>` is a valid GitHub access token
> Example: `ghp_lCrJPkr2BWBLu8AmwuVwdhClPbUHGRgN2BnkHa`

---

## Prerequisites

### 1. Setup environment file

```bash
# Copy template
cp deploy/.env.production.example deploy/.env.production

# Edit passwords (IMPORTANT: change all passwords!)
nano deploy/.env.production
```

Required variables in `.env.production`:
```env
POSTGRES_USER=rankr_admin
POSTGRES_PASSWORD=<STRONG_PASSWORD>
POSTGRES_DB=rankr

PROJECT_USER=project_user
PROJECT_PASS=<STRONG_PASSWORD>

WEBHOOK_USER=webhook_user
WEBHOOK_PASS=<STRONG_PASSWORD>

LEADERBOARDSCORING_USER=leaderboardscoring_user
LEADERBOARDSCORING_PASS=<STRONG_PASSWORD>

LEADERBOARDSTAT_USER=leaderboardstat_user
LEADERBOARDSTAT_PASS=<STRONG_PASSWORD>

# ... other services
```

### 2. Clean Docker (optional, fresh start)

```bash
docker system prune -a --volumes -f
```

---

## Steps

### 1. Start infrastructure

```bash
make infra-up-prod
```

This will:
- Create `rankr-prod-network`
- Create `rankr-postgres-prod-data` volume
- Start PostgreSQL
- Run `init-services-db.sh` to create all databases and users

Verify PostgreSQL is healthy:
```bash
docker ps | grep shared-postgres
```

### 2. Start all services

```bash
make services-up-prod
```

Or start individually:
```bash
make start-project-app-prod
make start-webhook-app-prod
make start-leaderboardscoring-app-prod
make start-leaderboardstat-app-prod
```

Check status:
```bash
make status-prod
```

---

## Usage

### Create project

```bash
curl -X POST http://localhost:8084/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{
    "name": "rankr",
    "slug": "rankr",
    "owner": "gocasters",
    "repo": "rankr",
    "vcsToken": "<GITHUB_TOKEN>",
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

Verify project in database:
```bash
docker exec -it shared-postgres psql -U project_user -d project_db -c "SELECT * FROM projects"
```

### Fetch historical events

```bash
docker exec webhook-app-prod /app/main fetch-historical \
  --owner=gocasters \
  --repo=rankr \
  --token=<GITHUB_TOKEN> \
  --event-types=pr \
  --include-reviews
```

### Check leaderboard data

Redis keys:
```bash
docker exec shared-redis redis-cli KEYS "leaderboard:*:all_time"
```

Output:
```
leaderboard:1028435569:all_time
leaderboard:global:all_time
```

### Get public leaderboard (gRPC)

```bash
grpcurl -plaintext \
  -d '{"project_id": 1028435569, "page_size": 10, "offset": 0}' \
  localhost:8091 leaderboardstat.LeaderboardStatService/GetPublicLeaderboard
```

---

## Quick Commands

| Command | Description |
|---------|-------------|
| `make up-prod` | Start everything (infra + services) |
| `make down-prod` | Stop everything |
| `make status-prod` | Show container status |
| `make infra-up-prod` | Start only infrastructure |
| `make services-up-prod` | Start only services |
| `make start-<service>-app-prod` | Start specific service |
| `make stop-<service>-app-prod` | Stop specific service |

---

## Troubleshooting

### Check logs

```bash
# Infrastructure
make infra-logs-prod

# Specific service
docker logs -f webhook-app-prod
docker logs -f project-app-prod
```

### Database connection issues

```bash
# Check PostgreSQL
docker exec -it shared-postgres psql -U rankr_admin -d rankr -c "\l"

# Check user exists
docker exec -it shared-postgres psql -U rankr_admin -d rankr -c "\du"
```

### Restart a service

```bash
make stop-webhook-app-prod
make start-webhook-app-prod
```