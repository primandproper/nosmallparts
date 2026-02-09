# Import existing resources (run from this directory)

Replace `YOUR_PROJECT_ID` with your actual GCP project ID (same as `project_id`).

```bash
# VPC Access Connector (name: prod-cloud-run-connector, region: us-central1)
terraform import 'module.api.google_vpc_access_connector.cloud_run' \
  projects/YOUR_PROJECT_ID/locations/us-central1/connectors/prod-cloud-run-connector

# Artifact Registry repository (id: containers, region: us-central1)
terraform import 'module.api.google_artifact_registry_repository.main' \
  projects/YOUR_PROJECT_ID/locations/us-central1/repositories/containers
```

Then run `terraform plan` again; it should show no create for these resources (only any desired changes).
