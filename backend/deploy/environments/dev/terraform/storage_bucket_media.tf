resource "google_storage_bucket" "api_media" {
  provider                    = google
  name                        = "media.dinnerdonebetter.dev"
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
    action {
      type = "Delete"
    }
    condition {
      age                = 7
      num_newer_versions = 3
    }
  }
}

resource "google_storage_bucket_access_control" "api_media_public_rule" {
  bucket = google_storage_bucket.api_media.name
  role   = "READER"
  entity = "allUsers"
}

resource "google_storage_bucket_iam_policy" "api_media_policy" {
  bucket      = google_storage_bucket.api_media.name
  policy_data = data.google_iam_policy.public_policy.policy_data
}

resource "cloudflare_record" "api_media" {
  zone_id = var.CLOUDFLARE_ZONE_ID
  name    = "media"
  content = "c.storage.googleapis.com"
  type    = "CNAME"
  proxied = true
  ttl     = 1
  comment = "Managed by Terraform"
}
