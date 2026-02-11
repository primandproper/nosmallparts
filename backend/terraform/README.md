# Backend Terraform (API + shared infra)

Deploys the API and shared GCP resources: VPC, Cloud SQL, Pub/Sub, Artifact Registry, GCS, Cloud Run API, secrets, Cloudflare.

**Apply first.** Frontend Terraform depends on outputs from this root (e.g. `cloud_run_api_url`, `artifact_registry_repository_id`). Use TFC workspace `backend` or run from `environments/prod` with `terraform.tfvars` / `TF_VAR_*`.

**From repo root:** `make terraformat` and `make lint_terraform` cover this directory and `frontend/terraform/`.

### Running plan/apply/destroy locally

Variable names are **snake_case**; Terraform maps env vars as `TF_VAR_<variable_name>`. Use the exact name:

```bash
cd backend/terraform/environments/prod
TF_VAR_project_id="no-small-parts" TF_VAR_google_cloud_credentials="$(cat ~/Downloads/no-small-parts-0764d3d2458f.json
)" terraform plan
docker build \
-f backend/deploy/dockerfiles/api_server.Dockerfile \
-t us-central1-docker.pkg.dev/no-small-parts/containers/nosmallparts-api:latest \
./backend
docker push us-central1-docker.pkg.dev/no-small-parts/containers/nosmallparts-api:latest
```

Use **`TF_VAR_google_cloud_credentials`** (lowercase), not `TF_VAR_GOOGLE_CLOUD_CREDENTIALS`. The prompt for sensitive vars often doesn’t accept paste; set the var via env (or a `terraform.tfvars` file that’s in `.gitignore`) so Terraform doesn’t prompt.
