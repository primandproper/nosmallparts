#!/usr/bin/env bash
set -euo pipefail

# Rewire dependency injection targets
# Usage: rewire.sh <package_prefix> <target1> [target2] ...

PACKAGE_PREFIX="${1}"
shift
TARGETS=("$@")

for target in "${TARGETS[@]}"; do
  rm -f "${PACKAGE_PREFIX}/${target}/wire_gen.go"
  wire gen "${PACKAGE_PREFIX}/${target}"
done
