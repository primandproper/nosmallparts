terraform {
  required_version = "1.14.4"

  backend "remote" {
    organization = "nosmallparts"

    workspaces {
      name = "backend"
    }
  }
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.40.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "6.14.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.35.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.0"
    }
  }
}
