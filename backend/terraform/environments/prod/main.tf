# -----------------------------------------------------------------------------
# Prod environment – backend (API) only. Run from this directory.
# -----------------------------------------------------------------------------
# terraform init && terraform plan && terraform apply

terraform {
  required_version = ">= 1.10"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.14"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = ">= 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0"
    }
  }
}

provider "google" {
  project     = var.project_id
  region      = var.gcp_region
  zone        = var.gcp_main_zone
  credentials = var.google_cloud_credentials
}

# Token is read from CLOUDFLARE_API_TOKEN env var (set in TFC workspace env vars).
provider "cloudflare" {}

module "api" {
  source = "../../modules/api"

  environment   = var.environment
  project_id    = var.project_id
  gcp_region    = var.gcp_region
  gcp_main_zone = var.gcp_main_zone

  private_network_name        = var.private_network_name
  private_ip_address_name     = var.private_ip_address_name
  vpc_connector_ip_cidr_range = var.vpc_connector_ip_cidr_range

  database_name      = var.database_name
  db_instance_name   = var.db_instance_name
  db_ssl_common_name = var.db_ssl_common_name

  pubsub_topic_prefix    = var.pubsub_topic_prefix
  artifact_repository_id = var.artifact_repository_id

  cloudflare_zone_id              = var.cloudflare_zone_id
  cloudflare_db_record_name       = var.cloudflare_db_record_name
  cloudflare_media_record_name    = var.cloudflare_media_record_name
  cloudflare_userdata_record_name = var.cloudflare_userdata_record_name

  # Default to project+env so bucket names are globally unique when not set in TFC
  media_bucket_name    = coalesce(var.media_bucket_name, "${var.project_id}-${var.environment}-media")
  userdata_bucket_name = coalesce(var.userdata_bucket_name, "${var.project_id}-${var.environment}-userdata")
  cors_origins         = var.cors_origins

  api_service_config_json = var.api_service_config_file != "" ? file("${path.module}/${var.api_service_config_file}") : var.api_service_config_json
}

output "cloud_run_api_url" {
  value       = module.api.cloud_run_api_url
  description = "Cloud Run API URL; use as frontend_backend_http_url in frontend Terraform"
}

output "database_private_ip" {
  value       = module.api.database_private_ip
  description = "Cloud SQL private IP"
}

output "environment" {
  value       = module.api.environment
  description = "Environment name"
}
