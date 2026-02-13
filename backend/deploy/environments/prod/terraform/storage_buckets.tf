data "google_iam_policy" "public_policy" {
  binding {
    role = "roles/storage.objectViewer"
    members = [
      "allUsers",
    ]
  }
}

# Non-domain names avoid GCS domain ownership verification (must use Load Balancer for media.nosmall.parts)
resource "google_storage_bucket" "api_media" {
  provider                    = google
  name                        = "no-small-parts-media"
  location                    = "US"
  uniform_bucket_level_access = false
  force_destroy               = true

  versioning {
    enabled = true
  }

  cors {
    origin          = ["https://nosmall.parts", "https://www.nosmall.parts", "https://app.nosmall.parts", "https://admin.nosmall.parts"]
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

resource "google_storage_bucket_iam_policy" "api_media_policy" {
  bucket      = google_storage_bucket.api_media.name
  policy_data = data.google_iam_policy.public_policy.policy_data
}

resource "google_storage_bucket" "user_data_storage" {
  provider                    = google
  name                        = "no-small-parts-userdata"
  location                    = "US"
  uniform_bucket_level_access = false
  force_destroy               = true

  versioning {
    enabled = true
  }

  cors {
    origin          = ["https://nosmall.parts", "https://www.nosmall.parts", "https://app.nosmall.parts", "https://admin.nosmall.parts"]
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

resource "google_storage_bucket_iam_policy" "user_data_policy" {
  bucket      = google_storage_bucket.user_data_storage.name
  policy_data = data.google_iam_policy.public_policy.policy_data
}
