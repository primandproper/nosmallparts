resource "google_project_iam_custom_role" "prod_cluster_role" {
  role_id     = "prod_cluster_role"
  title       = "Prod cluster role"
  description = "An IAM role for the prod cluster"
  permissions = [
    "secretmanager.versions.access",
    "cloudsql.instances.connect",
    "cloudsql.instances.get",
    "pubsub.topics.list",
    "pubsub.topics.publish",
    "cloudtrace.traces.patch",
    "logging.buckets.create",
    "logging.buckets.write",
    "logging.buckets.list",
    "logging.buckets.get",
    "storage.objects.list",
    "storage.objects.get",
    "storage.objects.update",
    "storage.objects.create",
    "storage.objects.delete",
  ]
}

resource "google_service_account" "prod_cluster_service_account" {
  account_id   = "prod-cluster"
  display_name = "Prod Cluster Service Account"
}

resource "google_project_iam_member" "prod_cluster" {
  project = local.project_id
  role    = google_project_iam_custom_role.prod_cluster_role.id
  member  = format("serviceAccount:%s", google_service_account.prod_cluster_service_account.email)
}

resource "google_container_cluster" "primary" {
  name     = local.environment
  location = local.gcp_region

  remove_default_node_pool = true
  initial_node_count       = 1
  deletion_protection      = false

  network    = google_compute_network.private_network.name
  subnetwork  = google_compute_subnetwork.primary.name

  secret_manager_config {
    enabled = true
  }

  workload_identity_config {
    workload_pool = "${local.project_id}.svc.id.goog"
  }
}

resource "google_container_node_pool" "primary_nodes" {
  location   = local.gcp_region
  name       = google_container_cluster.primary.name
  cluster    = google_container_cluster.primary.name
  node_count = 1

  node_locations = [
    local.gcp_main_zone,
  ]

  autoscaling {
    total_max_node_count = 1
    location_policy      = "BALANCED"
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  network_config {}

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    labels = {
      env = local.project_id
    }

    preemptible  = true
    machine_type = "e2-standard-2"
    tags         = ["gke-node", local.environment]
    metadata = {
      disable-legacy-endpoints = "true"
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }
}

resource "google_service_account" "workload_identity_sa" {
  account_id   = "workload-identity-sa"
  display_name = "Workload Identity Service Account"
  description  = "Service account for Kubernetes workloads to access GCP resources"
}

resource "google_project_iam_member" "workload_identity_secret_accessor" {
  project = local.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.workload_identity_sa.email}"
}

resource "google_service_account_iam_member" "workload_identity_binding" {
  service_account_id = google_service_account.workload_identity_sa.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${local.project_id}.svc.id.goog[prod/api-service-account]"

  depends_on = [google_container_cluster.primary]
}
