# Custom mode = subnets only where we need them (us-central1), not in every region.
# If network already exists in auto mode, run first:
#   gcloud compute networks update private-network --switch-to-custom-subnet-mode --project=no-small-parts
# Then delete unwanted regional subnets in the console or via gcloud.
resource "google_compute_network" "private_network" {
  provider = google

  name                    = "private-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "primary" {
  provider = google

  name          = "private-network-${local.gcp_region}"
  ip_cidr_range = "10.0.0.0/20"
  region        = local.gcp_region
  network       = google_compute_network.private_network.id
}

resource "google_compute_address" "static_ip" {
  name = "dev"
  labels = {
    (local.managed_by_label) = "terraform"
  }
}

# 192.168.0.0/16 is disjoint from our subnet (10.0.0.0/20) and avoids 10.128.0.0/9 (GCP auto subnets).
resource "google_compute_global_address" "private_ip_address" {
  provider = google

  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  address       = "192.168.0.0"
  prefix_length = 16
  network       = google_compute_network.private_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  provider = google

  network                 = google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}
