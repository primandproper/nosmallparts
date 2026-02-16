# AdSense diagnosis for nosmall.parts

## Quick diagnostic

1. Visit any page with `?ads_debug=1` (e.g. `https://nosmall.parts/?ads_debug=1`)
2. Open DevTools → Console
3. Look for `[AdSense] Diagnostic:` and `[AdSense] Ad slot diagnostic:` messages

## What the diagnostics mean

### `[AdSense] Diagnostic:`

- **NEXT_PUBLIC_ADSENSE_PUBLISHER_ID: NOT SET** → The AdSense script will never load. Ensure GitHub secrets and Terraform variables are configured (see below).
- **scriptWillLoad: true** → The script should load; check Network tab for `pagead2.googlesyndication.com`.

### `[AdSense] Ad slot diagnostic:` (only if AdBanner renders)

- **client: missing** or **slot: missing** → `ADSENSE_PUBLISHER_ID` / `ADSENSE_SLOT_ID` not set at runtime. Add them to Terraform Cloud (prod-backend workspace).
- **adsbygoogleLoaded: false** → The script hasn't loaded yet (or failed). Wait a moment; if it stays false, the publisher ID may be missing at build time.

## Required configuration

### 1. GitHub Actions secrets (for build-time inlining)

Add these repo secrets; they are passed to Skaffold for `NEXT_PUBLIC_*` build args:

- `NO_SMALL_PARTS_ADSENSE_PUBLISHER_ID` — e.g. `ca-pub-xxxxxxxxxxxxxxxx`
- `NO_SMALL_PARTS_ADSENSE_SLOT_ID` — e.g. `1234567890`

### 2. Terraform Cloud variables (for runtime / frontend-config secret)

In the **prod-backend** workspace, set:

- `ADSENSE_PUBLISHER_ID` — same value as above
- `ADSENSE_SLOT_ID` — same value as above

These are stored in the `frontend-config` K8s secret and mounted as env in the frontend pod.

### 3. AdSense account / site approval

- New sites can take a few days to 2–4 weeks for approval.
- Use the [AdSense ads not showing troubleshooter](https://support.google.com/adsense/workflow/10383839).
- Ensure the domain is added and verified in your AdSense account.
