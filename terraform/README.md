# Terraform (deprecated – infra moved)

Infrastructure is split by app:

- **Backend (API + shared infra):** [backend/terraform/](../backend/terraform/)  
  Deploys VPC, Cloud SQL, Pub/Sub, Artifact Registry, GCS, Cloud Run API, secrets, Cloudflare.  
  Apply **first**. TFC workspace: `backend` (org `nosmallparts`).

- **Frontend (Next.js):** [frontend/terraform/](../frontend/terraform/)  
  Deploys the Cloud Run frontend service only.  
  Apply **after** backend. Set `frontend_backend_http_url` and `frontend_grpc_backend_url` from backend Terraform outputs (or TFC). TFC workspace: `frontend`.

**Deploy order:** Backend → Frontend. Frontend variables can use backend outputs (e.g. in CI or TFC).

**Commands (from repo root):**  
`make terraformat` and `make lint_terraform` run over both `backend/terraform/` and `frontend/terraform/`.
