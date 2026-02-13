# Terraform Cloud remote backend – state stored in TFC.
# For CI: set secret NO_SMALL_PARTS_TERRAFORM_CLOUD_API_TOKEN. For local: terraform login or TF_TOKEN_app_terraform_io.
terraform {
  backend "remote" {
    organization = "nosmallparts"

    workspaces {
      name = "backend"
    }
  }
}
