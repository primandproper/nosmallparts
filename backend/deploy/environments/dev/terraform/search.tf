resource "algolia_index" "users_index" {
  name = "users"

  searchable_attributes = [
    "username",
  ]
}
