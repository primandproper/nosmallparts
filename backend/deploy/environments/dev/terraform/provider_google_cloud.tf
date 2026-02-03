variable "GOOGLE_CLOUD_CREDENTIALS" {}

locals {
  project_id    = "dinner-done-better-dev"
  gcp_region    = "us-central1"
  gcp_main_zone = "us-central1-a"
}

provider "google" {
  project     = local.project_id
  region      = local.gcp_region
  zone        = local.gcp_main_zone
  credentials = var.GOOGLE_CLOUD_CREDENTIALS
}

variable "GOOGLE_SSO_OAUTH2_CLIENT_ID" {}
variable "GOOGLE_SSO_OAUTH2_CLIENT_SECRET" {}
