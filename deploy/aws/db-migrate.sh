#!/usr/bin/env sh
set -eu

: "${ADMIN_DATABASE_URL:?ADMIN_DATABASE_URL is required}"
: "${APP_DATABASE_URL:?APP_DATABASE_URL is required}"

APP_DATABASE_NAME="${APP_DATABASE_NAME:-commerce}"

case "${APP_DATABASE_NAME}" in
  *[!A-Za-z0-9_]* | "")
    echo "APP_DATABASE_NAME must contain only letters, numbers, and underscores" >&2
    exit 1
    ;;
esac

if psql "${ADMIN_DATABASE_URL}" -tAc "SELECT 1 FROM pg_database WHERE datname = '${APP_DATABASE_NAME}'" | grep -q 1; then
  echo "database ${APP_DATABASE_NAME} already exists"
else
  echo "creating database ${APP_DATABASE_NAME}"
  psql "${ADMIN_DATABASE_URL}" -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"${APP_DATABASE_NAME}\""
fi

for migration in /app/db/migrations/*.sql; do
  echo "applying ${migration}"
  psql "${APP_DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${migration}"
done

echo "database migration complete"
