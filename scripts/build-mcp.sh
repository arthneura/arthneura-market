#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
mkdir -p dist
go build -o dist/arthneura-mcp ./cmd/arthneura-mcp
echo "WROTE $ROOT/dist/arthneura-mcp"
ls -l dist/arthneura-mcp
