terraform {
  required_version = ">= 1.10"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.14"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = ">= 4.40"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0"
    }
  }
}
