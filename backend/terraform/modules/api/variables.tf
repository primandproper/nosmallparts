# -----------------------------------------------------------------------------
# Environment and GCP
# -----------------------------------------------------------------------------
variable "environment" {
  description = "Environment name (e.g. dev, staging, prod). Used for resource naming and labels."
  type        = string
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "gcp_region" {
  description = "GCP region for Cloud Run, Cloud SQL, Pub/Sub, etc."
  type        = string
}

variable "gcp_main_zone" {
  description = "GCP zone (e.g. for any zonal resources)"
  type        = string
  default     = null
}

# -----------------------------------------------------------------------------
# Networking (VPC and reserved range are created by the module)
# -----------------------------------------------------------------------------
variable "private_network_name" {
  description = "Name of the VPC network for private service connection (created by module)"
  type        = string
  default     = "private-network"
}

variable "private_ip_address_name" {
  description = "Name of the global address range for service networking (created by module)"
  type        = string
  default     = "private-ip-address"
}

variable "vpc_connector_ip_cidr_range" {
  description = "CIDR for the Serverless VPC Access connector (must not overlap with other envs)"
  type        = string
  default     = "10.8.0.0/28"
}

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------
variable "database_name" {
  description = "Name of the PostgreSQL database"
  type        = string
  default     = "no-small-parts"
}

variable "db_instance_name" {
  description = "Cloud SQL instance name (must be unique per project)"
  type        = string
}

variable "db_ssl_common_name" {
  description = "Common name for the Cloud SQL client SSL cert"
  type        = string
  default     = null
}

# -----------------------------------------------------------------------------
# Pub/Sub (optional prefix for multiple envs in same project)
# -----------------------------------------------------------------------------
variable "pubsub_topic_prefix" {
  description = "Prefix for Pub/Sub topic names (e.g. 'dev-' for dev-outbound_emails). Empty for single env."
  type        = string
  default     = ""
}

# -----------------------------------------------------------------------------
# Artifact Registry
# -----------------------------------------------------------------------------
variable "artifact_repository_id" {
  description = "Artifact Registry repository ID for container images (API and frontend use same repo)"
  type        = string
  default     = "containers"
}

# -----------------------------------------------------------------------------
# Cloudflare (optional – omit zone_id to skip DNS records)
# -----------------------------------------------------------------------------
variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for DNS records. Omit or set to empty to skip Cloudflare resources."
  type        = string
  default     = ""
}

variable "cloudflare_db_record_name" {
  description = "DNS record name for the database (e.g. db for db.example.com)"
  type        = string
  default     = "db"
}

variable "cloudflare_media_record_name" {
  type        = string
  default     = "media"
  description = "DNS record name for media bucket"
}

variable "cloudflare_userdata_record_name" {
  type        = string
  default     = "userdata"
  description = "DNS record name for userdata bucket"
}

# -----------------------------------------------------------------------------
# Storage buckets (globally unique names)
# -----------------------------------------------------------------------------
variable "media_bucket_name" {
  description = "GCS bucket name for API media (must be globally unique)"
  type        = string
}

variable "userdata_bucket_name" {
  description = "GCS bucket name for user data (must be globally unique)"
  type        = string
}

variable "cors_origins" {
  description = "CORS origins for storage buckets (e.g. [\"https://nosmall.parts\"])"
  type        = list(string)
  default     = ["https://nosmall.parts"]
}

# -----------------------------------------------------------------------------
# API config (JSON string; mounted as secret at /etc/config/service-config.json)
# -----------------------------------------------------------------------------
variable "api_service_config_json" {
  description = "API service config as JSON (mounted in Cloud Run; set in TFC or tfvars)"
  type        = string
  default     = "{}"
  sensitive   = true
}

# -----------------------------------------------------------------------------
# API secrets (passed from environment)
# -----------------------------------------------------------------------------
variable "sendgrid_api_token" {
  type      = string
  default   = ""
  sensitive = true
}

variable "segment_api_token" {
  type      = string
  default   = ""
  sensitive = true
}

variable "posthog_api_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "posthog_personal_api_key" {
  type      = string
  default   = ""
  sensitive = true
}

variable "google_sso_oauth2_client_id" {
  type      = string
  default   = ""
  sensitive = true
}

variable "google_sso_oauth2_client_secret" {
  type      = string
  default   = ""
  sensitive = true
}
