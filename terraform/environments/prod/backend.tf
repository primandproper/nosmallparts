# Local backend – init works without Terraform Cloud.
# To use Terraform Cloud instead: replace this with backend "remote" and run terraform init -migrate-state.
terraform {
  backend "local" {
    path = "prod.tfstate"
  }
}
