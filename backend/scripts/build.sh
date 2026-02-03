#!/usr/bin/env bash
set -euo pipefail

# Build all packages
# Usage: build.sh <package_list>

PACKAGE_LIST="${1}"

# shellcheck disable=SC2086
go build ${PACKAGE_LIST}
