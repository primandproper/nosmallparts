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

# media/userdata CNAMEs removed: buckets renamed to no-small-parts-* to avoid GCS domain verification.
# Assets are served via https://storage.googleapis.com/no-small-parts-media/...
# To restore custom domain URLs, add a Load Balancer + Backend Bucket mapping media.nosmall.parts -> no-small-parts-media.
