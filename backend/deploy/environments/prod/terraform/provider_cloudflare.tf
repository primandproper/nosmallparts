# PROD_CLOUDFLARE_API_KEY and PROD_CLOUDFLARE_ZONE_ID defined in secrets.tf
provider "cloudflare" {
  api_token = var.PROD_CLOUDFLARE_API_KEY
}
