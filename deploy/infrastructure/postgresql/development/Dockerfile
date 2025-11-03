FROM postgres:17.2-alpine

# Copy initialization script
COPY init-services-db.sh /docker-entrypoint-initdb.d/

# Make it executable
RUN chmod +x /docker-entrypoint-initdb.d/init-services-db.sh