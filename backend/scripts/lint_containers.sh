#!/usr/bin/env bash
set -euo pipefail

# Lint Docker containers using conftest
# Usage: lint_containers.sh <container_runner> <linter_image> <run_container_as_user> <project_root>

CONTAINER_RUNNER="${1:-docker}"
LINTER_IMAGE="${2:-openpolicyagent/conftest:v0.54.0}"
RUN_CONTAINER_AS_USER="${3}"
PROJECT_ROOT="${4:-$(pwd)}"

"${CONTAINER_RUNNER}" pull --quiet "${LINTER_IMAGE}"

# Find all Dockerfiles and pass them to conftest
dockerfiles=()
while IFS= read -r file; do
  dockerfiles+=("${file}")
done < <(find "${PROJECT_ROOT}" -type f -name "*.Dockerfile")

if [ ${#dockerfiles[@]} -gt 0 ]; then
  # RUN_CONTAINER_AS_USER is a Make variable that expands to a command string
  # We need to execute it as a command, so we use eval with proper quoting
  # shellcheck disable=SC2086
  eval "${RUN_CONTAINER_AS_USER} \"${LINTER_IMAGE}\" test --policy containers.rego ${dockerfiles[*]}"
fi
