#!/usr/bin/env bash
set -euo pipefail

# Check struct usage
# Usage: struct_usage_check.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/verygoodsoftwarenotvirus/nosmallparts/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/struct_usage_checker"
