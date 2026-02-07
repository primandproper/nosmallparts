# nosmallparts Terraform module

Reusable module that deploys the No Small Parts stack for a single environment:

- **GCP:** Cloud SQL, Pub/Sub, Artifact Registry, GCS, Serverless VPC Access, Cloud Run (API + frontend)
- **Cloudflare:** DNS records (optional; set `cloudflare_zone_id`)

## Usage

From an environment root (e.g. `environments/dev`):

```hcl
module "nosmallparts" {
  source = "../../modules/nosmallparts"

  environment  = "dev"
  project_id   = "my-project"
  gcp_region   = "us-central1"

  db_instance_name = "dev"

  media_bucket_name   = "my-project-dev-media"
  userdata_bucket_name = "my-project-dev-userdata"

  frontend_backend_http_url = "https://nosmallparts-api-dev-xxx.run.app"
  frontend_grpc_backend_url = "nosmallparts-api-dev-xxx.run.app:443"
  frontend_oauth2_client_id = "..."
  frontend_oauth2_client_secret = "..."
  # ... other variables
}
```

## Multiple environments in one project

Set `pubsub_topic_prefix` (e.g. `"dev-"`) so topic names don’t clash, and use different `vpc_connector_ip_cidr_range` per env (e.g. `10.8.0.0/28`, `10.8.0.16/28`). Use distinct `db_instance_name`, `media_bucket_name`, and `userdata_bucket_name` per environment.

## Outputs

- `cloud_run_api_url`, `cloud_run_frontend_url`
- `database_private_ip`, `database_public_ip`
- `artifact_registry_repository_id`
- `environment`
