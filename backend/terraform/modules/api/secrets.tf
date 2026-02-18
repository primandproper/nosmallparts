data "google_project" "project" {
  project_id = var.project_id
}

# API service config – stored in Secret Manager and mounted in Cloud Run at /etc/config.
resource "google_secret_manager_secret" "api_service_config" {
  secret_id = "${var.environment}-api-service-config"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "api_service_config" {
  secret      = google_secret_manager_secret.api_service_config.id
  secret_data = var.api_service_config_json
}

# Allow Cloud Run default compute SA to read the config secret.
resource "google_secret_manager_secret_iam_member" "api_service_config_cloud_run" {
  secret_id = google_secret_manager_secret.api_service_config.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "random_string" "oauth2_token_encryption_key" {
  length  = 32
  special = false
}

resource "random_string" "jwt_signing_key" {
  length  = 32
  special = false
}
