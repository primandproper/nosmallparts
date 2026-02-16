# AdSense diagnosis for nosmall.parts

## Quick diagnostic

1. Visit any page with `?ads_debug=1` (e.g. `https://nosmall.parts/?ads_debug=1`)
2. Open DevTools → Console
3. Look for `[AdSense] Diagnostic:` and `[AdSense] Ad slot diagnostic:` messages

## What the diagnostics mean

### `[AdSense] Diagnostic:`

- **NEXT_PUBLIC_ADSENSE_PUBLISHER_ID: NOT SET** → The AdSense script will never load. This var is inlined at **build time**—runtime env vars are not available client-side. You must pass it as `--build-arg` during `docker build`.
- **scriptWillLoad: true** → The script should load; check Network tab for `pagead2.googlesyndication.com`.

### `[AdSense] Ad slot diagnostic:` (only if AdBanner renders)

- **client: missing** or **slot: missing** → `ADSENSE_PUBLISHER_ID` / `ADSENSE_SLOT_ID` are not set at runtime on the server. For GKE, add them to the frontend deployment env.
- **adsbygoogleLoaded: false** → The script hasn’t loaded yet (or failed). Wait a moment; if it stays false, the publisher ID may be missing at build time.

## Deployment checklist

### 1. Build time (required for script to load)

`NEXT_PUBLIC_ADSENSE_PUBLISHER_ID` must be present when the Next.js app is built. In Docker/Skaffold, pass it as a build arg:

```bash
docker build \
  --build-arg NEXT_PUBLIC_ADSENSE_PUBLISHER_ID=ca-pub-xxxxxxxxxxxxxxxx \
  -f frontend/Dockerfile frontend/
```

Skaffold’s default docker build does not pass build args. You need to configure it in `frontend/skaffold.yaml` or in the CI workflow.

### 2. Runtime (required for ad slots to render)

`ADSENSE_PUBLISHER_ID` and `ADSENSE_SLOT_ID` must be available to the Next.js server at runtime. For GKE:

- Add them to the `frontend-config` secret (backend Terraform), or
- Add them to the deployment env patch in `frontend/deploy/environments/prod/kustomize/patches/`.

The `deployment-backend-env.yaml` patch currently does **not** include AdSense vars.

### 3. AdSense account / site approval

- New sites can take a few days to 2–4 weeks for approval.
- Use the [AdSense ads not showing troubleshooter](https://support.google.com/adsense/workflow/10383839).
- Ensure the domain is added and verified in your AdSense account.

## Cloud Run vs GKE

- **Cloud Run** (frontend Terraform): `ADSENSE_PUBLISHER_ID` and `ADSENSE_SLOT_ID` are set in the Cloud Run service env. Build args still need to be passed when building the image.
- **GKE** (Skaffold deploy): The frontend gets env from `frontend-config` and `deployment-backend-env.yaml`. AdSense vars are not configured there; add them to enable ads.
