#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

DATABASE_URL="${DATABASE_URL:-postgres://commerce:commerce@localhost:5432/commerce?sslmode=disable}"

for migration in db/migrations/*.sql; do
  echo "applying ${migration}"
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${migration}"
done
