resource "google_artifact_registry_repository" "main" {
  location      = var.gcp_region
  repository_id = var.artifact_repository_id
  description   = "Container images for ${var.environment}"
  format        = "DOCKER"
}
