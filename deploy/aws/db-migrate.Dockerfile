FROM postgres:16-alpine

WORKDIR /app

COPY db/migrations ./db/migrations
COPY deploy/aws/db-migrate.sh ./db-migrate.sh
RUN chmod +x ./db-migrate.sh

ENTRYPOINT ["/app/db-migrate.sh"]
