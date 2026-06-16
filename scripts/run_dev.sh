#!/usr/bin/env bash
set -euo pipefail

if [ ! -f .env ]; then
  echo "Error: .env file not found. Run: cp .env.example .env"
  exit 1
fi

set -a && source .env && set +a

echo "Starting YeboBank in development mode..."
go run ./cmd/server
