# Terraform – No Small Parts

Deploy the service (Cloud Run API + frontend, Cloud SQL, Pub/Sub, GCS, Cloudflare) as a **reusable module**. One environment: **prod**.

## Layout

- **`modules/nosmallparts/`** – Module that creates one environment’s resources (parameterized by variables).
- **`environments/prod/`** – Prod Terraform root: providers, backend (TFC workspace `backend`), and a single module call.
- **`environments/README.md`** – How to deploy prod and add more envs later.

**Deploy:**

```bash
cd terraform/environments/prod
terraform init -upgrade
terraform plan
terraform apply
```

- **Terraform Cloud (remote runs):** In the workspace Variables tab, add variables with the **exact** Terraform names (snake_case), e.g. `cloudflare_api_token`, `cloudflare_zone_id`, `google_cloud_credentials` — not `CLOUDFLARE_API_TOKEN`. Mark secrets as Sensitive. These are only used when plan/apply run **in** TFC (e.g. VCS-triggered or “Run plan” in the UI).
- **Local runs:** When you run `terraform plan` from your machine, TFC does not inject variables. Use `terraform.tfvars` (from the example) or `export TF_VAR_cloudflare_api_token=...` etc. so you are not prompted.

## What the module creates

- **GCP:** Cloud SQL (Postgres 17), Pub/Sub, Artifact Registry, GCS buckets (media, userdata), VPC connector, **Cloud Run** (API + frontend)
- **Cloudflare:** DNS records (optional; set `cloudflare_zone_id`)

Cloud Run services: `nosmallparts-api-<environment>`, `nosmallparts-frontend-<environment>`.

## Adding another environment

See **`environments/README.md`**: copy `environments/prod` to a new dir, set a different TFC workspace and variables, then `terraform init` and `apply` there.

## Commands (from repo root)

- `make terraformat` – format all Terraform under `terraform/`
- `make lint_terraform` – init, validate, fmt-check (run from `environments/prod` for full lint)
