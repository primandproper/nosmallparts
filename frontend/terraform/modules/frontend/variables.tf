# -----------------------------------------------------------------------------
# Environment and GCP
# -----------------------------------------------------------------------------
variable "environment" {
  description = "Environment name (e.g. prod). Used for resource naming."
  type        = string
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region for Cloud Run"
  type        = string
}

# -----------------------------------------------------------------------------
# Artifact Registry (repo created by backend Terraform; pass repository_id)
# -----------------------------------------------------------------------------
variable "artifact_repository_id" {
  description = "Artifact Registry repository ID (from backend Terraform output)"
  type        = string
}

# -----------------------------------------------------------------------------
# Frontend env vars (backend URL and OAuth2 from backend / TFC)
# -----------------------------------------------------------------------------
variable "frontend_backend_http_url" {
  description = "Backend API URL (e.g. from backend Terraform output cloud_run_api_url)"
  type        = string
}

variable "frontend_oauth2_client_id" {
  type      = string
  sensitive = true
}

variable "frontend_oauth2_client_secret" {
  type      = string
  sensitive = true
}

variable "frontend_next_public_backend_ws_url" {
  type        = string
  default     = ""
  description = "WebSocket URL for stages (browser)"
}
