FROM postgres:17.2-alpine

COPY init-services-db.sh /docker-entrypoint-initdb.d/

RUN chmod +x /docker-entrypoint-initdb.d/init-services-db.sh