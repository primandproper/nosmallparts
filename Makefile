PWD           := $(shell pwd)
MYSELF        := $(shell id -u)
MY_GROUP      := $(shell id -g)

# CONTAINER VERSIONS
PROTOBUF_FORMAT := bufbuild/buf:1.5.0

# COMMANDS
CONTAINER_RUNNER      := docker
RUN_CONTAINER         := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD)
RUN_CONTAINER_AS_USER := $(CONTAINER_RUNNER) run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --user $(MYSELF):$(MY_GROUP)
FORMAT_PROTOBUFS      := $(RUN_CONTAINER) $(PROTOBUF_FORMAT)

.PHONY: ensure_yamlfmt_installed
ensure_yamlfmt_installed:
ifeq (, $(shell which yamlfmt))
	$(shell go install github.com/google/yamlfmt/cmd/yamlfmt@latest)
endif

.PHONY: ensure_protoc_installed
ensure_protoc_installed:
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell brew install protobuf)
endif

.PHONY: ensure_protoc-gen-go_installed
ensure_protoc-gen-go_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.4)
endif

.PHONY: ensure_protoc-gen-go-grpc_installed
ensure_protoc-gen-go-grpc_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-go-grpc))
	$(shell go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1)
endif

.PHONY: ensure_protoc-gen-swift_installed
ensure_protoc-gen-swift_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-swift))
	$(shell brew install swift-protobuf)
endif

.PHONY: ensure_protoc-gen-grpc-swift_installed
ensure_protoc-gen-grpc-swift_installed: ensure_protoc_installed
ifeq (, $(shell which protoc-gen-grpc-swift-2))
	$(shell brew install protoc-gen-grpc-swift)
endif

.PHONY: lint_markdown
lint_markdown:
	./scripts/lint_markdown.sh

.PHONY: setup
setup: ensure_yamlfmt_installed
	(cd backend && $(MAKE) setup)
	(cd frontend && $(MAKE) setup)

.PHONY: format
format:
	(cd backend && $(MAKE) format)
	(cd frontend && $(MAKE) format)

# Terraform: format from root; init/validate from prod environment
TERRAFORM_DIR := terraform
TERRAFORM_ENV := $(TERRAFORM_DIR)/environments/prod

.PHONY: terraformat
terraformat:
	cd $(TERRAFORM_DIR) && terraform fmt -recursive

.PHONY: lint_terraform
lint_terraform: terraformat
	cd $(TERRAFORM_ENV) && terraform init -upgrade && terraform validate && terraform fmt -check -recursive

.PHONY: format_yaml
format_yaml: ensure_yamlfmt_installed
	yamlfmt -conf .yamlfmt.yaml

.PHONY: lint
lint:
	(cd backend && $(MAKE) lint)
	(cd frontend && $(MAKE) lint)

.PHONY: test
test:
	(cd backend && $(MAKE) test)
	(cd frontend && $(MAKE) test)

.PHONY: format_proto
format_proto:
	$(FORMAT_PROTOBUFS) format --path proto --write

# PATHS
# Exclude monolithic {domain}/{domain}.proto files (they're duplicates of the split files)
# But keep common.proto and filtering.proto at the root, and uploaded_media.proto (hasn't been split yet)
PROTO_FILES_PATH          := $(shell find proto -name "*.proto" -type f ! -path "proto/audit/audit.proto" ! -path "proto/auth/auth.proto" ! -path "proto/dataprivacy/dataprivacy.proto" ! -path "proto/identity/identity.proto" ! -path "proto/internal_ops/internal_ops.proto" ! -path "proto/issue_reports/issue_reports.proto" ! -path "proto/notifications/notifications.proto" ! -path "proto/oauth/oauth.proto" ! -path "proto/settings/settings.proto" ! -path "proto/waitlists/waitlists.proto" ! -path "proto/webhooks/webhooks.proto")
PROTO_GO_OUTPUT_PATH      := backend
PROTO_OUTPUT_BACKEND_PATH := backend/internal/grpc
BACKEND_MODULE_NAME       := github.com/verygoodsoftwarenotvirus/nosmallparts/backend
PROTO_TS_OUTPUT_PATH      := frontend/generated

