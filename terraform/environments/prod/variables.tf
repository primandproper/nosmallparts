# -----------------------------------------------------------------------------
# Terraform Cloud
# In TFC workspace → Variables, use these *exact* keys (Terraform uses snake_case):
#   cloudflare_api_token, cloudflare_zone_id, google_cloud_credentials
#   (+ project_id, media_bucket_name, userdata_bucket_name, frontend_*, sendgrid_*, etc.)
# Mark secrets as "Sensitive". Do not use UPPERCASE keys (e.g. CLOUDFLARE_API_TOKEN).
# TFC variables are only injected when the run executes in TFC (remote run). For
# local `terraform plan`, use terraform.tfvars or TF_VAR_cloudflare_api_token etc.
# -----------------------------------------------------------------------------

# -----------------------------------------------------------------------------
# Environment
# -----------------------------------------------------------------------------
variable "environment" {
  type        = string
  default     = "prod"
  description = "Environment name (prod)"
}

# -----------------------------------------------------------------------------
# GCP
# -----------------------------------------------------------------------------
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
  # Set in TFC (sensitive)
}

# -----------------------------------------------------------------------------
# Networking
# -----------------------------------------------------------------------------
variable "private_network_name" {
  type        = string
  default     = "private-network"
  description = "VPC network name for private service connection"
}

variable "private_ip_address_name" {
  type        = string
  default     = "private-ip-address"
  description = "Global address name for service networking"
}

variable "vpc_connector_ip_cidr_range" {
  type        = string
  default     = "10.8.0.0/28"
  description = "CIDR for Cloud Run VPC connector"
}

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------
variable "database_name" {
  type    = string
  default = "no-small-parts"
}

variable "db_instance_name" {
  type        = string
  default     = "prod"
  description = "Cloud SQL instance name"
}

variable "db_ssl_common_name" {
  type        = string
  default     = null
  description = "SSL cert common name (default: project_id-environment)"
}

# -----------------------------------------------------------------------------
# Pub/Sub
# -----------------------------------------------------------------------------
variable "pubsub_topic_prefix" {
  type        = string
  default     = ""
  description = "Prefix for topic names when multiple envs share a project"
}

# -----------------------------------------------------------------------------
# Artifact Registry
# -----------------------------------------------------------------------------
variable "artifact_repository_id" {
  type    = string
  default = "containers"
}

# -----------------------------------------------------------------------------
# Cloudflare
# -----------------------------------------------------------------------------
variable "cloudflare_api_token" {
  type      = string
  sensitive = true
  # Set in TFC (sensitive)
}

variable "cloudflare_zone_id" {
  type        = string
  default     = ""
  description = "Cloudflare zone ID; leave empty to skip DNS records"
}

variable "cloudflare_db_record_name" {
  type    = string
  default = "db"
}

variable "cloudflare_media_record_name" {
  type    = string
  default = "media"
}

variable "cloudflare_userdata_record_name" {
  type    = string
  default = "userdata"
}

# -----------------------------------------------------------------------------
# Storage
# -----------------------------------------------------------------------------
variable "media_bucket_name" {
  type        = string
  description = "GCS bucket name for media (globally unique)"
}

variable "userdata_bucket_name" {
  type        = string
  description = "GCS bucket name for user data (globally unique)"
}

variable "cors_origins" {
  type        = list(string)
  default     = ["https://nosmall.parts"]
  description = "CORS origins for storage buckets"
}

# -----------------------------------------------------------------------------
# API secrets (set in TFC as sensitive)
# -----------------------------------------------------------------------------
variable "sendgrid_api_token" {
  type      = string
  sensitive = true
}

variable "segment_api_token" {
  type      = string
  sensitive = true
}

variable "posthog_api_key" {
  type      = string
  sensitive = true
}

variable "posthog_personal_api_key" {
  type      = string
  sensitive = true
}

variable "google_sso_oauth2_client_id" {
  type      = string
  sensitive = true
}

variable "google_sso_oauth2_client_secret" {
  type      = string
  sensitive = true
}

# -----------------------------------------------------------------------------
# Frontend
# -----------------------------------------------------------------------------
variable "frontend_backend_http_url" {
  type        = string
  description = "Backend API URL for the Next.js app (e.g. Cloud Run API URL)"
}

variable "frontend_grpc_backend_url" {
  type        = string
  description = "Backend gRPC address for the Next.js app"
}

variable "frontend_oauth2_client_id" {
  type      = string
  sensitive = true
  # Set in TFC (sensitive)
}

variable "frontend_oauth2_client_secret" {
  type      = string
  sensitive = true
  # Set in TFC (sensitive)
}

variable "frontend_next_public_backend_ws_url" {
  type        = string
  default     = ""
  description = "WebSocket URL for stages (browser)"
}
