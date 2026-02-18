resource "google_cloud_run_v2_service" "frontend" {
  name     = "nosmallparts-frontend-${var.environment}"
  location = var.gcp_region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 5
    }
    containers {
      image = "${var.gcp_region}-docker.pkg.dev/${var.project_id}/${var.artifact_repository_id}/nosmallparts-frontend:latest"

      env {
        name  = "BACKEND_HTTP_URL"
        value = var.frontend_backend_http_url
      }
      env {
        name  = "GRPC_BACKEND_URL"
        value = var.frontend_grpc_backend_url
      }
      env {
        name  = "OAUTH2_CLIENT_ID"
        value = var.frontend_oauth2_client_id
      }
      env {
        name  = "OAUTH2_CLIENT_SECRET"
        value = var.frontend_oauth2_client_secret
      }
      env {
        name  = "NEXT_PUBLIC_BACKEND_WS_URL"
        value = var.frontend_next_public_backend_ws_url
      }
      env {
        name  = "ADSENSE_PUBLISHER_ID"
        value = var.ADSENSE_PUBLISHER_ID
      }
      env {
        name  = "ADSENSE_SLOT_ID"
        value = var.ADSENSE_SLOT_ID
      }
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  location = google_cloud_run_v2_service.frontend.location
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
