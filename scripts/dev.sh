#!/usr/bin/env bash
set -euo pipefail

# Runs the Go server and the Vite dev server together. Open the Vite URL: it
# forwards /grpc, /api and /ws to Go, so the browser sees one origin, as it
# does in production. Ctrl-C stops both.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG="${NOSMALLPARTS_CONFIG_FILEPATH:-$ROOT/backend/config/localdev.json}"

# The server runs from a binary built here rather than under `go run`, whose
# child outlives it when it alone is killed.
BIN_DIR="$(mktemp -d)"

pids=()
stop() {
	trap - EXIT INT TERM
	for pid in ${pids[@]+"${pids[@]}"}; do
		kill "$pid" 2>/dev/null || true
	done
	wait || true
	rm -rf "$BIN_DIR"
}
trap stop EXIT INT TERM

(cd "$ROOT/backend" && go build -o "$BIN_DIR/nosmallparts" ./cmd/main)

# localdev keeps its SQLite file under backend/data, relative to the server.
mkdir -p "$ROOT/backend/data"
(cd "$ROOT/backend" && exec "$BIN_DIR/nosmallparts" --config "$CONFIG" serve) &
pids+=("$!")

pnpm --dir "$ROOT" --filter @nsp/web dev &
pids+=("$!")

# Either one exiting ends the session, so a server that failed to start is
# not hidden behind a dev server still answering. This polls rather than
# `wait -n`, which the bash macOS ships predates.
while kill -0 "${pids[0]}" 2>/dev/null && kill -0 "${pids[1]}" 2>/dev/null; do
	sleep 1
done
