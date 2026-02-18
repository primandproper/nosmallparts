terraform {
  required_version = "1.14.4"

  backend "remote" {
    organization = "nosmallparts"

    workspaces {
      name = "prod-infra"
    }
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.14.1"
    }
  }
}
