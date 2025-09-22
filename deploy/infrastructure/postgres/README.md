# PostgreSQL Database Setup

## Quick Start

```bash
# Start all services
./deploy/docker-compose-dev.bash up

# Start only leaderboardstat service with database
./deploy/docker-compose-dev-leaderboardstat-local.bash up --build

# Stop and clean up
./deploy/docker-compose-dev.bash down -v
```


### Query the related tables
```
docker exec -it rankr-rankr-shared-postgres-1 psql -U leaderboardstat_user -d rankr_db

SELECT * FROM leaderboardstat_schema.gorp_migrations;
SELECT * FROM leaderboardstat_schema.scores;
```
