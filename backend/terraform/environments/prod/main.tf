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

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "api" {
  source = "../../modules/api"

  environment   = var.environment
  project_id    = var.project_id
  gcp_region    = var.gcp_region
  gcp_main_zone = var.gcp_main_zone

  private_network_name       = var.private_network_name
  private_ip_address_name    = var.private_ip_address_name
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

  media_bucket_name   = var.media_bucket_name
  userdata_bucket_name = var.userdata_bucket_name
  cors_origins        = var.cors_origins

  api_service_config_json = var.api_service_config_file != "" ? file("${path.module}/${var.api_service_config_file}") : var.api_service_config_json

  sendgrid_api_token          = var.sendgrid_api_token
  segment_api_token           = var.segment_api_token
  posthog_api_key             = var.posthog_api_key
  posthog_personal_api_key    = var.posthog_personal_api_key
  google_sso_oauth2_client_id = var.google_sso_oauth2_client_id
  google_sso_oauth2_client_secret = var.google_sso_oauth2_client_secret
}

output "cloud_run_api_url" {
  value       = module.api.cloud_run_api_url
  description = "Cloud Run API URL; use as frontend_backend_http_url in frontend Terraform"
}

output "database_private_ip" {
  value       = module.api.database_private_ip
  description = "Cloud SQL private IP"
}

output "artifact_registry_repository_id" {
  value       = module.api.artifact_registry_repository_id
  description = "Artifact Registry repo ID; pass to frontend Terraform for image path"
}

output "environment" {
  value       = module.api.environment
  description = "Environment name"
}
