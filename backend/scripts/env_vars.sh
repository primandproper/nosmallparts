#!/usr/bin/env bash
set -euo pipefail

# Generate and format env vars
# Usage: env_vars.sh <package_prefix> <gofmt_command>

PACKAGE_PREFIX="${1:-github.com/verygoodsoftwarenotvirus/nosmallparts/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/valid_env_vars"
