variable "CLOUDFLARE_API_TOKEN" {}
variable "CLOUDFLARE_ZONE_ID" {}

provider "cloudflare" {
  api_token = var.CLOUDFLARE_API_TOKEN
}
