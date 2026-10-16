#!/usr/bin/env bash
set -euo pipefail

# Runs the Playwright suite in frontend/e2e against the built binary (make
# build), serving a fresh SQLite database on ports of its own so it can run
# beside `make dev`. Arguments go to `playwright test`.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="$ROOT/backend/artifacts/nosmallparts"
CONFIG="$ROOT/backend/config/localdev.json"
HTTP_PORT="${NSP_E2E_HTTP_PORT:-3101}"
GRPC_PORT="${NSP_E2E_GRPC_PORT:-50151}"

if [[ ! -x "$BIN" ]]; then
	echo "no server binary at ${BIN#"$ROOT"/}: run make build first" >&2
	exit 1
fi

WORK="$(mktemp -d)"
server_pid=""
# shellcheck disable=SC2329 # the EXIT trap calls it
cleanup() {
	if [[ -n "$server_pid" ]]; then
		kill "$server_pid" 2>/dev/null || true
		wait "$server_pid" 2>/dev/null || true
	fi
	rm -rf "$WORK"
}
trap cleanup EXIT

export NOSMALLPARTS_DATABASE_READ_CONNECTION_DATABASE="$WORK/e2e.db"
export NOSMALLPARTS_DATABASE_WRITE_CONNECTION_DATABASE="$WORK/e2e.db"
export NOSMALLPARTS_HTTP_SERVER_PORT="$HTTP_PORT"
export NOSMALLPARTS_GRPC_SERVER_PORT="$GRPC_PORT"

"$BIN" --config "$CONFIG" seed >"$WORK/seed.log" 2>&1 || {
	cat "$WORK/seed.log" >&2
	exit 1
}

"$BIN" --config "$CONFIG" serve >"$WORK/server.log" 2>&1 &
server_pid=$!

ready=""
for _ in $(seq 1 60); do
	if curl -sf -o /dev/null "http://localhost:$HTTP_PORT/readyz"; then
		ready=1
		break
	fi
	if ! kill -0 "$server_pid" 2>/dev/null; then
		break
	fi
	sleep 0.5
done

if [[ -z "$ready" ]]; then
	echo "the server did not become ready; its log:" >&2
	cat "$WORK/server.log" >&2
	exit 1
fi

# A no-op once the browser is installed.
pnpm --dir "$ROOT" --filter @nsp/web exec playwright install chromium

status=0
NSP_E2E_BASE_URL="http://localhost:$HTTP_PORT" \
	pnpm --dir "$ROOT" --filter @nsp/web exec playwright test "$@" || status=$?

if [[ "$status" -ne 0 ]]; then
	echo "the suite failed; the last of the server's log:" >&2
	tail -n 50 "$WORK/server.log" >&2
fi

exit "$status"
