PORT ?= 3001

.PHONY: lan-info
lan-info:
	@ip=$$(ipconfig getifaddr en0 2>/dev/null); \
	if [ -z "$$ip" ]; then ip=$$(ipconfig getifaddr en1 2>/dev/null); fi; \
	if [ -z "$$ip" ]; then ip=$$(route -n get default 2>/dev/null | awk '/interface:/ {print $$2}' | xargs -I{} ipconfig getifaddr {} 2>/dev/null); fi; \
	if [ -z "$$ip" ]; then echo "Could not determine LAN IP"; exit 1; fi; \
	echo "Server reachable at: http://$$ip:$(PORT)"
