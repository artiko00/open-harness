#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PACK_DIR=$(mktemp -d)
TEST_DIR=$(mktemp -d)
trap "rm -rf $PACK_DIR $TEST_DIR" EXIT

echo "=== smoke-test-npm: packing y instalando los 4 wrappers ==="
echo ""

# Pack all platform + wrapper packages into tarballs
for PKG in \
  linelens-darwin-arm64 linelens \
  dupelens-darwin-arm64 dupelens \
  secretlens-darwin-arm64 secretlens \
  testlens-darwin-arm64 testlens; do
  (cd "$ROOT/npm/@open-harness/$PKG" && npm pack --pack-destination "$PACK_DIR" --quiet)
done

echo "Packed tarballs:"
ls "$PACK_DIR"/*.tgz | xargs -I{} basename {}
echo ""

# Install from tarballs into a clean project
cd "$TEST_DIR"
echo '{"name":"smoke-test","version":"1.0.0","private":true}' > package.json

npm install --save-dev \
  "$PACK_DIR"/open-harness-linelens-darwin-arm64-*.tgz \
  "$PACK_DIR"/open-harness-linelens-*.tgz \
  "$PACK_DIR"/open-harness-dupelens-darwin-arm64-*.tgz \
  "$PACK_DIR"/open-harness-dupelens-*.tgz \
  "$PACK_DIR"/open-harness-secretlens-darwin-arm64-*.tgz \
  "$PACK_DIR"/open-harness-secretlens-*.tgz \
  "$PACK_DIR"/open-harness-testlens-darwin-arm64-*.tgz \
  "$PACK_DIR"/open-harness-testlens-*.tgz \
  2>/dev/null

echo ""
echo "--- versiones ---"
node_modules/.bin/linelens version
node_modules/.bin/dupelens version
node_modules/.bin/secretlens version
node_modules/.bin/testlens version

echo ""
echo "--- linelens check en tools/linelens ---"
node_modules/.bin/linelens check --dir "$ROOT/tools/linelens" --no-color || true

echo ""
echo "=== smoke-test-npm: PASSED ==="
