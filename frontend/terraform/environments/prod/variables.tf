# -----------------------------------------------------------------------------
# Set frontend_backend_http_url and frontend_grpc_backend_url from backend
# Terraform outputs (or TFC). artifact_repository_id should match backend output.
# -----------------------------------------------------------------------------

variable "environment" {
  type        = string
  default     = "prod"
  description = "Environment name (prod)"
}

variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "gcp_region" {
  type        = string
  default     = "us-central1"
  description = "GCP region"
}

variable "gcp_main_zone" {
  type        = string
  default     = "us-central1-a"
  description = "GCP zone"
}

variable "google_cloud_credentials" {
  type      = string
  sensitive = true
}

# From backend Terraform output (or TFC)
variable "artifact_repository_id" {
  type        = string
  default     = "containers"
  description = "Artifact Registry repo ID (from backend Terraform output)"
}

variable "frontend_backend_http_url" {
  type        = string
  description = "Backend API URL (e.g. backend Terraform output cloud_run_api_url)"
}

variable "frontend_grpc_backend_url" {
  type        = string
  description = "Backend gRPC address (host:port)"
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

variable "ADSENSE_PUBLISHER_ID" {
  type        = string
  default     = ""
  description = "Google AdSense publisher ID (ca-pub-...)"
}

variable "ADSENSE_SLOT_ID" {
  type        = string
  default     = ""
  description = "Google AdSense ad slot ID"
}
