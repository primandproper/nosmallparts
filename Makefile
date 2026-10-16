PORT ?= 3001
IMAGE ?= nosmallparts:local

PWD                   := $(shell pwd)
RUN_CONTAINER_AS_USER := docker run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --user $(shell id -u):$(shell id -g)
PROTOLINT             := $(RUN_CONTAINER_AS_USER) yoheimuta/protolint:0.57.0
GENERATED_PROTO_PATHS := backend/internal/proto frontend/src/generated

## FORMATTING, LINTING, TESTING

.PHONY: format
format: format_proto
	$(MAKE) -C backend format
	pnpm --filter @nsp/web format

.PHONY: lint
lint: lint_proto check_proto
	$(MAKE) -C backend lint
	pnpm --filter @nsp/web typecheck
	pnpm --filter @nsp/web format:check

# test runs the Go suite and the TypeScript server's suite, which stays until the
# TypeScript server is deleted, once every one of its cases has a ported counterpart.
.PHONY: test
test:
	$(MAKE) -C backend test
	pnpm --filter @nsp/web test
	pnpm test

## PROTOBUF

# proto formats what it generates with the backend's formatters, which touch the
# generated Go's imports, so check_proto compares like with like.
.PHONY: proto
proto:
	./scripts/proto.sh
	$(MAKE) -C backend format

.PHONY: format_proto
format_proto:
	$(PROTOLINT) lint -fix proto/

.PHONY: lint_proto
lint_proto:
	$(PROTOLINT) lint proto/

# check_proto regenerates and fails if that changed anything, so generated code
# cannot drift from the schema it claims to come from.
.PHONY: check_proto
check_proto: proto
	@git diff --exit-code -- $(GENERATED_PROTO_PATHS) || \
		{ echo "generated code is stale: run make proto and commit the result" >&2; exit 1; }
	@test -z "$$(git ls-files --others --exclude-standard -- $(GENERATED_PROTO_PATHS))" || \
		{ echo "generated files are not committed: run make proto and commit the result" >&2; exit 1; }

## RUNNING

.PHONY: lan-info
lan-info:
	@ip=$$(ipconfig getifaddr en0 2>/dev/null); \
	if [ -z "$$ip" ]; then ip=$$(ipconfig getifaddr en1 2>/dev/null); fi; \
	if [ -z "$$ip" ]; then ip=$$(route -n get default 2>/dev/null | awk '/interface:/ {print $$2}' | xargs -I{} ipconfig getifaddr {} 2>/dev/null); fi; \
	if [ -z "$$ip" ]; then echo "Could not determine LAN IP"; exit 1; fi; \
	echo "Server reachable at: http://$$ip:$(PORT)"

.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE) .

.PHONY: docker-run
docker-run:
	@mkdir -p .docker-data
	docker run --rm -it -p $(PORT):3001 \
		-e BETTER_AUTH_SECRET="$${BETTER_AUTH_SECRET:-dev-secret-not-for-prod}" \
		-e LIVEKIT_API_KEY -e LIVEKIT_API_SECRET -e LIVEKIT_URL \
		-e TRUSTED_ORIGINS="http://localhost:$(PORT)" \
		-e WEB_STATIC_DIR=/app/web \
		-v $(CURDIR)/.docker-data:/data \
		$(IMAGE)

.PHONY: fly-deploy
fly-deploy:
	flyctl deploy

.PHONY: fly-deploy-remote
fly-deploy-remote:
	flyctl deploy --remote-only
