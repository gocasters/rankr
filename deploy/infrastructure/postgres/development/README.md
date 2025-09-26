# PostgreSQL Database Setup

## Quick Start
```bash
# Start all services with shared PostgreSQL
./deploy/docker-compose-dev.bash up

# Stop and clean up
./deploy/docker-compose-dev.bash down -v
```


```bash
# Start only leaderboardstat service with database
./deploy/docker-compose-dev-leaderboardstat-local.bash up --build
```



### Query Databases
```bash
#  Connect as admin
docker exec -it rankr-rankr-shared-postgres-1 psql -U rankr_admin -d postgres

#see all databases
 \l

# See all users/roles
 \du
```

```bash
 # Connect to leaderboardstat database
 docker exec -it rankr-rankr-shared-postgres-1 psql -U leaderboardstat_user -d leaderboardstat_db
```

```
# Inside psql, check tables
\dt
SELECT * FROM gorp_migrations;
SELECT * FROM scores;
```

