# Terraform Cloud remote backend – state stored in TFC.
# Use a separate workspace from backend (e.g. "frontend" in org nosmallparts).
terraform {
  backend "remote" {
    organization = "nosmallparts"

    workspaces {
      name = "frontend"
    }
  }
}
