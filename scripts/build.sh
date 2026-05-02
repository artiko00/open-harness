#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

for tool_dir in "$ROOT"/tools/*/; do
  tool=$(basename "$tool_dir")
  echo "==> Building $tool..."
  go build -o "$tool_dir/$tool" "$tool_dir"
  echo "    OK: $tool_dir/$tool"
done
