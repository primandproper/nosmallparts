PORT ?= 3001
IMAGE ?= nosmallparts:local

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
		-v $(CURDIR)/.docker-data:/data \
		$(IMAGE)

.PHONY: fly-deploy
fly-deploy:
	flyctl deploy

.PHONY: fly-deploy-remote
fly-deploy-remote:
	flyctl deploy --remote-only
