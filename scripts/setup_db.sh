#!/usr/bin/env bash
set -euo pipefail

DB_NAME="${DB_NAME:-yebobank}"
DB_USER="${DB_USER:-yebobank}"
DB_PASSWORD="${DB_PASSWORD:-changeme}"

echo "Creating PostgreSQL user and database..."

psql -U postgres <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE ROLE ${DB_USER} WITH LOGIN PASSWORD '${DB_PASSWORD}';
  END IF;
END
\$\$;

SELECT 'CREATE DATABASE ${DB_NAME} OWNER ${DB_USER}'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_NAME}')\gexec
SQL

echo "Done. Database '${DB_NAME}' owned by '${DB_USER}'."
