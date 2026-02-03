resource "google_storage_bucket" "user_data_storage" {
  provider                    = google
  name                        = "userdata.dinnerdonebetter.dev"
  location                    = "US"
  uniform_bucket_level_access = false
  force_destroy               = true

  versioning {
    enabled = true
  }

  website {
    main_page_suffix = "index.html"
    not_found_page   = "index.html"
  }

  cors {
    origin          = ["https://dinnerdonebetter.dev]"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_storage_bucket_access_control" "user_data_public_rule" {
  bucket = google_storage_bucket.user_data_storage.name
  role   = "READER"
  entity = "allUsers"
}

resource "google_storage_bucket_iam_policy" "user_data_policy" {
  bucket      = google_storage_bucket.user_data_storage.name
  policy_data = data.google_iam_policy.public_policy.policy_data
}

resource "cloudflare_record" "user_data_storage" {
  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "userdata"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}
