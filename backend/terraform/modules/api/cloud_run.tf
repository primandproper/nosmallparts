resource "google_vpc_access_connector" "cloud_run" {
  name          = "${var.environment}-cloud-run-connector"
  region        = var.gcp_region
  network       = google_compute_network.private_network.name
  ip_cidr_range = var.vpc_connector_ip_cidr_range
}

resource "google_cloud_run_v2_service" "api" {
  name                = "nosmallparts-api-${var.environment}"
  location            = var.gcp_region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

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
