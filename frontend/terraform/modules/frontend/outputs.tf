output "cloud_run_frontend_url" {
  value       = google_cloud_run_v2_service.frontend.uri
  description = "Cloud Run Frontend service URL"
}

output "environment" {
  value       = var.environment
  description = "Environment name"
}
