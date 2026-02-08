# Topic names can be prefixed (e.g. pubsub_topic_prefix = "dev-") when deploying multiple envs in the same project.
locals {
  pubsub_prefix = var.pubsub_topic_prefix
}

# API publishes change events here; consumers (e.g. Cloud Functions) can subscribe (not created by this module).
resource "google_pubsub_topic" "data_changes_topic" {
  name = "${local.pubsub_prefix}data_changes"
}

resource "google_pubsub_topic" "outbound_emails_topic" {
  name = "${local.pubsub_prefix}outbound_emails"
}

resource "google_pubsub_topic" "outbound_emails_deadletter_topic" {
  name = "${local.pubsub_prefix}outbound_emails_deadletter"
}

resource "google_pubsub_subscription" "outbound_emails_topic" {
  name  = google_pubsub_topic.outbound_emails_topic.name
  topic = google_pubsub_topic.outbound_emails_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.outbound_emails_deadletter_topic.id
    max_delivery_attempts = 5
  }
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
  enable_exactly_once_delivery = true
}

resource "google_pubsub_topic" "search_index_requests_topic" {
  name = "${local.pubsub_prefix}search_index_requests"
}

resource "google_pubsub_topic" "search_index_requests_deadletter_topic" {
  name = "${local.pubsub_prefix}search_index_requests_deadletter"
}

resource "google_pubsub_subscription" "search_index_requests_topic" {
  name  = google_pubsub_topic.search_index_requests_topic.name
  topic = google_pubsub_topic.search_index_requests_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.search_index_requests_deadletter_topic.id
    max_delivery_attempts = 5
  }
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
  enable_exactly_once_delivery = true
}

resource "google_pubsub_topic" "user_data_aggregator_topic" {
  name = "${local.pubsub_prefix}user_data_aggregation_requests"
}

resource "google_pubsub_topic" "user_data_aggregator_deadletter_topic" {
  name = "${local.pubsub_prefix}user_data_aggregator_deadletter"
}

resource "google_pubsub_subscription" "user_data_aggregator_topic" {
  name  = google_pubsub_topic.user_data_aggregator_topic.name
  topic = google_pubsub_topic.user_data_aggregator_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.user_data_aggregator_deadletter_topic.id
    max_delivery_attempts = 5
  }
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
  enable_exactly_once_delivery = true
}

resource "google_pubsub_topic" "webhook_execution_requests_topic" {
  name = "${local.pubsub_prefix}webhook_execution_requests"
}

resource "google_pubsub_topic" "webhook_execution_requests_deadletter_topic" {
  name = "${local.pubsub_prefix}webhook_execution_requests_deadletter"
}

resource "google_pubsub_subscription" "webhook_execution_requests_topic" {
  name  = google_pubsub_topic.webhook_execution_requests_topic.name
  topic = google_pubsub_topic.webhook_execution_requests_topic.id

  message_retention_duration = "604800s"
  retain_acked_messages      = false
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.webhook_execution_requests_deadletter_topic.id
    max_delivery_attempts = 5
  }
  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }
  enable_exactly_once_delivery = true
}
