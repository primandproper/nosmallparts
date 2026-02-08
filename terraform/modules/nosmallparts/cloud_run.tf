resource "google_vpc_access_connector" "cloud_run" {
  name          = "${var.environment}-cloud-run-connector"
  region        = var.gcp_region
  network       = google_compute_network.private_network.name
  ip_cidr_range = var.vpc_connector_ip_cidr_range
}

resource "google_cloud_run_v2_service" "api" {
  name     = "nosmallparts-api-${var.environment}"
  location = var.gcp_region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }
    vpc_access {
      connector = google_vpc_access_connector.cloud_run.id
      egress    = "PRIVATE_RANGES_ONLY"
    }
    volumes {
      name = "config"
      secret {
        secret = google_secret_manager_secret.api_service_config.secret_id
        items {
          path    = "service-config.json"
          version = "latest"
        }
      }
    }
    containers {
      image = "${var.gcp_region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.main.repository_id}/nosmallparts-api:latest"

      env {
        name  = "PORT"
        value = "8080"
      }
      env {
        name  = "CONFIGURATION_FILEPATH"
        value = "/etc/config/service-config.json"
      }
      env {
        name  = "DATABASE_HOST"
        value = google_sql_database_instance.main.private_ip_address
      }
      env {
        name  = "DATABASE_USERNAME"
        value = local.api_database_username
      }
      env {
        name  = "DATABASE_PASSWORD"
        value = random_password.api_user_database_password.result
      }
      env {
        name  = "DATABASE_NAME"
        value = var.database_name
      }
      env {
        name  = "OAUTH2_TOKEN_ENCRYPTION_KEY"
        value = random_string.oauth2_token_encryption_key.result
      }
      env {
        name  = "JWT_SIGNING_KEY"
        value = base64encode(random_string.jwt_signing_key.result)
      }
      env {
        name  = "SENDGRID_API_TOKEN"
        value = var.sendgrid_api_token
      }
      env {
        name  = "SEGMENT_API_TOKEN"
        value = var.segment_api_token
      }
      env {
        name  = "POSTHOG_API_KEY"
        value = var.posthog_api_key
      }
      env {
        name  = "POSTHOG_PERSONAL_API_KEY"
        value = var.posthog_personal_api_key
      }
      env {
        name  = "GOOGLE_SSO_OAUTH2_CLIENT_ID"
        value = var.google_sso_oauth2_client_id
      }
      env {
        name  = "GOOGLE_SSO_OAUTH2_CLIENT_SECRET"
        value = var.google_sso_oauth2_client_secret
      }
      env {
        name  = "NO_SMALL_PARTS_QUEUES_DATA_CHANGES_TOPIC_NAME"
        value = google_pubsub_topic.data_changes_topic.name
      }
      env {
        name  = "NO_SMALL_PARTS_QUEUES_OUTBOUND_EMAILS_TOPIC_NAME"
        value = google_pubsub_topic.outbound_emails_topic.name
      }
      env {
        name  = "NO_SMALL_PARTS_QUEUES_SEARCH_INDEX_REQUESTS_TOPIC_NAME"
        value = google_pubsub_topic.search_index_requests_topic.name
      }
      env {
        name  = "NO_SMALL_PARTS_QUEUES_USER_DATA_AGGREGATION_TOPIC_NAME"
        value = google_pubsub_topic.user_data_aggregator_topic.name
      }
      env {
        name  = "NO_SMALL_PARTS_QUEUES_WEBHOOK_EXECUTION_REQUESTS_TOPIC_NAME"
        value = google_pubsub_topic.webhook_execution_requests_topic.name
      }

      volume_mounts {
        name       = "config"
        mount_path = "/etc/config"
      }
    }
  }

  depends_on = [google_secret_manager_secret_version.api_service_config]
}

resource "google_cloud_run_v2_service_iam_member" "api_public" {
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

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
      image = "${var.gcp_region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.main.repository_id}/nosmallparts-frontend:latest"

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
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "frontend_public" {
  location = google_cloud_run_v2_service.frontend.location
  name     = google_cloud_run_v2_service.frontend.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
