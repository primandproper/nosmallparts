#!/usr/bin/env bash
set -euo pipefail

# Lint Terraform files
# Usage: lint_terraform.sh <terraform_dir>

TERRAFORM_DIR="${1:-deploy/environments/dev/terraform}"

(
  cd "${TERRAFORM_DIR}" || exit 1
  terraform init -upgrade
  terraform validate
  terraform fmt
  terraform fmt -check
)
