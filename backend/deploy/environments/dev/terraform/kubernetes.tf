locals {
  k8s_namespace = "dev"
}

provider "kubernetes" {
  config_path    = "./terraform_kubeconfig"
  config_context = "${local.k8s_namespace}_context"
}

data "google_container_cluster" "dev_cluster" {
  name     = "dev"
  location = local.gcp_region
}

# Kubernetes secrets

resource "kubernetes_secret" "cloudflare_api_key" {
  metadata {
    name      = "cloudflare-api-key"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.dev_cluster]

  data = {
    "token" = var.CLOUDFLARE_API_TOKEN
  }
}

resource "kubernetes_config_map_v1" "pubsub_topics" {
  metadata {
    name      = "pubsub-topic-names"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.dev_cluster]

  data = {
    data_changes               = google_pubsub_topic.data_changes_topic.name
    outbound_emails            = google_pubsub_topic.outbound_emails_topic.name
    search_index_requests      = google_pubsub_topic.search_index_requests_topic.name
    user_data_aggregator       = google_pubsub_topic.user_data_aggregator_topic.name
    webhook_execution_requests = google_pubsub_topic.webhook_execution_requests_topic.name
  }
}

resource "kubernetes_secret" "api_service_config" {
  metadata {
    name      = "api-service-config"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.dev_cluster]

  data = {
    OAUTH2_TOKEN_ENCRYPTION_KEY     = random_string.oauth2_token_encryption_key.result
    JWT_SIGNING_KEY                 = base64encode(random_string.jwt_signing_key.result)
    DATABASE_HOST                   = google_sql_database_instance.dev.private_ip_address
    DATABASE_USERNAME               = local.api_database_username
    DATABASE_PASSWORD               = random_password.api_user_database_password.result
    SENDGRID_API_TOKEN              = var.SENDGRID_API_TOKEN
    SEGMENT_API_TOKEN               = var.SEGMENT_API_TOKEN
    POSTHOG_API_KEY                 = var.POSTHOG_API_KEY
    POSTHOG_PERSONAL_API_KEY        = var.POSTHOG_PERSONAL_API_KEY
    ALGOLIA_APPLICATION_ID          = var.ALGOLIA_APPLICATION_ID
    ALGOLIA_API_KEY                 = var.ALGOLIA_API_KEY
    GOOGLE_SSO_OAUTH2_CLIENT_ID     = var.GOOGLE_SSO_OAUTH2_CLIENT_ID
    GOOGLE_SSO_OAUTH2_CLIENT_SECRET = var.GOOGLE_SSO_OAUTH2_CLIENT_SECRET
  }
}

# this is the sort of resource that should probably ideally live in the infra folder, but it's here for now
# because I haven't yet wanted to fuss with figuring out how to manage the code that creates the cluster
# alongside the code that creates resources in that cluster.
resource "kubernetes_secret" "grafana_cloud_creds" {
  metadata {
    name      = "grafana-cloud-creds"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.dev_cluster]

  data = {
    GRAFANA_CLOUD_PROMETHEUS_USERNAME = var.GRAFANA_CLOUD_PROMETHEUS_USERNAME
    GRAFANA_CLOUD_PROMETHEUS_PASSWORD = var.GRAFANA_CLOUD_PROMETHEUS_PASSWORD
    GRAFANA_CLOUD_LOKI_USERNAME       = var.GRAFANA_CLOUD_LOKI_USERNAME
    GRAFANA_CLOUD_LOKI_PASSWORD       = var.GRAFANA_CLOUD_LOKI_PASSWORD
    GRAFANA_CLOUD_TEMPO_USERNAME      = var.GRAFANA_CLOUD_TEMPO_USERNAME
    GRAFANA_CLOUD_TEMPO_PASSWORD      = var.GRAFANA_CLOUD_TEMPO_PASSWORD
  }
}
