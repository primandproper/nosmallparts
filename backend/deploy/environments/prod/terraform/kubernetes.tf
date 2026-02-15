locals {
  k8s_namespace = "prod"
}

provider "kubernetes" {
  config_path    = "./terraform_kubeconfig"
  config_context = "prod_context"
}

data "google_container_cluster" "prod_cluster" {
  name     = "prod"
  location = local.gcp_region
}

resource "kubernetes_namespace" "prod" {
  metadata {
    name = local.k8s_namespace

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [data.google_container_cluster.prod_cluster]
}

resource "kubernetes_secret" "cloudflare_api_key" {
  count = var.PROD_CLOUDFLARE_API_KEY != "" ? 1 : 0

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

  depends_on = [kubernetes_namespace.prod]

  data = {
    "token" = var.PROD_CLOUDFLARE_API_KEY
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

  depends_on = [kubernetes_namespace.prod]

  data = {
    data_changes               = google_pubsub_topic.data_changes_topic.name
    outbound_emails            = google_pubsub_topic.outbound_emails_topic.name
    search_index_requests      = google_pubsub_topic.search_index_requests_topic.name
    user_data_aggregator       = google_pubsub_topic.user_data_aggregator_topic.name
    webhook_execution_requests = google_pubsub_topic.webhook_execution_requests_topic.name
  }
}

resource "kubernetes_secret" "api_service_secrets" {
  metadata {
    name      = "api-service-secrets"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace.prod]

  data = {
    DATABASE_HOST                   = google_sql_database_instance.prod.private_ip_address
    DATABASE_PASSWORD               = random_password.api_user_database_password.result
    OAUTH2_TOKEN_ENCRYPTION_KEY     = random_string.oauth2_token_encryption_key.result
    JWT_SIGNING_KEY                  = base64encode(random_string.jwt_signing_key.result)
    SEGMENT_API_TOKEN                = var.SEGMENT_API_TOKEN
    ALGOLIA_API_KEY                  = var.ALGOLIA_API_KEY
    ALGOLIA_APPLICATION_ID           = var.ALGOLIA_APPLICATION_ID
    SENDGRID_API_TOKEN               = var.SENDGRID_API_TOKEN
    POSTHOG_PERSONAL_API_KEY         = var.POSTHOG_PERSONAL_API_KEY
    POSTHOG_API_KEY                   = var.POSTHOG_API_KEY
    GOOGLE_SSO_OAUTH2_CLIENT_ID      = var.GOOGLE_SSO_OAUTH2_CLIENT_ID
    GOOGLE_SSO_OAUTH2_CLIENT_SECRET   = var.GOOGLE_SSO_OAUTH2_CLIENT_SECRET
    LIVEKIT_API_KEY                   = var.LIVEKIT_API_KEY
    LIVEKIT_API_SECRET                = var.LIVEKIT_API_SECRET
    LIVEKIT_URL                       = var.LIVEKIT_URL
  }
}

resource "kubernetes_secret" "admin_webapp_config" {
  metadata {
    name      = "admin-webapp-config"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace.prod]

  data = {
    COOKIE_NAME        = "no_small_parts_session"
    COOKIE_HASH_KEY    = random_string.oauth2_token_encryption_key.result
    COOKIE_BLOCK_KEY   = random_string.jwt_signing_key.result
    COOKIE_DOMAIN      = ".nosmall.parts"
    COOKIE_SECURE_ONLY = "true"
    API_SERVER_HTTP_URL  = "https://api.nosmall.parts"
    # In-cluster gRPC address; admin webapp uses insecure creds, so use plain TCP to API service, not HTTPS via ingress
    API_SERVER_GRPC_URL  = "no-small-parts-api-svc:8001"
    OAUTH2_CLIENT_ID     = var.ADMIN_APP_OAUTH2_CLIENT_ID
    OAUTH2_CLIENT_SECRET = var.ADMIN_APP_OAUTH2_CLIENT_SECRET
  }
}

resource "kubernetes_secret" "frontend_config" {
  metadata {
    name      = "frontend-config"
    namespace = local.k8s_namespace

    annotations = {
      (local.managed_by_label) = "terraform"
    }

    labels = {
      (local.managed_by_label) = "terraform"
    }
  }

  depends_on = [kubernetes_namespace.prod]

  data = {
    OAUTH2_CLIENT_ID     = var.WEB_APP_OAUTH2_CLIENT_ID
    OAUTH2_CLIENT_SECRET = var.WEB_APP_OAUTH2_CLIENT_SECRET
    COOKIE_DOMAIN        = ".nosmall.parts"
  }
}

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

  depends_on = [kubernetes_namespace.prod]

  data = {
    GRAFANA_CLOUD_PROMETHEUS_USERNAME = var.GRAFANA_CLOUD_PROMETHEUS_USERNAME
    GRAFANA_CLOUD_PROMETHEUS_PASSWORD = var.GRAFANA_CLOUD_PROMETHEUS_PASSWORD
    GRAFANA_CLOUD_LOKI_USERNAME       = var.GRAFANA_CLOUD_LOKI_USERNAME
    GRAFANA_CLOUD_LOKI_PASSWORD      = var.GRAFANA_CLOUD_LOKI_PASSWORD
    GRAFANA_CLOUD_TEMPO_USERNAME      = var.GRAFANA_CLOUD_TEMPO_USERNAME
    GRAFANA_CLOUD_TEMPO_PASSWORD      = var.GRAFANA_CLOUD_TEMPO_PASSWORD
  }
}
