# Infrastructure

Terraform for GKE cluster and networking; Kustomize for cluster add-ons (ingress, cert-manager, external-dns, otel).

## Dev environment

1. Apply infra Terraform: `infra/deploy/environments/prod/terraform`
2. Apply backend Terraform: `backend/deploy/environments/prod/terraform`
3. Deploy via Skaffold: `skaffold run --profile prod`
