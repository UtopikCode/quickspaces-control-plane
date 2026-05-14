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
  echo "DATABASE_URL is not set; skipping database initialization."
  exit 0
fi

MONGO_HOST="${MONGO_HOST:-mongo}"
MONGO_PORT="${MONGO_PORT:-27017}"

echo "Waiting for MongoDB at ${MONGO_HOST}:${MONGO_PORT}..."
for i in $(seq 1 30); do
  if bash -c "cat < /dev/null > /dev/tcp/${MONGO_HOST}/${MONGO_PORT}" >/dev/null 2>&1; then
    echo "MongoDB is ready"
    break
  fi
  echo "Waiting for MongoDB... ($i/30)"
  sleep 2
  if [ "$i" -eq 30 ]; then
    echo "ERROR: MongoDB did not become ready in time" >&2
    exit 1
  fi
 done

echo "Initializing MongoDB schema and indexes"
make init-db
