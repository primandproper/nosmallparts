#!/usr/bin/env bash
set -euo pipefail

# Lint SQL queries using sqlc
# Usage: queries_lint.sh <container_runner> <sql_generator_image> <sql_generator_command>

CONTAINER_RUNNER="${1:-docker}"
SQL_GENERATOR_IMAGE="${2:-sqlc/sqlc:1.26.0}"
SQL_GENERATOR="${3}"

"${CONTAINER_RUNNER}" pull --quiet "${SQL_GENERATOR_IMAGE}"
${SQL_GENERATOR} compile --no-remote
${SQL_GENERATOR} vet --no-remote
