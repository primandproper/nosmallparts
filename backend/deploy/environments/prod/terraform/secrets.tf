resource "random_string" "oauth2_token_encryption_key" {
  length  = 32
  special = false
}

resource "random_string" "jwt_signing_key" {
  length  = 32
  special = false
}

variable "SENDGRID_API_TOKEN" {
  default = ""
}
variable "SEGMENT_API_TOKEN" {
  default = ""
}
variable "POSTHOG_API_KEY" {
  default = ""
}
variable "POSTHOG_PERSONAL_API_KEY" {
  default = ""
  sensitive = true
}
variable "ALGOLIA_APPLICATION_ID" {
  default = ""
}
variable "ALGOLIA_API_KEY" {
  default = ""
  sensitive = true
}

variable "GRAFANA_CLOUD_PROMETHEUS_USERNAME" {
  default = ""
}
variable "GRAFANA_CLOUD_PROMETHEUS_PASSWORD" {
  default = ""
  sensitive = true
}
variable "GRAFANA_CLOUD_LOKI_USERNAME" {
  default = ""
}
variable "GRAFANA_CLOUD_LOKI_PASSWORD" {
  default = ""
  sensitive = true
}
variable "GRAFANA_CLOUD_TEMPO_USERNAME" {
  default = ""
}
variable "GRAFANA_CLOUD_TEMPO_PASSWORD" {
  default = ""
  sensitive = true
}

variable "PROD_CLOUDFLARE_API_KEY" {
  default = ""
  sensitive = true
}
variable "PROD_CLOUDFLARE_ZONE_ID" {
  default = ""
}
