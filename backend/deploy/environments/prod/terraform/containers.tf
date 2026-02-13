resource "google_artifact_registry_repository" "prod_repository" {
  location      = local.gcp_region
  repository_id = "containers"
  description   = "Container images for the prod environment"
  format        = "DOCKER"
}
