# Dinner Done Better API

[![backend - deploy](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_deploy_dev.yaml/badge.svg)](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_deploy_dev.yaml) [![backend - integration tests](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_integration_tests.yaml/badge.svg)](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_integration_tests.yaml) [![backend - lint](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_lint.yaml/badge.svg)](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_lint.yaml) [![backend - unit tests](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_unit_tests.yaml/badge.svg)](https://github.com/dinnerdonebetter/dinnerdonebetter/actions/workflows/backend_unit_tests.yaml)

## dev dependencies

The following tools are prerequisites for development work:

- [go](https://golang.org/) 1.25
- [docker](https://docs.docker.com/get-docker/) &&  [docker-compose](https://docs.docker.com/compose/install/)
- [wire](https://github.com/google/wire) for dependency management
- [make](https://www.gnu.org/software/make/) for task running
- [sqlc](https://sqlc.dev/) for generating database code
- [gci](https://www.github.com/daixiang0/gci) for sorting imports
- [tagalign](https://www.github.com/4meepo/tagalign) for aligning struct tags (`go install github.com/4meepo/tagalign/cmd/tagalign@latest`)
- [terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) for deploying/formatting
- [cloud_sql_proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) for production database access
- [fieldalignment](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/fieldalignment) (`go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest`)

## dev setup

It's a good idea to run `make quicktest lint integration_tests` before commits.

## running the server

1. clone this repository
2. run `make dev`
3. [http://localhost:8000/](http://localhost:8000/)

## infrastructure

```mermaid
flowchart LR
    PublicInternet("Public Internet")
    APIServer("API Server")
    Sendgrid("Sendgrid")
    Segment("Segment")
    Algolia("Algolia")
    Database("Database")
    DataChangesQueue("Data Changes Queue")
    DataChangesWorker("Data Changes Worker")
    MealPlanFinalizerWorker("Meal Plan Finalizer")
    MealPlanGroceryListInitializerWorker("Grocery List Initializer")
    MealPlanTaskCreatorWorker("Meal Plan Task Creator")
    OutboundEmailerWorker("Outbound Emailer")
    SearchDataIndexSchedulerWorker("Data Index Scheduler")
    SearchDataIndexerWorker("Search Data Indexer")
    PublicInternet-->APIServer
    Cron-->MealPlanGroceryListInitializerWorker
    Cron-->SearchDataIndexSchedulerWorker
    Cron-->MealPlanTaskCreatorWorker
    Cron-->MealPlanFinalizerWorker
    DataChangesWorker-->OutboundEmailerWorker
    DataChangesWorker-->SearchDataIndexerWorker
    MealPlanGroceryListInitializerWorker-.->DataChangesQueue
    MealPlanTaskCreatorWorker-.->DataChangesQueue
    MealPlanFinalizerWorker-.->DataChangesQueue
    SearchDataIndexSchedulerWorker-->SearchDataIndexerWorker
    SearchDataIndexerWorker-->Algolia
    DataChangesQueue-->DataChangesWorker
    DataChangesWorker-->Segment
    OutboundEmailerWorker-->Segment
    OutboundEmailerWorker-->Sendgrid
    Algolia-->APIServer
    APIServer-->Database
    Database-->APIServer
    APIServer-->DataChangesQueue
```
