# Frontend Terraform

Deploys the **frontend** (Next.js) Cloud Run service only. Backend infra (API, DB, Artifact Registry, etc.) is managed by [backend/terraform](../backend/terraform/).

- **Apply backend first** so the API and Artifact Registry exist.
- Set `frontend_backend_http_url` from backend Terraform outputs (or TFC variables). Set `artifact_repository_id` to match backend output so the frontend image path is correct.

**Layout:** `modules/frontend/` (Cloud Run service), `environments/prod/` (prod config).

**Run:** From `environments/prod/`: `terraform init && terraform plan && terraform apply`.
