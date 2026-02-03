#!/usr/bin/env bash
set -euo pipefail

# Generate querier code
# Usage: querier.sh <sql_generator_command> <generated_dir>

SQL_GENERATOR="${1}"
GENERATED_DIR="${2:-internal/database/postgres/generated}"

rm -rf "${GENERATED_DIR}"/*.go
${SQL_GENERATOR} generate --no-remote
