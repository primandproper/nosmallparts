# CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID defined in secrets.tf
provider "cloudflare" {
  api_token = var.CLOUDFLARE_API_TOKEN
}
