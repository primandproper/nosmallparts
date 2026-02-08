# -----------------------------------------------------------------------------
# Terraform Cloud: use exact snake_case keys (e.g. cloudflare_api_token).
# Mark secrets as Sensitive. For local runs use terraform.tfvars or TF_VAR_*.
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
  type        = string
  sensitive   = true
  description = "Cloudflare API token (set in workspace variable set or TF_VAR_cloudflare_api_token)"
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
  description = "GCS bucket name for media (globally unique). Set in workspace variable set."
}

variable "userdata_bucket_name" {
  type        = string
  description = "GCS bucket name for user data (globally unique). Set in workspace variable set."
}

variable "cors_origins" {
  type        = list(string)
  default     = ["https://nosmall.parts"]
  description = "CORS origins for storage buckets"
}

# -----------------------------------------------------------------------------
# API config (mounted as secret at /etc/config/service-config.json)
# -----------------------------------------------------------------------------
variable "api_service_config_file" {
  type        = string
  default     = ""
  description = "Path to service-config.json relative to this directory; add file to .gitignore if it contains secrets"
}

variable "api_service_config_json" {
  type        = string
  default     = "{}"
  sensitive   = true
  description = "API service config JSON; used when api_service_config_file is not set"
}
