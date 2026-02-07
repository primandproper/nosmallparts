output "cloud_run_api_url" {
  value       = google_cloud_run_v2_service.api.uri
  description = "Cloud Run API service URL"
}

output "cloud_run_frontend_url" {
  value       = google_cloud_run_v2_service.frontend.uri
  description = "Cloud Run Frontend service URL"
}

output "database_private_ip" {
  value       = google_sql_database_instance.main.private_ip_address
  description = "Cloud SQL private IP"
}

output "database_public_ip" {
  value       = google_sql_database_instance.main.public_ip_address
  description = "Cloud SQL public IP (if enabled)"
}

output "artifact_registry_repository_id" {
  value       = google_artifact_registry_repository.main.repository_id
  description = "Artifact Registry repository ID for container images"
}

output "environment" {
  value       = var.environment
  description = "Environment name"
}
