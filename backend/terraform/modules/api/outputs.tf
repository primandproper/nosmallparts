output "cloud_run_api_url" {
  value       = google_cloud_run_v2_service.api.uri
  description = "Cloud Run API service URL (use as frontend_backend_http_url for frontend Terraform)"
}

output "database_private_ip" {
  value       = google_sql_database_instance.main.private_ip_address
  description = "Cloud SQL private IP"
}

output "artifact_registry_repository_id" {
  value       = google_artifact_registry_repository.main.repository_id
  description = "Artifact Registry repository ID (API and frontend images; pass to frontend Terraform)"
}

output "environment" {
  value       = var.environment
  description = "Environment name"
}
