#!/usr/bin/env bash
# Manual benchmark: dupelens vs jscpd sobre el mismo input.
# Pre-requisito: npm + jscpd instalados (npm i -g jscpd).
#
# Uso:
#   bash scripts/bench-vs-jscpd.sh <directorio>
#
# Notas:
#   - dupelens corre con el config root del repo (o defaults si no hay).
#   - jscpd corre con thresholds equivalentes (--min-tokens 50 --min-lines 5)
#     para comparar peras con peras. Ajustar si tu config difiere.
#   - "time" en bash mide wall-clock + user/sys; tomar el "real" para comparar.

set -e
ROOT="${1:-tools/dupelens/testdata/e2e_fixture}"

if ! command -v jscpd >/dev/null; then
  echo "jscpd no encontrado en PATH. Instalar con: npm i -g jscpd"
  exit 1
fi

echo "==> dupelens sobre $ROOT"
DUPELENS_BIN="${DUPELENS_BIN:-tools/dupelens/dupelens}"
if [ ! -x "$DUPELENS_BIN" ]; then
  (cd tools/dupelens && go build -o dupelens .)
  DUPELENS_BIN="tools/dupelens/dupelens"
fi
time "$DUPELENS_BIN" check --dir "$ROOT" --no-color > /dev/null

echo ""
echo "==> jscpd sobre $ROOT"
time jscpd "$ROOT" --min-tokens 50 --min-lines 5 --silent > /dev/null

echo ""
echo "Tip: dupelens debería estar dentro de 2x del tiempo de jscpd."
echo "Si está más lento, revisar fingerprint cache hit ratio o tokenize overhead."
