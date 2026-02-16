# -----------------------------------------------------------------------------
# Prod environment – frontend (Next.js) only. Run from this directory.
# -----------------------------------------------------------------------------
# Apply backend Terraform first; then set frontend_backend_http_url (and optionally
# artifact_repository_id) from backend outputs, or use TFC variables.
# terraform init && terraform plan && terraform apply

terraform {
  required_version = ">= 1.10"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.14"
    }
  }
}

provider "google" {
  project     = var.project_id
  region      = var.gcp_region
  zone        = var.gcp_main_zone
  credentials = var.google_cloud_credentials
}

module "frontend" {
  source = "../../modules/frontend"

  environment            = var.environment
  project_id             = var.project_id
  gcp_region             = var.gcp_region
  artifact_repository_id = var.artifact_repository_id

  frontend_backend_http_url           = var.frontend_backend_http_url
  frontend_grpc_backend_url           = var.frontend_grpc_backend_url
  frontend_oauth2_client_id           = var.frontend_oauth2_client_id
  frontend_oauth2_client_secret       = var.frontend_oauth2_client_secret
  frontend_next_public_backend_ws_url = var.frontend_next_public_backend_ws_url
  ADSENSE_PUBLISHER_ID                = var.ADSENSE_PUBLISHER_ID
  ADSENSE_SLOT_ID                     = var.ADSENSE_SLOT_ID
}

output "cloud_run_frontend_url" {
  value       = module.frontend.cloud_run_frontend_url
  description = "Cloud Run Frontend service URL"
}

output "environment" {
  value       = module.frontend.environment
  description = "Environment name"
}