.PHONY: proto_golang
proto_golang: ensure_protoc_installed ensure_protoc-gen-go_installed ensure_protoc-gen-go-grpc_installed
	mkdir -p $(PROTO_OUTPUT_BACKEND_PATH)
	protoc --go_out=$(PROTO_GO_OUTPUT_PATH) \
		--go-grpc_out=$(PROTO_GO_OUTPUT_PATH) \
		--go_opt=module=$(BACKEND_MODULE_NAME) \
		--go-grpc_opt=module=$(BACKEND_MODULE_NAME) \
		--proto_path proto/ \
		$(PROTO_FILES_PATH);
	(cd backend && $(MAKE) format_golang)

.PHONY: proto_typescript
proto_typescript: ensure_protoc_installed
	mkdir -p $(PROTO_TS_OUTPUT_PATH)
	protoc --plugin=./frontend/node_modules/.bin/protoc-gen-ts_proto \
		--ts_proto_out=$(PROTO_TS_OUTPUT_PATH) \
		--ts_proto_opt=esModuleInterop=true \
		--ts_proto_opt=outputServices=nice-grpc \
		--ts_proto_opt=outputServices=generic-definitions \
		--ts_proto_opt=useExactTypes=false \
		--ts_proto_opt=snakeToCamel=true \
		--proto_path proto/ \
		$(PROTO_FILES_PATH)
	(cd frontend && $(MAKE) format)

.PHONY: proto
proto: format_proto proto_golang proto_typescript

# Docker targets for running the full stack
.PHONY: run
run:
	$(CONTAINER_RUNNER) compose up --build

# -----------------------------------------------------------------------------
# Google Cloud Run deployment
# -----------------------------------------------------------------------------
# Set GCP_PROJECT (defaults to gcloud config project), GCP_REGION, ARTIFACT_REPO as needed.
# Prereqs: gcloud CLI, Docker; run: gcloud auth configure-docker $(GCP_REGION)-docker.pkg.dev
GCP_PROJECT    ?= no-small-parts
GCP_REGION     ?= us-central1
ARTIFACT_REPO  ?= cloud-run-images

API_IMAGE      := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(ARTIFACT_REPO)/nosmallparts-api
FRONTEND_IMAGE := $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(ARTIFACT_REPO)/nosmallparts-frontend

.PHONY: gcr-ensure-repo
gcr-ensure-repo:
	gcloud artifacts repositories create $(ARTIFACT_REPO) \
		--repository-format=docker \
		--location=$(GCP_REGION) \
		--description="Images for No Small Parts Cloud Run" \
		|| true

.PHONY: gcr-deploy-api
gcr-deploy-api: gcr-ensure-repo
	$(CONTAINER_RUNNER) build -f backend/deploy/dockerfiles/api_server.Dockerfile -t $(API_IMAGE):latest ./backend
	$(CONTAINER_RUNNER) push $(API_IMAGE):latest
	gcloud run deploy nosmallparts-api \
		--image $(API_IMAGE):latest \
		--region $(GCP_REGION) \
		--platform managed \
		--allow-unauthenticated

.PHONY: gcr-deploy-frontend
gcr-deploy-frontend: gcr-ensure-repo
	$(CONTAINER_RUNNER) build -t $(FRONTEND_IMAGE):latest ./frontend
	$(CONTAINER_RUNNER) push $(FRONTEND_IMAGE):latest
	gcloud run deploy nosmallparts-frontend \
		--image $(FRONTEND_IMAGE):latest \
		--region $(GCP_REGION) \
		--platform managed \
		--allow-unauthenticated

.PHONY: gcr-deploy
gcr-deploy: gcr-deploy-api gcr-deploy-frontend
