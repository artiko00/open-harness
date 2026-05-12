#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TOOL=$1

if [ -z "$TOOL" ]; then
  echo "Usage: build-npm.sh <tool>"
  echo "  build-npm.sh linelens"
  exit 1
fi

TOOL_DIR="$ROOT/tools/$TOOL"
NPM_DIR="$ROOT/npm/@open_harness"
VERSION=$(grep '^const version' "$TOOL_DIR/main.go" | grep -oE '"[^"]+"' | tr -d '"')

echo "Building $TOOL v$VERSION for all platforms..."

build() {
  local goos=$1 goarch=$2 pkg=$3 ext=$4
  local out="$NPM_DIR/$TOOL-$pkg/bin/$TOOL$ext"
  echo "  → $pkg"
  GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w" -o "$out" "$TOOL_DIR"
  if [ -z "$ext" ]; then chmod +x "$out"; fi
}

build linux   amd64 linux-x64    ""
build darwin  arm64 darwin-arm64 ""
build darwin  amd64 darwin-x64   ""
build windows amd64 win32-x64    ".exe"

echo ""
echo "Updating versions to $VERSION..."
for pkg in $TOOL $TOOL-linux-x64 $TOOL-darwin-arm64 $TOOL-darwin-x64 $TOOL-win32-x64; do
  node -e "
    const fs = require('fs');
    const p = '$NPM_DIR/$pkg/package.json';
    const j = JSON.parse(fs.readFileSync(p, 'utf8'));
    j.version = '$VERSION';
    if (j.optionalDependencies) {
      Object.keys(j.optionalDependencies).forEach(k => j.optionalDependencies[k] = '$VERSION');
    }
    fs.writeFileSync(p, JSON.stringify(j, null, 2) + '\n');
  "
done

echo ""
echo "Done. To publish:"
for pkg in $TOOL-linux-x64 $TOOL-darwin-arm64 $TOOL-darwin-x64 $TOOL-win32-x64 $TOOL; do
  echo "  cd npm/@open_harness/$pkg && npm publish --access public"
done
