#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== Building all open-harness npm packages ==="
echo ""

for TOOL in linelens dupelens secretlens testlens; do
  bash "$ROOT/scripts/build-npm.sh" "$TOOL"
  echo ""
done

echo "=== All builds complete ==="
echo ""
echo "To publish all packages:"
echo "  bash scripts/publish-npm.sh"
