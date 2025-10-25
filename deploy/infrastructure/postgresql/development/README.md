# PostgreSQL Infrastructure for Rankr Microservices

This directory contains the Docker Compose setup for PostgreSQL used by **Rankr microservices**. It includes a base
database initialization script (`init-services-db.sql`) and a `.env` template to securely configure database
credentials.

---

## Usage Instructions

### 1. Configure `.env`

* Copy the example file to `.env`:

```bash
cp .env-example .env
```

* Edit `.env` to add credentials for your service(s):

```dotenv
# Example
LEADERBOARDSTAT_USER=leaderboardstat_user
LEADERBOARDSTAT_PASS=supersecret1
CONTRIBUTOR_USER=contributor_user
CONTRIBUTOR_PASS=supersecret2
PROJECT_USER=project_user
PROJECT_PASS=supersecret3
WEBHOOK_USER=webhook_user
WEBHOOK_PASS=supersecret4
LEADERBOARDSCORING_USER=leaderboardscoring_user
LEADERBOARDSCORING_PASS=supersecret5
```

> **Important:** Never commit `.env` to source control. Keep it local or use a secure vault.

---

### 2. Add Your Service to `init-services-db.sql`

* If your service needs its own database and user, append a block in `init-services-db.sql` following the pattern of
  existing users:

```sql
DO
$$
BEGIN
    IF
NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'MY_SERVICE_USER') THEN
        EXECUTE format(
            'CREATE USER %I WITH PASSWORD %L',
            :'MY_SERVICE_USER',
            :'MY_SERVICE_PASS'
        );
END IF;
END
$$;

SELECT format(
               'CREATE DATABASE my_service_db OWNER %I',
               :'MY_SERVICE_USER'
       ) WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'my_service_db') \gexec;

GRANT
ALL
PRIVILEGES
ON
DATABASE
my_service_db TO :'MY_SERVICE_USER';
```

* Then, add corresponding environment variables to your `.env`:

```dotenv
MY_SERVICE_USER=my_service_user
MY_SERVICE_PASS=supersecret6
```

---

### 3. Start PostgreSQL

Run the PostgreSQL container using Docker Compose:

```bash
docker compose up -d
```

* The container will initialize databases and users from `.env` via `init-services-db.sql`.

---

### 4. Notes

* This setup ensures **idempotent initialization** — safe to re-run multiple times.
* Use `.env` for all credentials to avoid hardcoding passwords in SQL.
* Each microservice should have a dedicated user and database to prevent conflicts.

---

This README makes it clear for developers:

1. Copy `.env-example` → `.env`.
2. Add credentials for any new services.
3. Update `init-services-db.sql` with new DB/user logic.
4. Start the container via Docker Compose.

