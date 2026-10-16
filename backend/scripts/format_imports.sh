#!/usr/bin/env bash
set -euo pipefail

# Format Go imports using gci
# Usage: format_imports.sh <package_prefix> <project_root>
#
# Which files those are is go_files.sh's question, not this one's.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_PREFIX="${1:-github.com/primandproper/nosmallparts/backend}"
PROJECT_ROOT="${2:-$(pwd)}"

# Through a file rather than `< <(go_files.sh)`: process substitution discards
# the exit status of what it runs, so a failure to produce the list would read
# here as a list of no files, and formatting nothing would look like success.
file_list="$(mktemp)"
trap 'rm -f "${file_list}"' EXIT

"${SCRIPT_DIR}/go_files.sh" "${PROJECT_ROOT}" >"${file_list}"

# Every Go file the module owns, passed to gci
go_files=()
while IFS= read -r -d '' file; do
  go_files+=("${file}")
done <"${file_list}"

if [ ${#go_files[@]} -gt 0 ]; then
  # The org section is spelled out rather than derived from PACKAGE_PREFIX: this module
  # lives at <org>/nosmallparts/backend, so its dirname is the repository, not the org.
  # Sections must stay in sync with the gci block in .golangci.yml, or `make format` and
  # `make lint` disagree about import order.
  go tool gci write \
    --section standard \
    --section "prefix(${PACKAGE_PREFIX})" \
    --section "prefix(github.com/primandproper)" \
    --section default \
    --custom-order \
    "${go_files[@]}"
fi
