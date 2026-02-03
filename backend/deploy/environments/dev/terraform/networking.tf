data "google_compute_network" "private_network" {
  provider = google
  name     = "private-network"
}

data "google_compute_global_address" "private_ip_address" {
  provider = google
  name     = "private-ip-address"
}

resource "google_service_networking_connection" "private_vpc_connection" {
  provider = google

  network                 = data.google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [data.google_compute_global_address.private_ip_address.name]
}

data "google_certificate_manager_certificates" "dev" {}



