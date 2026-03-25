# Fly.io Deployment

Parallel deployment of the API server and frontend to Fly.io. The Kubernetes deployment remains intact and can be switched back to at any time via DNS.

## Architecture

- **nsp-api**: Go backend (HTTP :8000 public, gRPC :8001 private network only)
- **nsp-frontend**: Next.js frontend (:3000 public)
- **nsp-db**: Fly Postgres

The frontend reaches the API over Fly's private `.internal` network. gRPC on port 8001 is never exposed publicly.

## One-Time Provisioning

### 1. Create Apps

```bash
fly apps create nsp-api
fly apps create nsp-frontend
```

### 2. Create Fly Postgres

```bash
fly postgres create --name nsp-db --region iad --vm-size shared-cpu-1x --initial-cluster-size 1 --volume-size 1
```

### 3. Attach Database

```bash
fly postgres attach nsp-db --app nsp-api --database-name no_small_parts
```

This outputs a `DATABASE_URL`. Parse the username and password from it for the next step.

### 4. Create Uploads Volume

```bash
fly volumes create uploads_data --app nsp-api --region iad --size 1
```

### 5. Set API Secrets

Replace placeholder values with the actual credentials from step 3.

```bash
fly secrets set --app nsp-api \
  NO_SMALL_PARTS_DATABASE_READ_CONNECTION_HOST="nsp-db.flycast" \
  NO_SMALL_PARTS_DATABASE_READ_CONNECTION_USERNAME="<from attach output>" \
  NO_SMALL_PARTS_DATABASE_READ_CONNECTION_PASSWORD="<from attach output>" \
  NO_SMALL_PARTS_DATABASE_READ_CONNECTION_DATABASE="no_small_parts" \
  NO_SMALL_PARTS_DATABASE_READ_CONNECTION_PORT="5432" \
  NO_SMALL_PARTS_DATABASE_WRITE_CONNECTION_HOST="nsp-db.flycast" \
  NO_SMALL_PARTS_DATABASE_WRITE_CONNECTION_USERNAME="<from attach output>" \
  NO_SMALL_PARTS_DATABASE_WRITE_CONNECTION_PASSWORD="<from attach output>" \
  NO_SMALL_PARTS_DATABASE_WRITE_CONNECTION_DATABASE="no_small_parts" \
  NO_SMALL_PARTS_DATABASE_WRITE_CONNECTION_PORT="5432" \
  NO_SMALL_PARTS_DATABASE_OAUTH2_TOKEN_ENCRYPTION_KEY="<generate a 32-character secret>" \
  NO_SMALL_PARTS_SERVICE_AUTH_TOKENS_BASE64_ENCODED_SIGNING_KEY="<base64-encoded 32-character secret>"
```

### 6. Set Frontend Secrets

```bash
fly secrets set --app nsp-frontend \
  OAUTH2_CLIENT_ID="<your OAuth2 client ID>" \
  OAUTH2_CLIENT_SECRET="<your OAuth2 client secret>"
```

### 7. Deploy

From the repo root:

```bash
# Deploy API
fly deploy --config fly/api/fly.toml --remote-only

# Deploy frontend
fly deploy --config fly/frontend/fly.toml \
  --dockerfile frontend/Dockerfile \
  --build-arg NEXT_PUBLIC_WS_URL=wss://nsp-frontend.fly.dev/ws/stage \
  frontend/
```

### 8. Verify

```bash
curl -I https://nsp-api.fly.dev/_ops_/ready
curl -I https://nsp-frontend.fly.dev/
```

### 9. Custom Domains (when ready)

```bash
fly certs create api.nosmall.parts --app nsp-api
fly certs create nosmall.parts --app nsp-frontend
fly certs create www.nosmall.parts --app nsp-frontend
```

### 10. GitHub Actions

Generate a deploy token and add it as a repository secret:

```bash
fly tokens create deploy --expiry 8760h
# Add the token as FLY_API_TOKEN in GitHub Settings > Secrets > Actions
```

## DNS Swap (Cloudflare)

### Point DNS to Fly

Get Fly IP addresses:

```bash
fly ips list --app nsp-api
fly ips list --app nsp-frontend
```

In Cloudflare, update records (set proxy to DNS-only / gray cloud so Fly handles TLS):

| Record | Type | Value |
|--------|------|-------|
| `api.nosmall.parts` | CNAME | `nsp-api.fly.dev` |
| `nosmall.parts` | CNAME | `nsp-frontend.fly.dev` (Cloudflare flattens at apex) |
| `www.nosmall.parts` | CNAME | `nsp-frontend.fly.dev` |

Alternatively, use A/AAAA records pointing to the IPs from `fly ips list`.

### Verify

```bash
curl -I https://api.nosmall.parts/_ops_/ready
curl -I https://nosmall.parts/
curl -I https://www.nosmall.parts/
```

### Rollback

Revert the Cloudflare DNS records to their previous values (GKE ingress IP).

## Useful Commands

```bash
# Logs
fly logs --app nsp-api
fly logs --app nsp-frontend

# SSH into a machine
fly ssh console --app nsp-api

# Postgres console
fly postgres connect --app nsp-db --database no_small_parts

# Scale (if you need always-on)
fly scale count 1 --app nsp-api
fly machine update --autostart --no-auto-stop --app nsp-api <machine-id>
```
