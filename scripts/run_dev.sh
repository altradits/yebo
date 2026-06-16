#!/usr/bin/env bash
set -euo pipefail

if [ ! -f .env ]; then
  echo "Error: .env file not found. Run: cp .env.example .env"
  exit 1
fi

# Start Postgres container if it exists but isn't running
if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q '^yebo-postgres$'; then
  if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -q '^yebo-postgres$'; then
    echo "Starting yebo-postgres container..."
    docker start yebo-postgres
    sleep 2
  fi
fi

set -a && source .env && set +a

echo "Starting YeboBank in development mode..."
go run ./cmd/server
