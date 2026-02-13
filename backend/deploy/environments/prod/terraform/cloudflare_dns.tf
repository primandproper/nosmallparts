# Optional: Cloudflare DNS records for database and storage. Set CLOUDFLARE_ZONE_ID and CLOUDFLARE_API_TOKEN to enable.
resource "cloudflare_record" "database_record" {
  count = var.CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "db"
  content = google_sql_database_instance.prod.public_ip_address
  type    = "A"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}

resource "cloudflare_record" "api_media" {
  count = var.CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "media"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}

resource "cloudflare_record" "user_data_storage" {
  count = var.CLOUDFLARE_ZONE_ID != "" ? 1 : 0

  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "userdata"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}
