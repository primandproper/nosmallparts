#!/usr/bin/env bash
set -euo pipefail

# Run MCP inspector
# Usage: mcp_inspector.sh [config_file]

CONFIG_FILE="${1:-mcp-server-config.json}"

npx @modelcontextprotocol/inspector --config "${CONFIG_FILE}"
