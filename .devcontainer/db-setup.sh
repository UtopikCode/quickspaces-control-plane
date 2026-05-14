#!/usr/bin/env bash
set -euo pipefail

cd /workspace

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  . .env
  set +a
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is not set; skipping database initialization and migration."
  exit 0
fi

PGHOST="${POSTGRES_HOST:-postgres}"
PGPORT="${POSTGRES_PORT:-5432}"
PGUSER="${POSTGRES_USER:-postgres}"
PGDATABASE="${POSTGRES_DB:-quickspaces}"

export PGPASSWORD="${POSTGRES_PASSWORD:-}"

echo "Waiting for PostgreSQL at ${PGHOST}:${PGPORT}..."
for i in $(seq 1 30); do
  if pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" >/dev/null 2>&1; then
    echo "PostgreSQL is ready"
    break
  fi
  echo "Waiting for Postgres... ($i/30)"
  sleep 2
  if [ "$i" -eq 30 ]; then
    echo "ERROR: Postgres did not become ready in time" >&2
    exit 1
  fi
done

echo "Initializing DB schema if required..."
if ! make init-db; then
  echo "init-db failed or was not needed; continuing with migrate-ent"
fi

echo "Applying Ent migrations"
make migrate-ent
