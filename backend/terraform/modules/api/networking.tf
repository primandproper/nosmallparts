# VPC and private service connection – created by the module so prod does not
# require a pre-existing network. Uses var.private_network_name and
# var.private_ip_address_name for the resource names.

resource "google_compute_network" "private_network" {
  name                    = var.private_network_name
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"
}

# Subnet in the primary region so the VPC connector can use the network.
resource "google_compute_subnetwork" "private_subnet" {
  name          = "${var.private_network_name}-${var.gcp_region}"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.gcp_region
  network       = google_compute_network.private_network.id
}

resource "google_compute_global_address" "private_ip_address" {
  name          = var.private_ip_address_name
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.private_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}
