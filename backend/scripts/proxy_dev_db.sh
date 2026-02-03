#!/usr/bin/env bash
set -euo pipefail

# Proxy dev database
# Usage: proxy_dev_db.sh <gcp_project_name>

GCP_PROJECT_NAME="${1:-dinner-done-better-dev}"

cloud_sql_proxy "${GCP_PROJECT_NAME}:us-central1:dev" --port 5434 --gcloud-auth
