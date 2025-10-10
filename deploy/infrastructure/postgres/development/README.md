# PostgreSQL Database Setup

## Quick Start

```bash
# Make setup script executable
chmod +x ./deploy/setup-network.bash
```

```bash
# Start all services with shared PostgreSQL
./deploy/docker-compose-dev.bash

# Stop and clean up
./deploy/docker-compose-dev.bash down -v

# Show containers logs
./deploy/docker-compose-dev.bash logs

# Show an specific container logs
./deploy/docker-compose-dev.bash logs leaderboardstat-service
```


```bash
# Start only leaderboardstat service with database
./deploy/docker-compose-dev-leaderboardstat-local.bash up --build

# Rebuild and restart just leaderboardstat service
./deploy/docker-compose-dev.bash up -d leaderboardstat-service --build

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

```psql
# Inside psql, check tables
\dt
SELECT * FROM gorp_migrations;
SELECT * FROM scores;
```

```bash
# List all containers in our rankr-development-network network
docker network inspect rankr-development-network --format='{{range .Containers}}{{.Name}} {{.IPv4Address}}{{"\n"}}{{end}}'
```

