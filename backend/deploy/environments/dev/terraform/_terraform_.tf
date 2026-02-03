terraform {
  required_version = "1.10.3"

  backend "remote" {
    organization = "dinnerdonebetter"

    workspaces {
      name = "dev-backend"
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
    archive = {
      source  = "hashicorp/archive"
      version = "2.5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.35.1"
    }
    algolia = {
      source  = "philippe-vandermoere/algolia"
      version = "0.7.0"
    }
  }
}
