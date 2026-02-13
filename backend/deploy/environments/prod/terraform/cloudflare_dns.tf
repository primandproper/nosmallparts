# Optional: Cloudflare DNS records for database and storage. Set PROD_PROD_CLOUDFLARE_ZONE_ID and PROD_CLOUDFLARE_API_KEY to enable.
resource "cloudflare_record" "database_record" {
  count = var.PROD_CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.PROD_CLOUDFLARE_ZONE_ID
  name    = "db"
  content = google_sql_database_instance.prod.public_ip_address
  type    = "A"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}

resource "cloudflare_record" "api_media" {
  count = var.PROD_CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.PROD_CLOUDFLARE_ZONE_ID
  name    = "media"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}

resource "cloudflare_record" "user_data_storage" {
  count = var.PROD_CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.PROD_CLOUDFLARE_ZONE_ID
  name    = "userdata"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}
