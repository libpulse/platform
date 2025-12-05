#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="../../.env.dev"

if [ -f "$ENV_FILE" ]; then
  echo "Loading env from $ENV_FILE"
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

echo "Starting LibPulse API..."
if command -v air >/dev/null 2>&1; then
  air
else
  go run .
fi