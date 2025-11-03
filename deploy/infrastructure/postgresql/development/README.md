# PostgreSQL Infrastructure for Rankr Microservices

This directory contains the Docker-based PostgreSQL setup for **Rankr microservices**. It includes an initialization script that automatically creates databases and users for all services using environment variables from `.env`.

---

## Structure

```
.
├── Dockerfile                          # PostgreSQL image with init script
├── docker-compose.postgresql.yml       # Docker Compose configuration
├── init-services-db.sh                 # Database initialization script
├── init-services-db.sql               # Alternative SQL-based init (optional)
├── .env-example                        # Template for environment variables
└── README.md                           # This file
```

---

## Quick Start

### 1. Setup Environment Variables

Copy the example file and configure your credentials:

```bash
cp .env-example .env
```

Edit `.env` with secure passwords:

```dotenv
# PostgreSQL Admin
POSTGRES_USER=rankr_admin
POSTGRES_PASSWORD=your_strong_admin_password
POSTGRES_DB=postgres

# Service Users
LEADERBOARDSTAT_USER=leaderboardstat_user
LEADERBOARDSTAT_PASS=secure_password_1

CONTRIBUTOR_USER=contributor_user
CONTRIBUTOR_PASS=secure_password_2

PROJECT_USER=project_user
PROJECT_PASS=secure_password_3

WEBHOOK_USER=webhook_user
WEBHOOK_PASS=secure_password_4

LEADERBOARDSCORING_USER=leaderboardscoring_user
LEADERBOARDSCORING_PASS=secure_password_5
```

> **⚠️ Security Note:** Never commit `.env` to version control. Add it to `.gitignore`.

---

### 2. Create Docker Network and Volume

Create the required external network and volume:

```bash
# Create network
docker network create rankr-shared-development-network

# Create volume
docker volume create rankr-shared-postgres-data
```

---

### 3. Build and Start PostgreSQL

Build the Docker image and start the container:

```bash
# Build the image
docker compose -f docker-compose.postgresql.yml build --no-cache

# Start the container
docker compose -f docker-compose.postgresql.yml up -d
```

---

### 4. Verify Installation

Check that all databases were created successfully:

```bash
# View container logs
docker logs rankr-shared-postgres

# Connect to PostgreSQL
docker exec -it rankr-shared-postgres psql -U rankr_admin -d postgres

# List all databases
\l

# List all users
\du

# Exit psql
\q
```

You should see:
- 5 service databases (leaderboardstat_db, contributor_db, project_db, webhook_db, leaderboardscoring_db)
- 5 service users with appropriate ownership

---

## Adding a New Service

### Option 1: Using Shell Script (Recommended)

Edit `init-services-db.sh` and add a new service call:

```bash
# Add after existing services
create_service_db "$NEWSERVICE_USER" "$NEWSERVICE_PASS" "newservice_db"
```

Add environment variables to `.env`:

```dotenv
NEWSERVICE_USER=newservice_user
NEWSERVICE_PASS=secure_password_6
```

### Option 2: Using SQL Script

If you prefer using `init-services-db.sql`, add a new block:

```sql
-- New Service
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'newservice_user') THEN
        CREATE USER newservice_user WITH PASSWORD 'newservice_pass';
        RAISE NOTICE 'User newservice_user created';
    ELSE
        RAISE NOTICE 'User newservice_user already exists';
    END IF;
END
$$;

SELECT 'CREATE DATABASE newservice_db OWNER newservice_user'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'newservice_db')
\gexec

GRANT ALL PRIVILEGES ON DATABASE newservice_db TO newservice_user;
```

Then update your Dockerfile to use the SQL script:

```dockerfile
FROM postgres:17.2-alpine
COPY init-services-db.sql /docker-entrypoint-initdb.d/
```

---

## Rebuild and Reset

### Rebuild After Changes

If you modify initialization scripts:

```bash
# Stop and remove container
docker compose -f docker-compose.postgresql.yml down

# Remove the volume (WARNING: deletes all data)
docker volume rm rankr-shared-postgres-data

# Recreate volume
docker volume create rankr-shared-postgres-data

# Rebuild and restart
docker compose -f docker-compose.postgresql.yml build --no-cache
docker compose -f docker-compose.postgresql.yml up -d
```

### Keep Data, Just Restart

If you only need to restart without losing data:

```bash
docker compose -f docker-compose.postgresql.yml restart
```

---

## Security Best Practices

1. **Never commit `.env`** - Add to `.gitignore`
2. **Use strong passwords** - Generate random passwords for production
3. **Rotate credentials regularly** - Update passwords periodically
4. **Limit network access** - Use Docker networks to isolate services
5. **Use secrets management** - Consider Docker secrets or vault solutions for production

---

## Database Connection Info

Each microservice should connect using its own credentials:

| Service | Database | User | Port (Host) |
|---------|----------|------|-------------|
| Leaderboard Stats | leaderboardstat_db | leaderboardstat_user | 5439 |
| Contributor | contributor_db | contributor_user | 5439 |
| Project | project_db | project_user | 5439 |
| Webhook | webhook_db | webhook_user | 5439 |
| Leaderboard Scoring | leaderboardscoring_db | leaderboardscoring_user | 5439 |

**Connection string example:**
```
postgresql://leaderboardstat_user:secure_password_1@localhost:5439/leaderboardstat_db
```

---

## Troubleshooting

### Container Keeps Restarting

Check logs for errors:
```bash
docker logs rankr-shared-postgres -f
```

### Database Not Initialized

Ensure the volume is clean before first run:
```bash
docker volume rm rankr-shared-postgres-data
docker volume create rankr-shared-postgres-data
```

### Permission Denied Errors

Make sure the initialization script is executable:
```bash
chmod +x init-services-db.sh
```

### Environment Variables Not Working

Verify `.env` file is in the same directory as `docker-compose.postgresql.yml` and properly formatted.
