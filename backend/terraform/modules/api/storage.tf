data "google_iam_policy" "public_policy" {
  binding {
    role    = "roles/storage.objectViewer"
    members = ["allUsers"]
  }
}

resource "google_storage_bucket" "api_media" {
  name                        = var.media_bucket_name
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
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    action { type = "Delete" }
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

resource "google_storage_bucket" "user_data_storage" {
  name                        = var.userdata_bucket_name
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
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    condition { age = 30 }
    action { type = "Delete" }
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
