#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="../../.env.dev"

if [ -f "$ENV_FILE" ]; then
  echo "Loading env from $ENV_FILE"
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

echo "Starting LibPulse API..."
go run .