# Cloudflare DNS records – only created when cloudflare_zone_id is set
resource "cloudflare_record" "database_record" {
  count   = length(var.cloudflare_zone_id) > 0 ? 1 : 0
  zone_id = var.cloudflare_zone_id
  name    = var.cloudflare_db_record_name
  content = google_sql_database_instance.main.public_ip_address
  type    = "A"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform (${var.environment})"
}

resource "cloudflare_record" "api_media" {
  count   = length(var.cloudflare_zone_id) > 0 ? 1 : 0
  zone_id = var.cloudflare_zone_id
  name    = var.cloudflare_media_record_name
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform (${var.environment})"
}

resource "cloudflare_record" "user_data_storage" {
  count   = length(var.cloudflare_zone_id) > 0 ? 1 : 0
  zone_id = var.cloudflare_zone_id
  name    = var.cloudflare_userdata_record_name
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform (${var.environment})"
}
