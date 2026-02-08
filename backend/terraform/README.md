# Backend Terraform (API + shared infra)

Deploys the API and shared GCP resources: VPC, Cloud SQL, Pub/Sub, Artifact Registry, GCS, Cloud Run API, secrets, Cloudflare.

**Apply first.** Frontend Terraform depends on outputs from this root (e.g. `cloud_run_api_url`, `artifact_registry_repository_id`). Use TFC workspace `backend` or run from `environments/prod` with `terraform.tfvars` / `TF_VAR_*`.

**From repo root:** `make terraformat` and `make lint_terraform` cover this directory and `frontend/terraform/`.
