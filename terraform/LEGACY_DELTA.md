# Delta: Legacy Terraform vs Current (prod module)

This summarizes differences between the old Cloud Run setup in `legacy_terraform/` (what worked two years ago) and the current `terraform/modules/nosmallparts` + `terraform/environments/prod` setup.

---

## 1. **Database connectivity**

| Legacy | Current |
|--------|---------|
| Cloud SQL **Proxy** via sidecar: `volumes { cloud_sql_instance }` and `volume_mounts { mount_path = "/cloudsql" }`. App used `DINNER_DONE_BETTER_DATABASE_INSTANCE_CONNECTION_NAME` (Unix socket). | **Private IP** only: Cloud Run → VPC connector → Cloud SQL private IP. App gets `DATABASE_HOST` = private IP, no proxy. |
| Instance had **public** `ip_configuration` (SSL only, no private network). | Instance has **private** `ip_configuration` only; requires existing VPC + `google_service_networking_connection`. |

**Implication:** Current setup assumes you have a VPC (`private_network_name`) and reserved range (`private_ip_address_name`) already; legacy did not.

---

## 2. **Secrets and config**

| Legacy | Current |
|--------|---------|
| **Secret Manager** for: DB password, SendGrid, Segment, PostHog, Algolia, Google SSO, OAuth2 encryption key, JWT signing key. Env vars used `value_source { secret_key_ref { secret = "..." } }`. | **Plain env vars** from Terraform (vars → Cloud Run env). No Secret Manager in Terraform; secrets come from TFC vars / tfvars. |
| **Config file** in Secret Manager: `api_service_config` → mounted at `/config/service-config.json`; `CONFIGURATION_FILEPATH = "/config/service-config.json"`. | `CONFIGURATION_FILEPATH = "/etc/config"` but **no volume or secret is mounted** at `/etc/config` in the current module. Config may be missing at runtime unless provided another way. |

**Implication:** Either mount a config secret at `/etc/config` (and set path to the file inside it) or ensure the app can run without that file (e.g. env-only config).

---

## 3. **Environment variable names (app contract)**

| Legacy (Dinner Done Better) | Current app (No Small Parts) |
|-----------------------------|------------------------------|
| `DINNER_DONE_BETTER_DATABASE_*`, `DINNER_DONE_BETTER_DATA_CHANGES_TOPIC_NAME`, etc. | Config uses prefixes: `DATABASE_*` (with nested `READ_CONNECTION_*`), `QUEUES_*` (topic names), etc. Full env keys use `NO_SMALL_PARTS_` prefix (e.g. `NO_SMALL_PARTS_QUEUES_DATA_CHANGES_TOPIC_NAME`). |

**Current Cloud Run API env vars in Terraform:** We set `DATABASE_HOST`, `DATABASE_USERNAME`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `CONFIGURATION_FILEPATH`, and various API keys. We do **not** set:

- `DATABASE_READ_CONNECTION_*` (app may expect `DATABASE_READ_CONNECTION_HOST` etc. depending on config loader).
- **Any queue topic names:** `QUEUES_DATA_CHANGES_TOPIC_NAME`, `QUEUES_OUTBOUND_EMAILS_TOPIC_NAME`, `QUEUES_SEARCH_INDEX_REQUESTS_TOPIC_NAME`, `QUEUES_USER_DATA_AGGREGATION_TOPIC_NAME`, `QUEUES_WEBHOOK_EXECUTION_REQUESTS_TOPIC_NAME`. The API and jobs require these (see `QueuesConfig` validation).

---

## 4. **Pub/Sub topics**

| Legacy | Current |
|--------|---------|
| **data_changes** topic + Cloud Function publishing to outbound_emails, search_index_requests, webhook_execution, user_data_aggregator. | **No `data_changes` topic** in `pubsub.tf`. We have: `outbound_emails`, `search_index_requests`, `user_data_aggregation_requests`, `webhook_execution_requests` (and dead-letter topics). |
| Topic name for user aggregation: `user_data_aggregator_topic`. | Topic name: `user_data_aggregation_requests` (different). |

**Implication:** The app’s `QueuesConfig` requires `DATA_CHANGES_TOPIC_NAME`. Either add a `data_changes` topic (and wire it if something should publish to it) or change the app to make it optional. Topic names must match what the app expects (e.g. `QUEUES_USER_DATA_AGGREGATION_TOPIC_NAME` → current Terraform topic name).

---

## 5. **Cloud Functions and jobs (legacy only)**

Legacy had; **current Terraform has none:**

- **Cloud Functions:** data_changes handler, outbound emailer, search indexer, user data aggregator, webhook executor.
- **Cloud Run jobs:** db_cleaner, email_prober, meal_plan_finalizer, meal_plan_grocery_list_initializer, meal_plan_task_creator, search_index_scheduler.

Current setup is **API + frontend** Cloud Run services only. Any async behavior that relied on those functions/jobs would need to be reimplemented (e.g. workers in-cluster or different triggers).

---

## 6. **Algolia**

| Legacy | Current |
|--------|---------|
| Algolia provider + `search.tf`; secrets `ALGOLIA_APPLICATION_ID`, `ALGOLIA_API_KEY`; API server env from Secret Manager. | **Removed.** No Algolia in module or prod. |

---

## 7. **IAM and service identity**

| Legacy | Current |
|--------|---------|
| Custom role `api_server_role` + dedicated SA `api-server`; Cloud Run used `service_account = google_service_account.api_user_service_account.email`. | No custom role or dedicated SA in module; Cloud Run uses **default** compute SA (no `service_account` set). |

---

## 8. **Domain and DNS**

| Legacy | Current |
|--------|---------|
| `google_cloud_run_domain_mapping` for `api.dinnerdonebetter.dev`; Cloudflare CNAME for api (pointed at `ghs.googlehosted.com` in the snippet). | No domain mapping in module. Optional **Cloudflare DNS** records (db, media, userdata) when `cloudflare_zone_id` is set; no API/frontend domain mapping. |

---

## 9. **Frontend**

Legacy had **one** Cloud Run service (API). Current adds a **second** Cloud Run service for the Next.js frontend (`nosmallparts-frontend-*`) with its own env vars (backend URL, OAuth2 client, etc.).

---

## 10. **Artifact Registry and image names**

| Legacy | Current |
|--------|---------|
| Repository `containers`; image `containers/api_server`. | Repository from var `artifact_repository_id` (default `containers`); images `nosmallparts-api:latest`, `nosmallparts-frontend:latest`. |

---

## Summary: gaps to close for parity / correctness

1. **VPC / private DB:** Ensure prod has the VPC and reserved range the module expects, or add a path for “no VPC” (e.g. public DB or proxy) if you want legacy-style connectivity.
2. **Config file:** **Done.** The module creates a Secret Manager secret from `api_service_config_json`, mounts it at `/etc/config/service-config.json`, and sets `CONFIGURATION_FILEPATH`. Set `api_service_config_json` (sensitive) in prod.
3. **Queue topic names:** **Done.** `pubsub.tf` has `data_changes_topic`; the API container gets all five `NO_SMALL_PARTS_QUEUES_*_TOPIC_NAME` env vars.
4. **Database env shape:** Confirm whether the app expects `DATABASE_HOST` or `DATABASE_READ_CONNECTION_HOST` (and same for other fields) and align Cloud Run env with that.
5. **Secrets:** Current design uses Terraform/TFC vars → env; legacy used Secret Manager. For production you may want to reintroduce Secret Manager for sensitive values and reference them in Cloud Run.
6. **Async workers:** If you need the behavior of the old Cloud Functions/jobs, add equivalent workers or triggers in the new setup.
