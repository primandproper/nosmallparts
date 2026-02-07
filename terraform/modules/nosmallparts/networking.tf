data "google_compute_network" "private_network" {
  name = var.private_network_name
}

data "google_compute_global_address" "private_ip_address" {
  name = var.private_ip_address_name
}

resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = data.google_compute_network.private_network.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [data.google_compute_global_address.private_ip_address.name]
}
