# API server oauth2 token encryption key
resource "random_string" "oauth2_token_encryption_key" {
  length  = 32
  special = false
}

# JWT Signing key
resource "random_string" "jwt_signing_key" {
  length  = 32
  special = false
}

### External API services ###

# Sendgrid token
variable "SENDGRID_API_TOKEN" {}

# Segment API token
variable "SEGMENT_API_TOKEN" {}

# PostHog API token
variable "POSTHOG_API_KEY" {}

# PostHog personal API key
variable "POSTHOG_PERSONAL_API_KEY" {}

# Grafana Cloud API keys
# NOTE: the passwords are all effectively the same, but they maybe won't be one day? Who knows.
variable "GRAFANA_CLOUD_PROMETHEUS_USERNAME" {}
variable "GRAFANA_CLOUD_PROMETHEUS_PASSWORD" {}
variable "GRAFANA_CLOUD_LOKI_USERNAME" {}
variable "GRAFANA_CLOUD_LOKI_PASSWORD" {}
variable "GRAFANA_CLOUD_TEMPO_USERNAME" {}
variable "GRAFANA_CLOUD_TEMPO_PASSWORD" {}
