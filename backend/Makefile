# ENVIRONMENT
PWD      := $(shell pwd)
MYSELF   := $(shell id -u)
MY_GROUP := $(shell id -g)

# PATHS
THIS                   := github.com/verygoodsoftwarenotvirus/nosmallparts/backend
ARTIFACTS_DIR          := artifacts
SCRIPTS_DIR            := scripts
COVERAGE_OUT           := $(ARTIFACTS_DIR)/coverage.out
ENVIRONMENTS_DIR       := deploy/environments
TEST_COMPOSE_FILES_DIR := $(ENVIRONMENTS_DIR)/testing/compose_files
GENERATED_QUERIES_DIR  := internal/database/postgres/generated
PROTO_FILES_PATH       := proto/*.proto
PROTO_OUTPUT_PATH      := internal/grpc
WIRE_TARGETS           := internal/build/services/api/grpc \
	 internal/build/services/api/http \
	 internal/build/functions/data_change_message_handler \
	 internal/build/jobs/db_cleaner \
	 internal/build/jobs/search_data_index_scheduler

# COMPUTED
TOTAL_PACKAGE_LIST    := `go list $(THIS)/...`
TESTABLE_PACKAGE_LIST := `go list $(THIS)/... | grep -Ev '(cmd|integration|mock|fakes|converters|utils|generated)'`

# CONTAINER VERSIONS
CONTAINER_LINTER_IMAGE := openpolicyagent/conftest:v0.54.0
SQL_GENERATOR_IMAGE    := sqlc/sqlc:1.26.0
LINTER_IMAGE           := golangci/golangci-lint:v2.7.2
SHELLCHECK_IMAGE       := koalaman/shellcheck:stable
GO_IMAGE               := golang:1.25-trixie

# COMMANDS
GO_FORMAT             := gofmt -s -w
GO_TEST               := CGO_ENABLED=1 go test -shuffle=on -race -vet=all
CONTAINER_RUNNER      := docker
RUN_CONTAINER         := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --network=host
RUN_CONTAINER_AS_USER := $(RUN_CONTAINER) --user $(MYSELF):$(MY_GROUP)
LINTER                := $(RUN_CONTAINER) $(LINTER_IMAGE) golangci-lint
SQL_GENERATOR         := $(RUN_CONTAINER_AS_USER) $(SQL_GENERATOR_IMAGE)
CONTAINED_GO          := $(RUN_CONTAINER_AS_USER) --env HOME=/home/gouser --volume $(HOME)/.cache/go:/home/gouser $(GO_IMAGE) go

## non-PHONY folders/files

clean:
	$(SCRIPTS_DIR)/clean.sh $(ARTIFACTS_DIR)

$(ARTIFACTS_DIR):
	@mkdir -p $(ARTIFACTS_DIR)

clean-$(ARTIFACTS_DIR):
	@rm -rf $(ARTIFACTS_DIR)

## PREREQUISITES

.PHONY: ensure_wire_installed
ensure_wire_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh wire "go install github.com/google/wire/cmd/wire@v0.7.0"

.PHONY: ensure_fieldalignment_installed
ensure_fieldalignment_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh fieldalignment "go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@v0.37.0"

.PHONY: ensure_tagalign_installed
ensure_tagalign_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh tagalign "go install github.com/4meepo/tagalign/cmd/tagalign@v1.4.3"

.PHONY: ensure_gci_installed
ensure_gci_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh gci "go install github.com/daixiang0/gci@v0.13.5"

.PHONY: ensure_goimports_installed
ensure_goimports_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh goimports "go install golang.org/x/tools/cmd/goimports@v0.29.0"

.PHONY: ensure_protoc-gen-go_installed
ensure_protoc-gen-go_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh protoc-gen-go "go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.4"

.PHONY: ensure_protoc-gen-go-grpc_installed
ensure_protoc-gen-go-grpc_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh protoc-gen-go-grpc "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1"

.PHONY: ensure_air_installed
ensure_air_installed:
	@$(SCRIPTS_DIR)/ensure_tool_installed.sh air "go install github.com/air-verse/air@v1.63.0"

.PHONY: setup
setup: $(ARTIFACTS_DIR) revendor rewire configs ensure_wire_installed ensure_fieldalignment_installed ensure_tagalign_installed ensure_gci_installed

.PHONY: clean_vendor
clean_vendor:
	$(SCRIPTS_DIR)/clean_vendor.sh

vendor:
	$(SCRIPTS_DIR)/vendor.sh

.PHONY: revendor
revendor: clean_vendor vendor

## DEPENDENCY INJECTION

.PHONY: rewire
rewire: ensure_wire_installed
	$(SCRIPTS_DIR)/rewire.sh $(THIS) $(WIRE_TARGETS)

## FORMATTING

.PHONY: format_imports
format_imports: ensure_gci_installed
	$(SCRIPTS_DIR)/format_imports.sh $(THIS) $(PWD)

.PHONY: format_go_fieldalignment
format_go_fieldalignment: ensure_fieldalignment_installed
	@$(SCRIPTS_DIR)/format_go_fieldalignment.sh

.PHONY: format_go_tag_alignment
format_go_tag_alignment: ensure_tagalign_installed
	@$(SCRIPTS_DIR)/format_go_tag_alignment.sh

.PHONY: format_golang
format_golang: goimports format_imports format_go_fieldalignment # format_go_tag_alignment
	@$(SCRIPTS_DIR)/format_golang.sh $(PWD) "$(GO_FORMAT)"

.PHONY: terraformat
terraformat:
	@$(MAKE) -C .. terraformat

.PHONY: format
format: format_golang terraformat

.PHONY: fmt
fmt: format

.PHONY: goimports
goimports: ensure_goimports_installed
	$(SCRIPTS_DIR)/goimports.sh

## LINTING

.PHONY: lint_containers
lint_containers:
	@$(SCRIPTS_DIR)/lint_containers.sh $(CONTAINER_RUNNER) $(CONTAINER_LINTER_IMAGE) "$(RUN_CONTAINER_AS_USER)" $(PWD)

.PHONY: queries_lint
queries_lint:
	@$(SCRIPTS_DIR)/queries_lint.sh $(CONTAINER_RUNNER) $(SQL_GENERATOR_IMAGE) "$(SQL_GENERATOR)"

.PHONY: golang_lint
golang_lint: struct_usage_check
	@$(SCRIPTS_DIR)/golang_lint.sh $(CONTAINER_RUNNER) $(LINTER_IMAGE) "$(LINTER)"

.PHONY: lint_terraform
lint_terraform:
	@$(MAKE) -C .. lint_terraform

.PHONY: shellcheck
shellcheck:
	@$(SCRIPTS_DIR)/shellcheck.sh $(CONTAINER_RUNNER) $(SHELLCHECK_IMAGE) $(SCRIPTS_DIR)

.PHONY: lint
lint: lint_containers queries_lint golang_lint shellcheck

## GENERATED FILES

.PHONY: configs
configs:
	$(SCRIPTS_DIR)/configs.sh $(THIS)

.PHONY: queries
queries:
	$(SCRIPTS_DIR)/queries.sh $(THIS)

.PHONY: querier
querier: queries queries_lint
	@$(SCRIPTS_DIR)/querier.sh "$(SQL_GENERATOR)"

.PHONY: struct_usage_check
struct_usage_check:
	$(SCRIPTS_DIR)/struct_usage_check.sh $(THIS)

.PHONY: env_vars
env_vars:
	$(SCRIPTS_DIR)/env_vars.sh $(THIS)
	$(MAKE) format_golang

## INTEGRATION TESTING

.PHONY: integration-tests
integration-tests: integration_tests

.PHONY: integration_tests
integration_tests: integration_tests_postgres

.PHONY: integration_tests_postgres
integration_tests_postgres:
	$(SCRIPTS_DIR)/integration_tests_postgres.sh $(THIS)

## EXECUTION

.PHONY: build
build:
	$(SCRIPTS_DIR)/build.sh $(TOTAL_PACKAGE_LIST)

.PHONY: test
test: $(ARTIFACTS_DIR) vendor
	$(SCRIPTS_DIR)/test.sh

.PHONY: dev
dev: $(ARTIFACTS_DIR)
	$(SCRIPTS_DIR)/dev.sh $(THIS)

AIR_PORT := 9999

.PHONY: admin
admin: $(ARTIFACTS_DIR) ensure_air_installed
	$(SCRIPTS_DIR)/admin.sh $(THIS) $(ARTIFACTS_DIR)


.PHONY: mcp
mcp: $(ARTIFACTS_DIR) ensure_air_installed
	$(SCRIPTS_DIR)/mcp.sh $(THIS) $(ARTIFACTS_DIR)

.PHONY: mcp_inspector
mcp_inspector:
	$(SCRIPTS_DIR)/mcp_inspector.sh

.PHONY: agent
agent:
	$(SCRIPTS_DIR)/agent.sh $(THIS)

# OPERATIONS

GCP_DEV_PROJECT_NAME := no-small-parts-dev

# https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#connect-tcp
.PHONY: proxy_dev_db
proxy_dev_db:
	$(SCRIPTS_DIR)/proxy_dev_db.sh $(GCP_DEV_PROJECT_NAME)

# Deploy backend to Docker Desktop Kubernetes (localdev profile)
.PHONY: deploy_localdev
deploy_localdev:
	KO_DOCKER_REPO=ko.local skaffold run --filename=skaffold.yaml --build-concurrency 1 --profile localdev
