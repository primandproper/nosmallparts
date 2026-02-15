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

# OAuth2 client credentials for apps (set in Terraform Cloud; created via playground)
variable "ADMIN_APP_OAUTH2_CLIENT_ID" {
  default     = ""
  description = "OAuth2 client ID for admin webapp (admin_app)"
}
variable "ADMIN_APP_OAUTH2_CLIENT_SECRET" {
  default     = ""
  sensitive   = true
  description = "OAuth2 client secret for admin webapp (admin_app)"
}
variable "WEB_APP_OAUTH2_CLIENT_ID" {
  default     = ""
  description = "OAuth2 client ID for frontend webapp (web_app)"
}
variable "WEB_APP_OAUTH2_CLIENT_SECRET" {
  default     = ""
  sensitive   = true
  description = "OAuth2 client secret for frontend webapp (web_app)"
}

# LiveKit (optional - video rooms use mock mode when unset)
variable "LIVEKIT_API_KEY" {
  default     = ""
  sensitive   = true
  description = "LiveKit Cloud API key (from dashboard). Leave empty for mock mode."
}
variable "LIVEKIT_API_SECRET" {
  default     = ""
  sensitive   = true
  description = "LiveKit Cloud API secret. Leave empty for mock mode."
}
variable "LIVEKIT_URL" {
  default     = ""
  description = "LiveKit server URL (e.g. wss://your-project.livekit.cloud). Leave empty for mock mode."
}
