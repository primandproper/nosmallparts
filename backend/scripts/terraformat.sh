#!/usr/bin/env bash
set -euo pipefail

# Format Terraform files
# Usage: terraformat.sh [terraform_dir]

# Terraform is at repo root; default when run from backend is ../../terraform
TERRAFORM_DIR="${1:-../../terraform}"

(
  cd "${TERRAFORM_DIR}" || exit 1
  terraform fmt
)
