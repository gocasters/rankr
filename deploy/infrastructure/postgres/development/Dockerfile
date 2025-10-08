FROM postgres:17.2-alpine

# Copy initialization script
COPY init-db.sql /docker-entrypoint-initdb.d/

# Set environment variables
ENV POSTGRES_USER=rankr_admin
ENV POSTGRES_PASSWORD=rankr_password_123