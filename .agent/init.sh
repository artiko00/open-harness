#!/usr/bin/env bash
set -e

echo "==> Bootstrapping open-harness (Go workspace + npm)..."

echo "  → go work sync"
go work sync

echo "  → go build + test each tool"
for tool_dir in tools/*/; do
  tool=$(basename "$tool_dir")
  echo "    [$tool] build..."
  (cd "$tool_dir" && go build .)
  echo "    [$tool] test..."
  (cd "$tool_dir" && go test ./...)
done

echo "==> Environment ready."
echo "    Tools disponibles:"
for tool_dir in tools/*/; do
  tool=$(basename "$tool_dir")
  echo "      - $tool"
done
