#!/usr/bin/env bash
set -euo pipefail

# Generate configs
# Usage: configs.sh <package_prefix>

PACKAGE_PREFIX="${1:-github.com/verygoodsoftwarenotvirus/nosmallparts/backend}"

go run "${PACKAGE_PREFIX}/cmd/tools/codegen/configs"
