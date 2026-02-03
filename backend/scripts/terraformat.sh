#!/usr/bin/env bash
set -euo pipefail

# Format Terraform files
# Usage: terraformat.sh [terraform_dir]

TERRAFORM_DIR="${1:-deploy/environments/dev/terraform}"

(
  cd "${TERRAFORM_DIR}" || exit 1
  terraform fmt
)
