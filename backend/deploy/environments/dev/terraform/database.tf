locals {
  database_name = "dinner-done-better"
}

resource "google_sql_database_instance" "dev" {
  name                = "dev"
  database_version    = "POSTGRES_17"
  region              = local.gcp_region
  deletion_protection = false

  depends_on = [google_service_networking_connection.private_vpc_connection]

  settings {
    tier                  = "db-f1-micro"
    disk_size             = 20
    disk_autoresize       = true
    disk_autoresize_limit = 50
    edition               = "ENTERPRISE"

    insights_config {
      query_insights_enabled  = true
      query_string_length     = 4096
      record_application_tags = false
      record_client_address   = false
    }

    maintenance_window {
      day          = 7
      hour         = 3
      update_track = "stable"
    }

    ip_configuration {
      ssl_mode                                      = "ENCRYPTED_ONLY"
      ipv4_enabled                                  = true
      private_network                               = data.google_compute_network.private_network.id
      enable_private_path_for_google_cloud_services = true
    }

    password_validation_policy {
      min_length     = 30 # [0, 30]
      reuse_interval = 1
      # A combination of lowercase, uppercase, numeric, and non-alphanumeric characters. The only other option is "COMPLEXITY_UNSPECIFIED"
      complexity                  = "COMPLEXITY_DEFAULT"
      disallow_username_substring = true
      password_change_interval    = "1s"
      enable_password_policy      = true
    }
  }
}

resource "google_sql_ssl_cert" "client_cert" {
  common_name = "dinner-done-better-dev"
  instance    = google_sql_database_instance.dev.name
}

resource "google_sql_database" "api_database" {
  name     = local.database_name
  instance = google_sql_database_instance.dev.name
}

resource "cloudflare_record" "database_record" {
  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "db"
  content = google_sql_database_instance.dev.public_ip_address
  type    = "A"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}
