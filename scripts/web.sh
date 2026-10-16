#!/usr/bin/env bash
set -euo pipefail

# Builds the web app and copies it to where the Go binary embeds it from:
# go:embed cannot reach outside the backend module. The copy replaces the
# previous build, so assets a rebuild renamed do not linger in the binary.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/frontend/dist/"
DEST="$ROOT/backend/internal/web/dist/"

pnpm --dir "$ROOT" --filter @nsp/web build

rsync -a --delete --exclude README.md "$SRC" "$DEST"

echo "web app copied to ${DEST#"$ROOT"/}"
