#!/usr/bin/env bash
# Build platform-specific Python wheels for an open-harness tool.
#
# Usage:
#   bash scripts/build-pypi.sh <tool>     # build all 4 platform wheels
#   bash scripts/build-pypi.sh open-harness  # build the meta (pure Python) wheel
#
# Output: pypi/<pkg_dir>/dist/*.whl
#
# Required tools on PATH:  go, python3, python -m build (pip install build)

set -euo pipefail

TOOL="${1:?usage: $0 <tool>}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# ── meta package: pure Python, no binary ─────────────────────────────────────
if [ "$TOOL" = "open-harness" ] || [ "$TOOL" = "open_harness" ]; then
  cd "$ROOT/pypi/open_harness"
  echo "Building open-harness meta (pure-Python wheel + sdist)..."
  rm -rf dist build *.egg-info
  pyproject-build --no-isolation
  echo ""
  echo "Done. Upload with: twine upload pypi/open_harness/dist/*"
  exit 0
fi

# ── per-tool packages: native wheels per platform ────────────────────────────
PKG="open_harness_${TOOL}"
PKG_DIR="$ROOT/pypi/${PKG}"

if [ ! -d "$PKG_DIR" ]; then
  echo "ERROR: directory $PKG_DIR does not exist" >&2
  exit 1
fi

VERSION=$(grep '^const version' "$ROOT/tools/${TOOL}/main.go" | grep -oE '"[^"]+"' | tr -d '"')
echo "Building ${PKG} v${VERSION} wheels for all platforms..."

# (goos, goarch, ext, python-plat-tag)
BUILDS=(
  "linux   amd64 ''    manylinux2014_x86_64"
  "darwin  arm64 ''    macosx_11_0_arm64"
  "darwin  amd64 ''    macosx_10_9_x86_64"
  "windows amd64 .exe  win_amd64"
)

cd "$PKG_DIR"
rm -rf dist build *.egg-info

BIN_DIR="src/${PKG}/bin"

for entry in "${BUILDS[@]}"; do
  # parse fields
  eval "set -- ${entry}"
  goos=$1
  goarch=$2
  ext=$3
  plat=$4
  # normalize ext (the empty quoted '' becomes literal '' from eval)
  [ "$ext" = "''" ] && ext=""

  echo ""
  echo "  → ${plat}"

  # 1. cross-compile go binary and stage it
  rm -f "$BIN_DIR"/* 2>/dev/null || true
  GOOS=$goos GOARCH=$goarch go build -ldflags="-s -w" \
    -o "$BIN_DIR/${TOOL}${ext}" "$ROOT/tools/${TOOL}"

  # 2. build the wheel with the forced platform tag.
  # --no-isolation reuses the current env (pipx-injected setuptools+wheel);
  # -C--build-option=--plat-name=<tag> forwards to bdist_wheel via PEP 517.
  pyproject-build --wheel --no-isolation \
    -C--build-option=--plat-name="$plat" >/dev/null
done

# Clean staging binary (last platform); the wheels already contain copies
rm -f "$BIN_DIR"/* 2>/dev/null || true
touch "$BIN_DIR/.gitkeep"

echo ""
echo "Wheels generated:"
ls -1 dist/*.whl
echo ""
echo "Upload with: twine upload pypi/${PKG}/dist/*"
