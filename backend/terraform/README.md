# Backend Terraform (API + shared infra)

Deploys the API and shared GCP resources: VPC, Cloud SQL, Pub/Sub, Artifact Registry, GCS, Cloud Run API, secrets, Cloudflare.

**Apply first.** Frontend Terraform depends on outputs from this root (e.g. `cloud_run_api_url`, `artifact_registry_repository_id`). Use TFC workspace `backend` or run from `environments/prod` with `terraform.tfvars` / `TF_VAR_*`.

**From repo root:** `make terraformat` and `make lint_terraform` cover this directory and `frontend/terraform/`.

### How the Cloud Run API service gets its config

1. **Source of the JSON** – In `environments/prod`, the API config is either from a file (`api_service_config_file`, e.g. `service-config.json`) or the variable `api_service_config_json` (default `"{}"`).
2. **Where it's stored** – That JSON is written into **Secret Manager** as `prod-api-service-config` (`modules/api/secrets.tf`).
3. **How Cloud Run sees it** – The secret is mounted so the item `service-config.json` appears at **`/etc/config/service-config.json`**. Env **`CONFIGURATION_FILEPATH`** is set to that path.
4. **How the app loads it** – On startup the API calls `config.LoadConfigFromEnvironment[APIServiceConfig]()`: it reads the file at `CONFIGURATION_FILEPATH`, then overlays **`NO_SMALL_PARTS_*`** env vars. So the file is the base config; env vars override.

If Terraform ran without `api_service_config_file` set, the secret contains `"{}"` and only env vars supply config. The deploy workflow now sets `TF_VAR_api_service_config_file=service-config.json` after copying the dev config, so the secret gets real JSON.

### Running plan/apply/destroy locally

Variable names are **snake_case**; Terraform maps env vars as `TF_VAR_<variable_name>`. Use the exact name:

```bash
cd backend/terraform/environments/prod
TF_VAR_project_id="no-small-parts" TF_VAR_google_cloud_credentials="$(cat ~/Downloads/no-small-parts-0764d3d2458f.json
)" terraform plan
docker build \
-f backend/deploy/dockerfiles/api_server.Dockerfile \
-t us-central1-docker.pkg.dev/no-small-parts/containers/nosmallparts-api:latest \
./backend && docker push us-central1-docker.pkg.dev/no-small-parts/containers/nosmallparts-api:latest
```

Use **`TF_VAR_google_cloud_credentials`** (lowercase), not `TF_VAR_GOOGLE_CLOUD_CREDENTIALS`. The prompt for sensitive vars often doesn’t accept paste; set the var via env (or a `terraform.tfvars` file that’s in `.gitignore`) so Terraform doesn’t prompt.
