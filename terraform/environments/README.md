# Terraform environments

There is a single environment at the moment: **prod**.

## Deploying prod

```bash
cd prod
terraform init -upgrade
terraform plan -out=tfplan
terraform apply tfplan
```

Copy `terraform.tfvars.example` to `terraform.tfvars` and set variables (or use TFC variables / `-var`). Do not commit `terraform.tfvars`.

The prod root uses the Terraform Cloud workspace **`backend`** (see `prod/backend.tf`). Create that workspace in the `nosmallparts` organization if it doesn’t exist.

## Adding more environments later

When you add staging or another env: copy `prod/` to a new directory (e.g. `staging/`), set a different TFC workspace name in `backend.tf`, and set environment-specific variables (project, `db_instance_name`, bucket names, `vpc_connector_ip_cidr_range`, `pubsub_topic_prefix`, etc.).
