resource "google_compute_network" "private_network" {
  provider                = google
  name                    = "private-network"
  auto_create_subnetworks = false 
}

resource "google_compute_subnetwork" "custom_subnet" {
  name          = "private-subnet-1"
  ip_cidr_range = "10.0.1.0/24"
  region        = "us-central1"
  network       = google_compute_network.private_network.id
}

resource "google_compute_global_address" "private_ip_address" {
  name          = "private-ip-address"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 24
  address       = "10.160.0.0" 
  network       = google_compute_network.private_network.id
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_address.name]
}
