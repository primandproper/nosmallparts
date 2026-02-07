resource "random_string" "oauth2_token_encryption_key" {
  length  = 32
  special = false
}

resource "random_string" "jwt_signing_key" {
  length  = 32
  special = false
}
