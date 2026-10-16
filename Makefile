PORT ?= 3001
IMAGE ?= nosmallparts:local

PWD                   := $(shell pwd)
RUN_CONTAINER_AS_USER := docker run --rm --volume $(PWD):$(PWD) --workdir=$(PWD) --user $(shell id -u):$(shell id -g)
PROTOLINT             := $(RUN_CONTAINER_AS_USER) yoheimuta/protolint:0.57.0
SHELLCHECK            := $(RUN_CONTAINER_AS_USER) koalaman/shellcheck:v0.11.0
GENERATED_PROTO_PATHS := backend/internal/proto frontend/src/generated

## FORMATTING, LINTING, TESTING

.PHONY: format
format: format_proto
	$(MAKE) -C backend format
	pnpm --filter @nsp/web format

.PHONY: lint
lint: lint_proto check_proto lint_scripts
	$(MAKE) -C backend lint
	pnpm --filter @nsp/web typecheck
	pnpm --filter @nsp/web format:check

.PHONY: test
test:
	$(MAKE) -C backend test
	pnpm --filter @nsp/web test

.PHONY: lint_scripts
lint_scripts:
	$(SHELLCHECK) scripts/*.sh

# e2e runs the Playwright suite against the built binary on a fresh database.
.PHONY: e2e
e2e: build
	./scripts/e2e.sh

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

## BUILDING AND RUNNING

# web builds the web app into the directory the Go binary embeds it from.
.PHONY: web
web:
	./scripts/web.sh

# build produces the one binary that serves everything, web app included.
.PHONY: build
build: web
	$(MAKE) -C backend build

# seed writes the built-in sample plays into the localdev database.
.PHONY: seed
seed:
	@mkdir -p backend/data
	$(MAKE) -C backend run ARGS="--config config/localdev.json seed"

# dev runs the Go server and the Vite dev server; open the Vite URL.
.PHONY: dev
dev:
	./scripts/dev.sh

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

# docker-run serves the production image locally, with localdev's keys (no
# secret: they are in backend/config/localdev.json) in place of the ones a
# deploy sets. LiveKit stays off unless its variables are set.
.PHONY: docker-run
docker-run:
	@mkdir -p .docker-data
	docker run --rm -it -p $(PORT):3001 \
		-e NOSMALLPARTS_TOKENS_SIGNING_KEY=bm9zbWFsbHBhcnRzLWxvY2FsZGV2LXNpZ25pbmctMDA= \
		-e NOSMALLPARTS_COOKIES_HASH_KEY=bm9zbWFsbHBhcnRzLWxvY2FsZGV2LWNvb2tpZS1oYXNoLWtleS0wMDA= \
		-e NOSMALLPARTS_COOKIES_BLOCK_KEY=bm9zbWFsbHBhcnRzLWxvY2FsZGV2LWJsb2Nra2V5MDA= \
		-e NOSMALLPARTS_LIVEKIT_API_KEY -e NOSMALLPARTS_LIVEKIT_API_SECRET -e NOSMALLPARTS_LIVEKIT_URL \
		-v $(CURDIR)/.docker-data:/data \
		$(IMAGE)

.PHONY: fly-deploy
fly-deploy:
	flyctl deploy

.PHONY: fly-deploy-remote
fly-deploy-remote:
	flyctl deploy --remote-only
